package main

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// tzdata is embedded so Europe/Copenhagen resolves in a scratch image.
	_ "time/tzdata"

	"github.com/bjarke-xyz/fuelprices/internal/fuel"
	"github.com/bjarke-xyz/fuelprices/internal/logging"
	"github.com/bjarke-xyz/fuelprices/internal/server"
	_ "modernc.org/sqlite"
)

//go:embed all:dist
var distFS embed.FS

//go:embed migrations/01_create_fuelprices.sql
var schema string

func main() {
	logging.Setup()
	if err := run(); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		dbPath = env("DB_PATH", "./fuelprices.db")
		port   = env("PORT", "3000")
		jobKey = os.Getenv("JOB_KEY")
	)
	if jobKey == "" {
		slog.Warn("JOB_KEY is not set; /api/job will reject every request")
	}

	// journal_mode is set explicitly, not left to the file: it is persisted in
	// the database header, and the Turso export arrived in WAL mode. A WAL
	// database cannot be opened from a read-only mount, because SQLite must
	// create the -shm wal-index, and the backup job mounts it read-only. There
	// is one writer, once a day, so WAL would buy nothing here anyway.
	db, err := sql.Open("sqlite", "file:"+dbPath+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(delete)")
	if err != nil {
		return err
	}
	defer db.Close()

	if _, err := db.Exec(schema); err != nil {
		return err
	}

	dist, err := fs.Sub(distFS, "dist")
	if err != nil {
		return err
	}

	store := fuel.NewStore(db)
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           server.New(store, fuel.NewFetcher(store), jobKey, dist).Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()

	slog.Info("listening", "addr", srv.Addr, "db", dbPath)
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
