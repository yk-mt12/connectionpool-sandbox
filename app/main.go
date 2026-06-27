package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	poolDB *sql.DB
	dsn    string
)

func main() {
	dsn = os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "root:root@tcp(localhost:3306)/sandbox?parseTime=true"
	}

	var err error
	poolDB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("open pool DB: %v", err)
	}
	defer poolDB.Close()

	poolDB.SetMaxOpenConns(10)
	poolDB.SetMaxIdleConns(5)
	poolDB.SetConnMaxLifetime(5 * time.Minute)

	if err := poolDB.Ping(); err != nil {
		log.Fatalf("ping MySQL: %v", err)
	}
	log.Println("connected to MySQL (with pool)")

	mux := http.NewServeMux()
	mux.HandleFunc("/with-pool", withPoolHandler)
	mux.HandleFunc("/without-pool", withoutPoolHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

type Response struct {
	Mode       string  `json:"mode"`
	DurationMs float64 `json:"duration_ms"`
	Error      string  `json:"error,omitempty"`
}

func withPoolHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	_, err := poolDB.ExecContext(r.Context(), "SELECT SLEEP(0.001)")

	resp := Response{
		Mode:       "with-pool",
		DurationMs: float64(time.Since(start).Microseconds()) / 1000.0,
	}
	if err != nil {
		resp.Error = err.Error()
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func withoutPoolHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// TODO(human): コネクションプールなしの実装をここに追加してください
	// 毎リクエストで新しい sql.DB を生成し、クエリ実行後にクローズします
	// dsn 変数を使って接続を開いてください
	// ヒント: SetMaxIdleConns(0) でアイドル接続を保持しないようにできます

	resp := Response{
		Mode:       "without-pool",
		DurationMs: float64(time.Since(start).Microseconds()) / 1000.0,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
