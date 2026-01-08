package main

import (
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Question struct {
	ID           int
	QuestionText string
	Answer       string
}

// 問題データ
var questions = []Question{
	{ID: 1, QuestionText: "晴耕雨讀", Answer: "せいこううどく"},
	{ID: 2, QuestionText: "加賀鳶", Answer: "かがとび"},
	{ID: 3, QuestionText: "天照", Answer: "てんしょう"},
	{ID: 4, QuestionText: "猿川", Answer: "さるこう"},
	{ID: 5, QuestionText: "CHILL GREEN", Answer: "ちるぐりーん"},
	{ID: 6, QuestionText: "久米島の久米仙", Answer: "くめじまのくめせん"},
	{ID: 7, QuestionText: "壱岐", Answer: "いき"},
	{ID: 8, QuestionText: "翠", Answer: "すい"},
	{ID: 9, QuestionText: "山椒", Answer: "さんしょう"},
	{ID: 10, QuestionText: "白雪", Answer: "しらゆき"},
	{ID: 11, QuestionText: "上喜元", Answer: "じょうきげん"},
	{ID: 12, QuestionText: "〆張鶴", Answer: "しめはりつる"},
	{ID: 13, QuestionText: "天狗舞", Answer: "てんぐまい"},
	{ID: 14, QuestionText: "鍋島", Answer: "なべしま"},
}

var css = `
/* モバイルのCSS */
:root{
  --bg: #f3f4f6;
  --card: #ffffff;
  --accent: #2563eb;
  --text: #111827;
  --muted: #6b7280;
  --success: #16a34a;
  --danger: #dc2626;
  --radius: 10px;
}

/* レイアウトの調整 */
* { box-sizing: border-box; margin: 0; padding: 0; }
html,body { height: 100%; }
body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Hiragino Kaku Gothic ProN", "Noto Sans JP", sans-serif;
  background: var(--bg);
  color: var(--text);
  line-height: 1.4;
  padding: 16px;
  -webkit-font-smoothing:antialiased;
  -moz-osx-font-smoothing:grayscale;
}

/* 四角いオブジェクト */
.container {
  width: min(100%, 960px);
  max-width: 920px;
  margin: 18px auto;
  background: var(--card);
  border-radius: var(--radius);
  padding: 18px;
  box-shadow: 0 6px 20px rgba(2,6,23,0.06);
}

/* 見出しと説明 */
h1 { font-size: clamp(1.05rem, 4vw, 1.5rem); margin-bottom: 12px; }
p.question { margin-bottom: 14px; font-size: clamp(1rem, 3.5vw, 1.1rem); }
small { color: var(--muted); display:block; margin-top:8px; }

/* ホーム */
form { display:block; margin-top:8px; }
.input-row { display:flex; gap:8px; align-items:center; flex-wrap:wrap; }
input[type="text"] {
  flex: 1 1 220px;
  min-width: 0;
  padding: 12px 14px;
  border: 1px solid #e6e9ee;
  border-radius: 8px;
  font-size: clamp(0.95rem, 2.8vw, 1rem);
}
button {
  appearance: none;
  -webkit-appearance: none;
  background: var(--accent);
  color: #fff;
  border: none;
  padding: 10px 14px;
  border-radius: 8px;
  cursor: pointer;
  font-size: clamp(0.95rem, 2.8vw, 1rem);
}

/* 整理整頓よう */
.footer {
  margin-top: 14px;
  display:flex;
  gap:12px;
  align-items:center;
  flex-wrap:wrap;
  color: var(--muted);
  font-size: 0.95rem;
}

/* */
.result-correct { color: var(--success); font-weight:700; }
.result-wrong { color: var(--danger); font-weight:700; }

/* 結果の場面の設定 */
@media (min-width: 720px) {
  .container { padding: 28px; }
  .input-row { gap: 12px; }
  button { padding: 12px 16px; }
}

/* 文字を押せるようにする */
a.action {
  color: var(--accent);
  text-decoration: none;
  padding: 8px 10px;
  border-radius: 8px;
  background: rgba(37,99,235,0.06);
  font-weight: 600;
}

/* 選択したやつがどこにあるかわかりやすくしたやつ */
button:focus, input:focus, a:focus {
  outline: 3px solid rgba(37,99,235,0.15);
  outline-offset: 2px;
  border-radius: 8px;
}
`

type Session struct {
	Order   []int // questions のインデックス順
	Pos     int   // 次に出す問題の位置（0..len-1）
	Correct int   // 正解数
}

var (
	sessions      = map[string]*Session{}
	sessionsMutex sync.Mutex
)

const sessionCookieName = "nn_session"

func main() {
	rand.Seed(time.Now().UnixNano())

	http.HandleFunc("/", handleIndex)        // ホーム
	http.HandleFunc("/start", handleStart)   // 問題出題ページ
	http.HandleFunc("/submit", handleSubmit) // 回答 → 結果表示（即時）
	http.HandleFunc("/done", handleDone)     // 全問終了ページ
	http.HandleFunc("/add", handleAddForm)   // 問題追加フォーム
	http.HandleFunc("/add/submit", handleAdd)
	http.HandleFunc("/style.css", handleStyle)

	addr := ":8080"
	log.Printf("listening on %s ...", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func handleStyle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	_, _ = w.Write([]byte(css))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	const tpl = `<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width,initial-scale=1">
	<title>酒の名前を当てるゲーム</title>
	<link rel="stylesheet" href="/style.css">
</head>
<body>
	<div class="container">
		<h1>酒の振り仮名ゲームへようこそ</h1>
		<p>問題数: {{.Count}}</p>
		<p><a class="action" href="/start">スタート</a></p>
		<small>入力欄はひらがなで入力お願いします</small>
	</div>
</body>
</html>
`
	data := struct{ Count int }{Count: len(questions)}
	t := template.Must(template.New("index").Parse(tpl))
	if err := t.Execute(w, data); err != nil {
		http.Error(w, "テンプレート表示エラー", http.StatusInternalServerError)
		log.Printf("index execute error: %v", err)
	}
}

func handleStart(w http.ResponseWriter, r *http.Request) {
	sess := getOrCreateSession(w, r)
	// もし全部解き終わっていたら /done へ
	if sess.Pos >= len(sess.Order) {
		http.Redirect(w, r, "/done", http.StatusSeeOther)
		return
	}

	qidx := sess.Order[sess.Pos]
	question := questions[qidx]

	const tpl = `<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width,initial-scale=1">
	<title>問題ページ</title>
	<link rel="stylesheet" href="/style.css">
</head>
<body>
	<div class="container">
		<h1>問題</h1>
		<p class="question">{{ .QuestionText }}</p>
		<form action="/submit" method="POST" autocomplete="off">
			<div class="input-row">
				<input type="text" name="answer" placeholder="答えを入力" autocomplete="off" required>
				<button type="submit">回答する</button>
			</div>
		</form>
		<p class="footer">進捗: {{.Pos}} / {{.Total}} <a href="/">ホームへ</a></p>
	</div>
</body>
</html>
`
	data := struct {
		QuestionText string
		Pos          int
		Total        int
	}{
		QuestionText: question.QuestionText,
		Pos:          sess.Pos + 1, // 表示用は1始まり
		Total:        len(sess.Order),
	}

	t := template.Must(template.New("question").Parse(tpl))
	if err := t.Execute(w, data); err != nil {
		http.Error(w, "テンプレート表示エラー", http.StatusInternalServerError)
		log.Printf("question execute error: %v", err)
	}
}

// 回答を受け取り「結果ページ」を表示
func handleSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/start", http.StatusSeeOther)
		return
	}
	sess := getSession(r)
	if sess == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	answer := strings.TrimSpace(r.FormValue("answer"))
	sessionsMutex.Lock()
	// safety: pos が終わっていたら /done へ
	if sess.Pos >= len(sess.Order) {
		sessionsMutex.Unlock()
		http.Redirect(w, r, "/done", http.StatusSeeOther)
		return
	}
	qidx := sess.Order[sess.Pos]
	correctAnswer := strings.TrimSpace(questions[qidx].Answer)
	isCorrect := answer == correctAnswer
	if isCorrect {
		sess.Correct++
	}

	currentPosDisplay := sess.Pos + 1
	total := len(sess.Order)
	sess.Pos++
	hasNext := sess.Pos < len(sess.Order)
	sessionsMutex.Unlock()

	const tpl = `<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width,initial-scale=1">
	<title>回答結果</title>
	<link rel="stylesheet" href="/style.css">
</head>
<body>
	<div class="container">
		{{ if .IsCorrect }}
			<h1 class="result-correct">正解です！</h1>
		{{ else }}
			<h1 class="result-wrong">不正解です！</h1>
			<p>正解は「{{ .CorrectAnswer }}」でした。</p>
		{{ end }}
		<p>あなたの回答: 「{{ .YourAnswer }}」</p>
		<p>進捗: {{ .Current }} / {{ .Total }}</p>
		<p class="footer">
			{{ if .HasNext }}
				<a class="action" href="/start">次へ</a>
			{{ else }}
				<a class="action" href="/done">結果を見る</a>
			{{ end }}
			<a href="/">ホームへ</a>
		</p>
	</div>
</body>
</html>
`
	data := struct {
		IsCorrect     bool
		CorrectAnswer string
		YourAnswer    string
		Current       int
		Total         int
		HasNext       bool
	}{
		IsCorrect:     isCorrect,
		CorrectAnswer: correctAnswer,
		YourAnswer:    answer,
		Current:       currentPosDisplay,
		Total:         total,
		HasNext:       hasNext,
	}

	t := template.Must(template.New("result").Parse(tpl))
	if err := t.Execute(w, data); err != nil {
		http.Error(w, "テンプレート表示エラー", http.StatusInternalServerError)
		log.Printf("result execute error: %v", err)
	}
}

