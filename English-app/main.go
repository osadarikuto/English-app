package main

import (
	page "english-app/page_data"

	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/textproto"
	"os"
	"strconv"
	"strings"
	"syscall"
)

type netSocket struct {
	fd int
}

func (ns netSocket) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	n, err := syscall.Read(ns.fd, p)
	if err != nil {
		n = 0
	}
	return n, err
}

func (ns *netSocket) Accept() (*netSocket, error) {
	//接続を受け入れる
	nfd, _, err := syscall.Accept(ns.fd)
	if err == nil {
		syscall.CloseOnExec(nfd)
	}
	if err != nil {
		return nil, err
	}
	return &netSocket{nfd}, err
}

func (ns *netSocket) Close() error {
	return syscall.Close(ns.fd)
}

func (ns netSocket) Write(p []byte) (int, error) {
	n, err := syscall.Write(ns.fd, p)
	if err != nil {
		n = 0
	}
	return n, err
}

func Crate_NatSocket(ip net.IP, port int) (*netSocket, error) {
	//フォークが発生しないようにロック
	syscall.ForkLock.Lock()
	//ソケット作成
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_IP)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	syscall.ForkLock.Unlock()

	//IP・PORT振り分け
	addr := syscall.SockaddrInet4{Port: 8080}
	copy(addr.Addr[:], net.ParseIP("127.0.0.1").To4())

	err = syscall.Bind(fd, &addr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	//通信の受付開始
	err = syscall.Listen(fd, syscall.SOMAXCONN)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return &netSocket{fd}, nil
}

// HTTPリクエストを解析
func parseRequest(c *netSocket) (*page.Request, error) {
	//リクエスト読み込み
	b := bufio.NewReader(*c)
	//リクエストの要素ごとに読み取る
	tp := textproto.NewReader(b)
	req := new(page.Request)

	var s string

	s, err := tp.ReadLine()
	if err != nil {
		return nil, fmt.Errorf("failed to read request line: %v", err)
	}

	sp := strings.Split(s, " ")
	if len(sp) < 3 {
		return nil, fmt.Errorf("invalid HTTP request line: %s", s)
	}

	//リクエスト上部の分割した要素を各変数に代入
	req.Method, req.Uri, req.Proto = sp[0], sp[1], sp[2]

	mimeHeader, _ := tp.ReadMIMEHeader()
	req.Header = mimeHeader

	if req.Method == "GET" || req.Method == "HEAD" {
		return req, nil
	}
	if len(req.Header["Content-Length"]) == 0 {
		return nil, errors.New("no content length")
	}
	//ボディ部分のサイズの計測
	length, err := strconv.Atoi(req.Header["Content-Length"][0])
	if err != nil {
		return nil, err
	}
	body := make([]byte, length)
	//ボディの部分を読み込む
	if _, err = io.ReadFull(b, body); err != nil {
		return nil, err
	}
	req.Body = body
	return req, nil
}

func Get_form(uri string) (string, string) {
	//URLとパラメータを分割
	parts := strings.SplitN(uri, "?", 2)
	path := parts[0]
	form := ""
	if len(parts) == 2 {
		//パラメータの値のみを抽出
		form = parts[1]
	}
	return path, form
}

func router(req *page.Request) string {

	path, form := Get_form(req.Uri)
	log.Printf("Requested path: %s", path)

	//リクエストされたURIのルーティング
	switch path {
	case "/top":
		return page.Start_page()
	case "/question":
		if form != "" {
			return page.Question_page(req)
		}

		//値が帰ってこなかった場合のエラー画面表示
		er := "<h1>400 Bad Request: Missing query parameters</h1>"
		headers := fmt.Sprintf(
			"HTTP/1.1 200 OK\r\n"+
				"Content-Type: text/html; charset=utf-8\r\n"+
				"Path=/; HttpOnly\r\n"+
				"Content-Length: %d\r\n\r\n", len(er),
		)

		return headers

	case "/answer":
		return page.Answer_page(req)

	case "/registration":
		return page.Registration_page(req)

	case "/completion":
		return page.Completion_page(req)

	default:
		// ルーティングにアタハマらない場合のエラー画面作成
		er := "<h1>404 Not Found</h1>"
		headers := fmt.Sprintf(
			"HTTP/1.1 200 OK\r\n"+
				"Content-Type: text/html; charset=utf-8\r\n"+
				"Path=/; HttpOnly\r\n"+
				"Content-Length: %d\r\n\r\n", len(er),
		)
		return headers
	}
}

func main() {
	//アドレスとポートの受け取り
	ipFlag := flag.String("ip_addr", "127.0.0.1", "The IP address to use")
	portFlag := flag.Int("port", 8080, "The port to use.")
	flag.Parse()

	//受け取った値を変数に代入
	ip := net.ParseIP(*ipFlag)
	port := *portFlag
	//ソケットの作成
	socket, err := Crate_NatSocket(ip, port)
	defer socket.Close()
	if err != nil {
		panic(err)
	}

	log.Print("Server Started!")
	log.Printf("addr: http://%s:%d", ip, port)

	for {
		var res string
		//ソケットの設定
		rw, err := socket.Accept()
		log.Printf("Incoming connection")
		if err != nil {
			panic(err)
		}

		//リクエストの解析
		req, _ := parseRequest(rw)

		//URLのルーティング
		res = router(req)

		log.Print("Writing response")
		//レスポンスの送信
		_, err = io.WriteString(rw, res)
		if err != nil {
			log.Printf("Failed to write response: %v", err)
		}
		defer rw.Close()
	}
}
