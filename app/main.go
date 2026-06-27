package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"connectionpool-sandbox/adapter/handler"
	"connectionpool-sandbox/adapter/repository"
	"connectionpool-sandbox/usecase"
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
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "root:root@tcp(localhost:3306)/sandbox?parseTime=true"
	}

	poolDB, err := sql.Open("mysql", dsn)
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

	poolRepo := repository.NewPoolRepository(poolDB)
	noPoolRepo := repository.NewNoPoolRepository(dsn)
	withPoolUC := usecase.NewRecordUsecase(poolRepo)
	withoutPoolUC := usecase.NewRecordUsecase(noPoolRepo)
	h := handler.New(withPoolUC, withoutPoolUC)

	mux := http.NewServeMux()
	mux.HandleFunc("/with-pool", h.WithPool)
	mux.HandleFunc("/without-pool", h.WithoutPool)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.Handle("/metrics", promhttp.Handler())

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
