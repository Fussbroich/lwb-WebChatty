package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

// Einige Laufzeit-Objekte für die Web-app
type app struct {
	db        *sql.DB            // Datenbankhandle
	address   string             // IP des Web-Servers
	templates *template.Template //HTML-Template
	mux       *http.ServeMux
}

func main() {
	db, err := sql.Open("postgres", "user=lewein password=niewel dbname=lwb sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	app := NewWebApp(db, "192.168.188.24:8080")
	app.Starte()
}

func NewWebApp(db *sql.DB, server_addr string) *app {
	app := &app{
		db:        db,
		address:   server_addr,
		mux:       http.NewServeMux(),
		templates: template.Must(template.ParseFiles("chat.html")),
	}

	// Routen und Handler registrieren
	app.mux.Handle("GET /{$}", http.RedirectHandler("/chat/default", http.StatusMovedPermanently))
	app.mux.HandleFunc("GET /chat/{key}", app.handleChatGET)
	app.mux.HandleFunc("POST /chat/{key}", app.handleChatPOST)
	// Dateizugriff für Icons und Styles
	app.mux.Handle("GET /favicon.ico", http.FileServer(http.Dir("static")))
	app.mux.Handle("GET /", http.FileServer(http.Dir("static")))

	return app
}

func (s *app) Starte() {
	fmt.Println("Starte lwb-WebChatty unter http://localhost:8080")
	http.ListenAndServe(s.address, s.mux)
}

// Die Modell-Objekte und die Handler

type Message struct {
	MsgID  uint64
	Text   string
	TStamp time.Time
}

type AppParams struct {
	Key      string     // Chatraum-Schlüssel
	Messages []*Message // anzuzeigende Nachrichten
}

func (s *app) handleChatGET(w http.ResponseWriter, r *http.Request) {
	var key string
	var err error
	var chatroomID uint64
	if key = r.PathValue("key"); key == "" {
		http.Error(w, "Schluessel fehlt", http.StatusBadRequest)
		return
	}
	if chatroomID, err = s.ensureChatroom(key); err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}

	// Fetch and display messages
	var messages []*Message
	if messages, err = s.fetchMessages(chatroomID); err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}
	if err = s.templates.ExecuteTemplate(w, "chat.html",
		AppParams{
			Key:      key,
			Messages: messages}); err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}
}

func (s *app) handleChatPOST(w http.ResponseWriter, r *http.Request) {
	var key string
	var err error
	var chatroomID uint64
	key = r.PathValue("key") // extract 'key' from the URL
	if chatroomID, err = s.ensureChatroom(key); err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// Process posted message
	r.ParseForm()
	if text := r.FormValue("message"); text != "" {
		if err = s.insertMessage(chatroomID, text); err != nil {
			http.Error(w, "Failed to insert message", http.StatusInternalServerError)
			return
		}
	}

	// Redirect or re-render the page
	http.Redirect(w, r, fmt.Sprintf("/chat/%s", key), http.StatusFound)
}

// #### Datenbankzugriffe für die App

func (s *app) ensureChatroom(key string) (uint64, error) {
	var id uint64
	var err error
	if err = s.db.QueryRow(`
INSERT INTO chatrooms (key)
VALUES ($1)
ON CONFLICT (key)
DO UPDATE SET key=EXCLUDED.key RETURNING id`, key).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (s *app) fetchMessages(chatroomID uint64) ([]*Message, error) {
	var messages []*Message
	var rows *sql.Rows
	var err error
	if rows, err = s.db.Query(`
SELECT id, message, timestamp
FROM messages
WHERE chatroom_id = $1
ORDER BY timestamp ASC`, chatroomID); err != nil {
		return nil, err
	}
	defer rows.Close()

	messages = []*Message{}
	for rows.Next() {
		var msg = Message{}
		if err := rows.Scan(&msg.MsgID, &msg.Text, &msg.TStamp); err != nil {
			return nil, err
		}
		messages = append(messages, &msg)
	}
	return messages, nil
}

func (s *app) insertMessage(chatroomID uint64, text string) error {
	var err error
	if _, err = s.db.Exec(`
INSERT
INTO messages (chatroom_id, message)
VALUES ($1, $2)`, chatroomID, text); err != nil {
		return err
	}
	return nil
}
