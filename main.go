package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
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

	http.HandleFunc("/new-key", func(w http.ResponseWriter, r *http.Request) {
		key := newKey()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(base64.URLEncoding.EncodeToString(key))
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
		decryptedTag, err := decrypt(tag, []byte(authKey))
		if err != nil {
			http.Error(w, "Failed to decrypt tag", http.StatusInternalServerError)
			return
		}
		tags = append(tags, decryptedTag)
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
		decryptedTarget, err := decrypt(target, []byte(authKey))
		if err != nil {
			http.Error(w, "Failed to decrypt target", http.StatusInternalServerError)
			return
		}
		targets = append(targets, decryptedTarget)
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

	target, err := encrypt(req.Target, []byte(authKey))
	if err != nil {
		http.Error(w, "Failed to encrypt target", http.StatusInternalServerError)
		return
	}
	tag, err := encrypt(req.Tag, []byte(authKey))
	if err != nil {
		http.Error(w, "Failed to encrypt tag", http.StatusInternalServerError)
		return
	}
	_, err = db.Exec("INSERT INTO targets_tags (encryption_key_hash, target, tag) VALUES (?, ?, ?)", keyHash, target, tag)
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

	target, err := encrypt(req.Target, []byte(authKey))
	if err != nil {
		http.Error(w, "Failed to encrypt target", http.StatusInternalServerError)
		return
	}
	_, err = db.Exec("INSERT INTO likes (encryption_key_hash, target, like_count) VALUES (?, ?, 1) ON CONFLICT(encryption_key_hash, target) DO UPDATE SET like_count = like_count + 1", keyHash, target)
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
	decryptedTarget, err := decrypt(target, []byte(authKey))
	if err != nil {
		http.Error(w, "Failed to decrypt target", http.StatusInternalServerError)
		return
	}
	var likeCount int
	err = db.QueryRow("SELECT like_count FROM likes WHERE encryption_key_hash = ? AND target = ?", keyHash, decryptedTarget).Scan(&likeCount)
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

func newKey() []byte {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		log.Fatal(err)
	}
	return key
}

func encrypt(plainText string, key []byte) (string, error) {
	// AESブロック暗号を作成
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// 初期化ベクトルを生成
	cipherText := make([]byte, aes.BlockSize+len(plainText))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	// 暗号化処理
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], []byte(plainText))

	// base64エンコードして返す
	return base64.URLEncoding.EncodeToString(cipherText), nil
}

// 復号化
func decrypt(cipherText string, key []byte) (string, error) {
	// base64デコード
	cipherTextBytes, err := base64.URLEncoding.DecodeString(cipherText)
	if err != nil {
		return "", err
	}

	// AESブロック暗号を作成
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// 初期化ベクトルを取り出す
	if len(cipherTextBytes) < aes.BlockSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	iv := cipherTextBytes[:aes.BlockSize]
	cipherTextBytes = cipherTextBytes[aes.BlockSize:]

	// 復号化処理
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherTextBytes, cipherTextBytes)

	// 復号化した文字列を返す
	return string(cipherTextBytes), nil
}
