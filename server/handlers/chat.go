package handlers

import (
	"backend/server/models"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

type ChatHandler struct {
	db *sql.DB
}

func NewChatHandler(db *sql.DB) *ChatHandler {
	return &ChatHandler{db: db}
}

func (h *ChatHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`
        SELECT id, user_email, username, message_text, created_at 
        FROM global_messages 
        ORDER BY created_at ASC
    `)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		err := rows.Scan(&msg.ID, &msg.UserEmail, &msg.Username, &msg.Text, &msg.CreatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		messages = append(messages, msg)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func (h *ChatHandler) PostMessage(w http.ResponseWriter, r *http.Request) {
	var msg models.Message
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.db.QueryRow(`
        INSERT INTO global_messages (user_email, username, message_text, created_at)
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at
    `, msg.UserEmail, msg.Username, msg.Text, time.Now()).Scan(&msg.ID, &msg.CreatedAt)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(msg)
}
