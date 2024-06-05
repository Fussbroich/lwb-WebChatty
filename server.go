package main

import (
	"database/sql"
	"html/template"
	"net/http"

	_ "github.com/lib/pq"
)

type Server struct {
	db        *sql.DB
	templates *template.Template
}

func main() {
	db, err := sql.Open("postgres", "postgres://lewein:niewel@localhost/lwb?sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	server := &Server{
		db:        db,
		templates: template.Must(template.ParseFiles("chat.html")),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", server.handleHome)
	mux.HandleFunc("/chat", server.handleChat)

	http.ListenAndServe(":8080", mux)
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	s.templates.ExecuteTemplate(w, "chat.html", nil)
}

func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Key is required", http.StatusBadRequest)
		return
	}

	chatroomID, err := s.ensureChatroom(key)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	if r.Method == "POST" {
		r.ParseForm()
		message := r.FormValue("message")
		if message != "" {
			_, err := s.db.Exec("INSERT INTO messages (chatroom_id, message) VALUES ($1, $2)", chatroomID, message)
			if err != nil {
				http.Error(w, "Failed to insert message", http.StatusInternalServerError)
				return
			}
		}
	}

	rows, err := s.db.Query("SELECT message FROM messages WHERE chatroom_id = $1 ORDER BY timestamp ASC", chatroomID)
	if err != nil {
		http.Error(w, "Failed to fetch messages", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	messages := []string{}
	for rows.Next() {
		var msg string
		if err := rows.Scan(&msg); err != nil {
			http.Error(w, "Failed to read messages", http.StatusInternalServerError)
			return
		}
		messages = append(messages, msg)
	}

	s.templates.ExecuteTemplate(w, "chat.html", messages)
}

func (s *Server) ensureChatroom(key string) (int, error) {
	var id int
	err := s.db.QueryRow("INSERT INTO chatrooms (key) VALUES ($1) ON CONFLICT (key) DO UPDATE SET key=EXCLUDED.key RETURNING id", key).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}
