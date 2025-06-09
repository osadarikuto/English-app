package page

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func Registration_page(req *Request) string {
	fileName := "page_data/html/registration.html" //htmlを取り込む
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}

	result := string(bytes)
	headers := fmt.Sprintf(
		"HTTP/1.1 200 OK\r\n"+
			"Content-Type: text/html; charset=utf-8\r\n"+
			"Path=/; HttpOnly\r\n"+
			"Content-Length: %d\r\n\r\n", len(result),
	)
	return headers + result
}

func disassembly(form string) []string {
	var values []string
	uri := strings.Split(form, "&")
	for i := 0; i < 4; i++ {
		value := strings.Split(uri[i], "=")
		values = append(values, value[1])
	}
	return values
}

func Completion_page(req *Request) string {
	//DBに接続
	db := Open_question_db()
	//クエリの実行
	con, err := db.Prepare("INSERT INTO english_word( word_id, word, word_read, part_of_speech, file, Other) VALUES(?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer con.Close()

	var count int
	//設定するIDの取得
	err = db.QueryRow("SELECT COUNT(*) FROM english_word").Scan(&count)
	if err != nil {
		log.Fatal(err)
	}
	//値の受け取り
	values := disassembly(req.Uri)
	fmt.Println(values)
	word, read, path, other := values[0], values[1], values[2], values[3]

	//DBに登録
	_, err = con.Exec(count+1, word, read, path, word, other)
	if err != nil {
		log.Fatal(err)
	}
	fileName := "page_data/html/completion.html" //htmlを取り込む
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}

	result := string(bytes)
	headers := fmt.Sprintf(
		"HTTP/1.1 200 OK\r\n"+
			"Content-Type: text/html; charset=utf-8\r\n"+
			"Path=/; HttpOnly\r\n"+
			"Content-Length: %d\r\n\r\n", len(result),
	)
	return headers + result
}
