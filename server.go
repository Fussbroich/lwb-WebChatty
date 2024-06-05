package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"

	_ "github.com/lib/pq"
)

type Server struct {
	db        *sql.DB
	templates *template.Template
}

func main() {
	db, err := sql.Open("postgres", "postgres://username:password@localhost/dbname?sslmode=disable")
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

	// Setup route for GET and POST using method-specific patterns
	mux.HandleFunc("GET /chat/{key}", server.handleChatGET)
	mux.HandleFunc("POST /chat/{key}", server.handleChatPOST)

	http.ListenAndServe(":8080", mux)
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	s.templates.ExecuteTemplate(w, "chat.html", nil)
}

func (s *Server) handleChatGET(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key") // Adjust as necessary to extract 'key' from the URL
	chatroomID, err := s.ensureChatroom(key)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// Fetch and display messages
	messages, err := s.fetchMessages(chatroomID)
	if err != nil {
		http.Error(w, "Failed to fetch messages", http.StatusInternalServerError)
		return
	}
	s.templates.ExecuteTemplate(w, "chat.html", messages)
}

func (s *Server) handleChatPOST(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key") // Adjust as necessary to extract 'key' from the URL
	chatroomID, err := s.ensureChatroom(key)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// Process posted message
	if err := s.processPostedMessage(chatroomID, r); err != nil {
		http.Error(w, "Failed to post message", http.StatusInternalServerError)
		return
	}

	// Redirect or re-render the page
	http.Redirect(w, r, fmt.Sprintf("/chat/%s", key), http.StatusFound)
}

func (s *Server) ensureChatroom(key string) (int, error) {
	// Implementation assumes the chatroom is ensured in the DB and returns its ID
	return 0, nil
}

func (s *Server) fetchMessages(chatroomID int) ([]string, error) {
	// Fetch messages from the database
	return nil, nil
}

func (s *Server) processPostedMessage(chatroomID int, r *http.Request) error {
	// Process the incoming POST request to add a message to the chatroom
	return nil
}