// おつかれさまページ
func handleDone(w http.ResponseWriter, r *http.Request) {
	sess := getSession(r)
	if sess == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	sessionsMutex.Lock()
	total := len(sess.Order)
	correct := sess.Correct
	sessionsMutex.Unlock()

	clearSession(w, r)

	const tpl = `<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width,initial-scale=1">
	<title>おつかれさま</title>
	<link rel="stylesheet" href="/style.css">
</head>
<body>
	<div class="container">
		<h1>おつかれさま！</h1>
		<p>全問終了しました。</p>
		<p>正解数: {{.Correct}} / {{.Total}}</p>
		<p class="footer"><a class="action" href="/">ホームへ</a> <a class="action" href="/start">もう一度挑戦</a></p>
	</div>
</body>
</html>
`
	data := struct {
		Correct int
		Total   int
	}{Correct: correct, Total: total}

	t := template.Must(template.New("done").Parse(tpl))
	if err := t.Execute(w, data); err != nil {
		http.Error(w, "テンプレート表示エラー", http.StatusInternalServerError)
		log.Printf("done execute error: %v", err)
	}
}

func handleAddForm(w http.ResponseWriter, r *http.Request) {
	const tpl = `<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width,initial-scale=1">
	<title>問題を追加する</title>
	<link rel="stylesheet" href="/style.css">
</head>
<body>
	<div class="container">
		<h1>新しい問題を追加</h1>
		<form action="/add/submit" method="POST" autocomplete="off">
			<p>問題: <input type="text" name="questionText" autocomplete="off" required></p>
			<p>正解: <input type="text" name="answer" autocomplete="off" required></p>
			<button type="submit">追加する</button>
		</form>
		<p class="footer"><a href="/">戻る</a></p>
	</div>
</body>
</html>
`
	t := template.Must(template.New("addForm").Parse(tpl))
	if err := t.Execute(w, nil); err != nil {
		http.Error(w, "テンプレート表示エラー", http.StatusInternalServerError)
		log.Printf("addForm execute error: %v", err)
	}
}

func handleAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/add", http.StatusSeeOther)
		return
	}
	questionText := strings.TrimSpace(r.FormValue("questionText"))
	answer := strings.TrimSpace(r.FormValue("answer"))
	if questionText == "" || answer == "" {
		http.Error(w, "入力が不正です", http.StatusBadRequest)
		return
	}

	questions = append(questions, Question{
		ID:           nextQuestionID(),
		QuestionText: questionText,
		Answer:       answer,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func nextQuestionID() int {
	maxID := 0
	for _, q := range questions {
		if q.ID > maxID {
			maxID = q.ID
		}
	}
	return maxID + 1
}

func newSession() *Session {
	perm := rand.Perm(len(questions))
	return &Session{
		Order:   perm,
		Pos:     0,
		Correct: 0,
	}
}

func genSessionID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36) + "-" + strconv.Itoa(rand.Intn(1<<30))
}

func getOrCreateSession(w http.ResponseWriter, r *http.Request) *Session {

	sid, err := r.Cookie(sessionCookieName)
	if err == nil {
		sessionsMutex.Lock()
		if s, ok := sessions[sid.Value]; ok {
			sessionsMutex.Unlock()
			return s
		}
		sessionsMutex.Unlock()
	}

	sidVal := genSessionID()
	s := newSession()

	sessionsMutex.Lock()
	sessions[sidVal] = s
	sessionsMutex.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sidVal,
		Path:     "/",
		HttpOnly: true,
	})
	return s
}

func getSession(r *http.Request) *Session {
	c, err := r.Cookie(sessionCookieName)
	if err != nil {
		return nil
	}
	sessionsMutex.Lock()
	defer sessionsMutex.Unlock()
	return sessions[c.Value]
}

func clearSession(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie(sessionCookieName)
	if err != nil {
		return
	}

	sessionsMutex.Lock()
	delete(sessions, c.Value)
	sessionsMutex.Unlock()
	// delete cookie client-side
	http.SetCookie(w, &http.Cookie{
		Name:   sessionCookieName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
}

func init() {

	fmt.Println("main.go loaded")
}
