アプリ概要
英語学習を支援するアプリケーションです。英語単語を登録し、英語と日本語の変換問題を出題し、正誤判定を行います。

実装機能
-　英単語(スペル・読み・部首)登録
- 英→日、日→英の出題機能
- 解答結果と正誤判定の記録

使用技術
- Go
- HTML/CSS
- MariaDB


動作環境
- OS Mac OS
- Go go1.24.1
- ブラウザ safari
- MariaDB: 10.5

使用方法
1. go run main.goで起動
2. ブラウザで127.0.0.1:8080にアクセス

ディレクトリ構成
main.go エントリーポイント
page_data/　ページごとの処理
page_data/html 各ページのHTML

注意点
STARTボタンを押下してもページ遷移が行われない場合は再度ボタンを押下してください
事前にMariaDBにてEnglishというスキーマと以下のテーブルの作成が必要です
 CREATE TABLE english_word (
   word_id INT PRIMARY KEY,
   word VARCHAR(100) Not Null,
   word_read VARCHAR(100) Not Null,
   part_of_speech _VARCHAR(10) Not Null,
   file VARCHAR(255) Not Null,
   Other VARCHAR(255) Not Null );
  
CREATE TABLE question_result (
    id INT AUTO_INCREMENT PRIMARY KEY,
    word CHAR(100) NOT NULL,
    question CHAR(100) NOT NULL,
    answer CHAR(100) NOT NULL,
    result CHAR(10) NOT NULL,
    time DATETIME NOT NULL
);
