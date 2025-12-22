package main

import (
	"fmt"
	"html/template"
	"net/http"
)

// PageData はHTMLテンプレートに渡すデータ構造
type PageData struct {
	ImageURL string
}

// 関数1に対応するハンドラー
func startHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("start called")
	serveImage(w, "image1.jpg")

}

// 関数2に対応するハンドラー
func rankingHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("ranking called")
	serveImage(w, "image2.jpg")
}

// 関数3に対応するハンドラー
func loginHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("login called")
	serveImage(w, "image3.jpg")
}

// 画像のURLをHTMLに埋め込んでレスポンスを返す
func serveImage(w http.ResponseWriter, imageURL string) {
	// HTMLテンプレート
	const tmpl = `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>関数と画像のペア表示</title>
	</head>
	<body>
		<h1>画像表示</h1>
		<img src="{{.ImageURL}}" alt="表示する画像" style="width:300px; height:200px;">
		<br>
		<a href="/">メインメニューに戻る</a>
	</body>
	</html>
	`

	// テンプレートをパース
	t := template.Must(template.New("imagePage").Parse(tmpl))

	// データを埋め込んでテンプレートを出力
	data := PageData{ImageURL: imageURL}
	if err := t.Execute(w, data); err != nil {
		http.Error(w, "テンプレートの実行に失敗しました", http.StatusInternalServerError)
	}
}

// メインメニューのハンドラー
func menuHandler(w http.ResponseWriter, r *http.Request) {
	const menu = `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>メインメニュー</title>
	</head>
	<body>
            <h1>関数を選んで画像を表示</h1>
			<a href="/start">スタート</a>
			<a href="/ranking">ランキング</a>
			<a href="/login">ログイン</a>
	</body>
	</html>
	`
	// メニュー画面を表示
	w.Write([]byte(menu))
}

func main() {
	// 静的ファイルの配信 (画像ファイルの提供)
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// 各ハンドラーを登録
	http.HandleFunc("/", menuHandler)
	http.HandleFunc("/start", startHandler)
	http.HandleFunc("/ranking", rankingHandler)
	http.HandleFunc("/login", loginHandler)

	// サーバーを起動
	fmt.Println("サーバーをポート8080で起動中...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("サーバーの起動に失敗しました:", err)
	}
}
