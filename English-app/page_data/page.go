package page

import (
	"fmt"
	"io/ioutil"

	_ "github.com/go-sql-driver/mysql"
)

func Start_page() string {
	//htmlを取り込む
	fileName := "page_data/html/start.html"
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
	//応答を作成して返答
	return headers + result
}
