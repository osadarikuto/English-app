package page

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func Get_Answer(req *Request) string {
	parts := strings.Split(req.Uri, "?")
	ans := strings.Split(parts[1], "=")
	return ans[1]
}

func Correct_judgement(answer, pro string) string {
	for _, problem := range strings.Split(pro, ",") {
		if answer == problem {
			return "正解"
		}
	}
	return "不正解"
}

func Answer_page(req *Request) string {
	answer := Get_Answer(req)
	//セッションの取得
	sid, session, _ := getSession(req)

	//値の取り出し
	quesition, _ := session["answer"]
	word, _ := session["word"]

	//正誤判定
	resu := Correct_judgement(answer, quesition)

	//結果を記録するコードを記述
	db := Open_question_db()
	defer db.Close()
	format, err := db.Prepare("INSERT INTO question_result (word, question, answer, result, time) VALUES (?, ?, ?, ?, Now())")
	defer format.Close()
	if err != nil {
		fmt.Println(err)
	}
	format.Exec(word, quesition, answer, resu)

	fileName := "page_data/html/result.html" //htmlを取り込む
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}
	result := string(bytes)

	judgement := resu

	result = strings.Replace(result, "{{.Question}}", word, 1)
	result = strings.Replace(result, "{{.Correct}}", quesition, 1)
	result = strings.Replace(result, "{{.Answer}}", answer, 1)
	result = strings.Replace(result, "{{.Result}}", judgement, 1)

	headers := fmt.Sprintf(
		"HTTP/1.1 200 OK\r\n"+
			"Content-Type: text/html; charset=utf-8\r\n"+
			"Set-Cookie: SID=%s; Path=/; HttpOnly\r\n"+
			"Content-Length: %d\r\n\r\n",
		sid, len(result),
	)
	return headers + result
}
