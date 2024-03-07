package main

import (
	"ScoreTableApi/internal/data"
	"ScoreTableApi/internal/jsonlog"
	"context"
	"database/sql"
	"flag"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

type config struct {
	port int
	env  string
	db   struct {
		dsn string
	}
}

type application struct {
	logger *jsonlog.Logger
	config config
	models data.Models
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 8008, "http server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "dsn", os.Getenv("SCORETABLE_DSN"), "DB connection string")

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)
	logger.PrintInfo("starting scoretable server", map[string]string{
		"port": strconv.Itoa(cfg.port),
		"env":  cfg.env,
	})

	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	defer db.Close()
	logger.PrintInfo("database connection pool established", nil)

	app := &application{
		logger: logger,
		config: cfg,
		models: data.NewModels(db),
	}

	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(app.config.port),
		Handler: app.routes(),
	}

	err = srv.ListenAndServe()
	app.logger.PrintFatal(err, nil)
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
