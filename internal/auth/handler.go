package auth

import (
	"net/http"
	"strings"
	"encoding/json"
	"image-processing/internal/database"
	"image-processing/internal/models"
	"image-processing/internal/utils"
)

func RegisterUser(w http.ResponseWriter, r *http.Request){
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
        return
	}

	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
        return
	}

	query := "INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id"
	err = database.DB.QueryRow(query, user.Username, hashedPassword).Scan(&user.ID)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
            http.Error(w, "Email already exists", http.StatusConflict)
        } else {
            http.Error(w, "Database error", http.StatusInternalServerError)
        }
        return
	}

	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
        return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
        Token string `json:"token"`
    }{Token: token})
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	var creds models.User
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
        return
	}

	var user models.User
	query := "SELECT id, username, password_hash FROM users WHERE username=$1"
	err = database.DB.QueryRow(query, creds.Username).Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
        http.Error(w, "Username not found", http.StatusUnauthorized)
        return
    }

	if !utils.CheckPassword(user.Password, creds.Password){
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	token, err := utils.GenerateToken(user.ID)
    if err != nil {
        http.Error(w, "Error generating token", http.StatusInternalServerError)
        return
    }

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(struct {
        Token string `json:"token"`
    }{Token: token})
}