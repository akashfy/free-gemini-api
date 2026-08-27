package db

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

var (
	DB   *sql.DB
	once sync.Once
)

const DBFileName = "data/gemini.db"

// InitDB initializes SQLite database connection and auto-migrates all tables
func InitDB() (*sql.DB, error) {
	var initErr error
	once.Do(func() {
		dir := filepath.Dir(DBFileName)
		if err := os.MkdirAll(dir, 0755); err != nil {
			initErr = err
			return
		}

		database, err := sql.Open("sqlite", DBFileName+"?_pragma=journal_mode(wal)&_pragma=synchronous(normal)&_pragma=busy_timeout(5000)")
		if err != nil {
			initErr = err
			return
		}

		// Set connection pool
		database.SetMaxOpenConns(25)
		database.SetMaxIdleConns(10)
		database.SetConnMaxLifetime(time.Hour)

		// Create tables
		schema := `
		CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			conversation_id TEXT,
			user_id TEXT,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			model TEXT,
			tokens INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_messages_conv ON messages(conversation_id);
		CREATE INDEX IF NOT EXISTS idx_messages_user ON messages(user_id);

		CREATE TABLE IF NOT EXISTS media_generations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			type TEXT NOT NULL, -- 'image', 'video', 'music'
			prompt TEXT NOT NULL,
			file_name TEXT NOT NULL,
			file_path TEXT NOT NULL,
			url TEXT,
			aspect_ratio TEXT,
			response_id TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_media_type ON media_generations(type);

		CREATE TABLE IF NOT EXISTS request_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			endpoint TEXT NOT NULL,
			method TEXT NOT NULL,
			user_ip TEXT,
			status_code INTEGER,
			elapsed_ms REAL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS accounts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			account_id TEXT UNIQUE NOT NULL,
			cookie_file TEXT NOT NULL,
			status TEXT DEFAULT 'active',
			total_requests INTEGER DEFAULT 0,
			last_used_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		`

		if _, err := database.Exec(schema); err != nil {
			initErr = err
			return
		}

		DB = database
		log.Println("🗄️  SQLite Database initialized successfully at data/gemini.db (WAL Mode)")
	})

	return DB, initErr
}

// LogMessage saves a chat turn into SQLite
func LogMessage(conversationID, userID, role, content, model string, tokens int) error {
	if DB == nil {
		return nil
	}
	query := `INSERT INTO messages (conversation_id, user_id, role, content, model, tokens) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := DB.Exec(query, conversationID, userID, role, content, model, tokens)
	return err
}

// LogMediaGeneration records a generated image, video or music track
func LogMediaGeneration(mediaType, prompt, fileName, filePath, url, aspectRatio, responseID string) error {
	if DB == nil {
		return nil
	}
	query := `INSERT INTO media_generations (type, prompt, file_name, file_path, url, aspect_ratio, response_id) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := DB.Exec(query, mediaType, prompt, fileName, filePath, url, aspectRatio, responseID)
	return err
}

// LogRequest records an API request log
func LogRequest(endpoint, method, userIP string, statusCode int, elapsedMs float64) error {
	if DB == nil {
		return nil
	}
	query := `INSERT INTO request_logs (endpoint, method, user_ip, status_code, elapsed_ms) VALUES (?, ?, ?, ?, ?)`
	_, err := DB.Exec(query, endpoint, method, userIP, statusCode, elapsedMs)
	return err
}

// GetRecentMessages retrieves conversation history
func GetRecentMessages(conversationID string, limit int) ([]map[string]any, error) {
	if DB == nil {
		return nil, nil
	}
	rows, err := DB.Query(`SELECT role, content, model, created_at FROM messages WHERE conversation_id = ? ORDER BY id DESC LIMIT ?`, conversationID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]any
	for rows.Next() {
		var role, content, model, createdAt string
		if err := rows.Scan(&role, &content, &model, &createdAt); err == nil {
			result = append(result, map[string]any{
				"role":       role,
				"content":    content,
				"model":      model,
				"created_at": createdAt,
			})
		}
	}
	return result, nil
}

// SearchMessages searches past chat history across all messages
func SearchMessages(query string, limit int) ([]map[string]any, error) {
	if DB == nil {
		return nil, nil
	}
	if limit <= 0 {
		limit = 10
	}
	searchPattern := "%" + query + "%"
	rows, err := DB.Query(`SELECT id, COALESCE(role, ''), COALESCE(content, ''), COALESCE(model, ''), COALESCE(created_at, '') FROM messages WHERE content LIKE ? ORDER BY id DESC LIMIT ?`, searchPattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]any
	for rows.Next() {
		var id int
		var role, content, model, createdAt string
		if err := rows.Scan(&id, &role, &content, &model, &createdAt); err == nil {
			result = append(result, map[string]any{
				"id":         id,
				"role":       role,
				"content":    content,
				"model":      model,
				"created_at": createdAt,
			})
		}
	}
	return result, nil
}

// GetMediaByPrompt retrieves generated media items matching a filter/query
func GetMediaByPrompt(mediaType, query string, limit int) ([]map[string]any, error) {
	if DB == nil {
		return nil, nil
	}
	if limit <= 0 {
		limit = 10
	}
	sqlQuery := `SELECT id, type, prompt, file_name, url, created_at FROM media_generations WHERE 1=1`
	var args []any
	if mediaType != "" && mediaType != "all" {
		sqlQuery += ` AND type = ?`
		args = append(args, mediaType)
	}
	if query != "" {
		sqlQuery += ` AND prompt LIKE ?`
		args = append(args, "%"+query+"%")
	}
	sqlQuery += ` ORDER BY id DESC LIMIT ?`
	args = append(args, limit)

	rows, err := DB.Query(sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]any
	for rows.Next() {
		var id int
		var mType, prompt, fileName, urlStr, createdAt string
		if err := rows.Scan(&id, &mType, &prompt, &fileName, &urlStr, &createdAt); err == nil {
			result = append(result, map[string]any{
				"id":         id,
				"type":       mType,
				"prompt":     prompt,
				"file_name":  fileName,
				"url":        urlStr,
				"created_at": createdAt,
			})
		}
	}
	return result, nil
}

// GetSystemStats returns aggregate statistics from SQLite
func GetSystemStats() (map[string]any, error) {
	if DB == nil {
		return map[string]any{"status": "database_not_connected"}, nil
	}
	var totalMessages, totalMedia, totalRequests int
	_ = DB.QueryRow(`SELECT COUNT(*) FROM messages`).Scan(&totalMessages)
	_ = DB.QueryRow(`SELECT COUNT(*) FROM media_generations`).Scan(&totalMedia)
	_ = DB.QueryRow(`SELECT COUNT(*) FROM request_logs`).Scan(&totalRequests)

	return map[string]any{
		"total_messages":   totalMessages,
		"total_media":      totalMedia,
		"total_requests":   totalRequests,
		"engine":           "SQLite WAL Mode",
		"database_file":    DBFileName,
	}, nil
}

