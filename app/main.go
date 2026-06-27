package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	poolDB *sql.DB
	dsn    string
)

var (
	dbOpenConns = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "db_open_connections",
		Help: "Number of open connections to the database",
	}, []string{"mode"})
	dbInUseConns = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "db_in_use_connections",
		Help: "Number of connections currently in use",
	}, []string{"mode"})
	dbIdleConns = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "db_idle_connections",
		Help: "Number of idle connections",
	}, []string{"mode"})
	tableRowCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "db_table_row_count",
		Help: "Exact row count of sandbox.requests table",
	})
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

	poolDB.SetMaxOpenConns(25)
	poolDB.SetMaxIdleConns(25)
	poolDB.SetConnMaxLifetime(5 * time.Minute)

	if err := poolDB.Ping(); err != nil {
		log.Fatalf("ping MySQL: %v", err)
	}
	log.Println("connected to MySQL (with pool)")

	go func() {
		for range time.Tick(2 * time.Second) {
			s := poolDB.Stats()
			dbOpenConns.WithLabelValues("with-pool").Set(float64(s.OpenConnections))
			dbInUseConns.WithLabelValues("with-pool").Set(float64(s.InUse))
			dbIdleConns.WithLabelValues("with-pool").Set(float64(s.Idle))
		}
	}()

	go func() {
		for range time.Tick(5 * time.Second) {
			var count float64
			if err := poolDB.QueryRow("SELECT COUNT(*) FROM requests").Scan(&count); err == nil {
				tableRowCount.Set(count)
			}
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/with-pool", withPoolHandler)
	mux.HandleFunc("/without-pool", withoutPoolHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.Handle("/metrics", promhttp.Handler())

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

	db, err := sql.Open("mysql", dsn)
	if err == nil {
		db.SetMaxOpenConns(1)
		db.SetMaxIdleConns(0)
		defer db.Close()
		_, err = db.ExecContext(r.Context(), "SELECT SLEEP(0.001)")
	}

	resp := Response{
		Mode:       "without-pool",
		DurationMs: float64(time.Since(start).Microseconds()) / 1000.0,
	}
	if err != nil {
		resp.Error = err.Error()
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
