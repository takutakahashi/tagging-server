package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// DB schema
const createTableSQL = `
CREATE TABLE IF NOT EXISTS targets_tags (
    encryption_key_hash TEXT NOT NULL,
    target TEXT NOT NULL,
    tag TEXT NOT NULL,
    PRIMARY KEY(encryption_key_hash, target, tag)
);

CREATE TABLE IF NOT EXISTS likes (
    encryption_key_hash TEXT NOT NULL,
    target TEXT NOT NULL,
    like_count INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY(encryption_key_hash, target)
);`

// Response structure for the GET requests
type Response struct {
	Tags      []string `json:"tags,omitempty"`
	Targets   []string `json:"targets,omitempty"`
	LikeCount int      `json:"like_count,omitempty"`
}

func main() {
	// Open SQLite DB
	db, err := sql.Open("sqlite3", os.Getenv("DB_PATH"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create table if not exists
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/get-tags", func(w http.ResponseWriter, r *http.Request) {
		handleGetTags(w, r, db)
	})

	http.HandleFunc("/get-targets", func(w http.ResponseWriter, r *http.Request) {
		handleGetTargets(w, r, db)
	})

	http.HandleFunc("/add-tag", func(w http.ResponseWriter, r *http.Request) {
		handleAddTag(w, r, db)
	})

	http.HandleFunc("/like-target", func(w http.ResponseWriter, r *http.Request) {
		handleLikeTarget(w, r, db)
	})

	http.HandleFunc("/get-likes", func(w http.ResponseWriter, r *http.Request) {
		handleGetLikes(w, r, db)
	})

	// Start the server
	log.Println("Server started")
	port := os.Getenv("PORT")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func hashAuthorizationKey(authKey string) string {
	hash := sha256.Sum256([]byte(authKey))
	return hex.EncodeToString(hash[:])
}

func handleGetTags(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	authKey := r.Header.Get("Authorization")
	if authKey == "" {
		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}

	keyHash := hashAuthorizationKey(authKey)
	target := r.URL.Query().Get("target")
	if target == "" {
		http.Error(w, "Missing target", http.StatusBadRequest)
		return
	}

	rows, err := db.Query("SELECT tag FROM targets_tags WHERE encryption_key_hash = ? AND target = ?", keyHash, target)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			http.Error(w, "Failed to read from database", http.StatusInternalServerError)
			return
		}
		tags = append(tags, tag)
	}

	w.Header().Set("Content-Type", "application/json")
	response := Response{Tags: tags}
	json.NewEncoder(w).Encode(response)
}

func handleGetTargets(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	authKey := r.Header.Get("Authorization")
	if authKey == "" {
		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}

	keyHash := hashAuthorizationKey(authKey)
	tag := r.URL.Query().Get("tag")
	if tag == "" {
		http.Error(w, "Missing tag", http.StatusBadRequest)
		return
	}

	rows, err := db.Query("SELECT target FROM targets_tags WHERE encryption_key_hash = ? AND tag = ?", keyHash, tag)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var targets []string
	for rows.Next() {
		var target string
		if err := rows.Scan(&target); err != nil {
			http.Error(w, "Failed to read from database", http.StatusInternalServerError)
			return
		}
		targets = append(targets, target)
	}

	w.Header().Set("Content-Type", "application/json")
	response := Response{Targets: targets}
	json.NewEncoder(w).Encode(response)
}

func handleAddTag(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	authKey := r.Header.Get("Authorization")
	if authKey == "" {
		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}

	keyHash := hashAuthorizationKey(authKey)

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Target string `json:"target"`
		Tag    string `json:"tag"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	_, err := db.Exec("INSERT INTO targets_tags (encryption_key_hash, target, tag) VALUES (?, ?, ?)", keyHash, req.Target, req.Tag)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handleLikeTarget(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	authKey := r.Header.Get("Authorization")
	if authKey == "" {
		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}

	keyHash := hashAuthorizationKey(authKey)

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Target string `json:"target"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	_, err := db.Exec("INSERT INTO likes (encryption_key_hash, target, like_count) VALUES (?, ?, 1) ON CONFLICT(encryption_key_hash, target) DO UPDATE SET like_count = like_count + 1", keyHash, req.Target)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handleGetLikes(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	authKey := r.Header.Get("Authorization")
	if authKey == "" {
		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}

	keyHash := hashAuthorizationKey(authKey)
	target := r.URL.Query().Get("target")
	if target == "" {
		http.Error(w, "Missing target", http.StatusBadRequest)
		return
	}

	var likeCount int
	err := db.QueryRow("SELECT like_count FROM likes WHERE encryption_key_hash = ? AND target = ?", keyHash, target).Scan(&likeCount)
	if err != nil {
		if err == sql.ErrNoRows {
			likeCount = 0
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	response := Response{LikeCount: likeCount}
	json.NewEncoder(w).Encode(response)
}
