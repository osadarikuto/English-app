package page

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/textproto"
	"os"
	"strconv"
	"strings"
	"sync"
)

// var store = sessions.NewCookieStore([]byte("secret-key"))
var sessionStore = make(map[string]map[string]string)
var sessionMutex = sync.Mutex{}

type Request struct {
	Method string // GET, POST, etc.
	Header textproto.MIMEHeader
	Body   []byte
	Uri    string // The raw URI from the request
	Proto  string // "HTTP/1.1"
}

type Question struct {
	Id     int
	Word   string
	Answer string
	Part   string
	File   string
	Other  string
}

func Open_question_db() *sql.DB {
	//データベースに接続
	db, err := sql.Open("mysql", "root:English@tcp(192.168.10.11:3306)/English")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	//接続に問題がないか再度確認
	err = db.Ping()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return db
}

func Random_numeric(min, max int) int {
	between := max - min + 1
	result := rand.Intn(between) + min
	return result
}

func Search_word(num int) Question {
	db := Open_question_db()
	defer db.Close()
	id := strconv.Itoa(num)
	result, err := db.Query("SELECT * FROM english_word  WHERE word_id=" + id)
	if err != nil {
		log.Println(err)
	}
	defer result.Close()
	var question Question
	for result.Next() {
		err = result.Scan(&question.Id, &question.Word, &question.Answer, &question.Part, &question.File, &question.Other)
		if err != nil {
			fmt.Println(err)
		}
	}
	return question
}

func disassembly_word(form string) []string {
	var values []string
	uri := strings.Split(form, "&")
	for i := 0; i < 3; i++ {
		value := strings.Split(uri[i], "=")
		values = append(values, value[1])
	}
	return values
}

func Template(cont string, question Question) (result string) {
	result = strings.Replace(cont, "{{.Word}}", question.Word, 1)
	result = strings.Replace(result, "{{.Part}}", question.Part, 1)
	result = strings.Replace(result, "{{.Other}}", question.Other, 1)
	result = strings.Replace(result, "{{.File}}", question.File, 1)
	return result
}

func generateSessionID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func parseCookie(header textproto.MIMEHeader) string {
	//セッションIDの取得
	cookies := header["Cookie"]
	if len(cookies) == 0 {
		return ""
	}

	//共有mapのIDと一致する確認
	pairs := strings.Split(cookies[0], ";")
	for _, p := range pairs {
		kv := strings.SplitN(strings.TrimSpace(p), "=", 2)
		if len(kv) == 2 && kv[0] == "SID" {
			return kv[1]
		}
	}
	return ""
}

func getSession(req *Request) (string, map[string]string, bool) {
	sid := parseCookie(req.Header)
	sessionMutex.Lock()
	defer sessionMutex.Unlock()

	sess, exists := sessionStore[sid]
	if exists {
		return sid, sess, true
	}

	// 新規セッション作成
	newSid := generateSessionID()
	sessionStore[newSid] = map[string]string{}
	return newSid, sessionStore[newSid], false
}

func Question_page(req *Request) string {
	//URIから値を取得
	values := disassembly_word(req.Uri)
	mode, mi, ma := values[0], values[1], values[2]
	min, _ := strconv.Atoi(mi)
	max, _ := strconv.Atoi(ma)

	//セッションを作成・IDの確認
	sid, session, _ := getSession(req)

	if mode == "" {
		mode = session["mode"]
		min, _ = strconv.Atoi(session["min"])
		max, _ = strconv.Atoi(session["max"])

	}
	//乱数生成
	num := Random_numeric(min, max)
	//番号に対応した単語情報を代入
	question := Search_word(num)

	//モードに合わせて問題を変換
	if mode == "Jap" {
		question.Word, question.Answer = question.Answer, question.Word
	}

	fileName := "page_data/html/question.html" //htmlを取り込む
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}
	//セッションに値を登録
	session["mode"] = mode
	session["min"] = strconv.Itoa(min)
	session["max"] = strconv.Itoa(max)
	session["answer"] = question.Answer
	session["word"] = question.Word

	result := string(bytes)

	result = Template(result, question)
	headers := fmt.Sprintf(
		"HTTP/1.1 200 OK\r\n"+
			"Content-Type: text/html; charset=utf-8\r\n"+
			"Set-Cookie: SID=%s; Path=/; HttpOnly\r\n"+
			"Content-Length: %d\r\n\r\n",
		sid, len(result),
	)
	return headers + result
}
