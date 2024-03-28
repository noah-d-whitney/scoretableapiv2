package main

import (
	"ScoreTableApi/internal/data"
	"ScoreTableApi/internal/jsonlog"
	"ScoreTableApi/internal/mailer"
	"context"
	"database/sql"
	"errors"
	"expvar"
	"flag"
	"fmt"
	"os"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

type config struct {
	version string
	port    int
	env     string
	db      struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
	cors struct {
		trustedOrigins []string
	}
}

type application struct {
	logger          *jsonlog.Logger
	config          config
	models          data.Models
	mailer          mailer.Mailer
	gamesInProgress map[string]*data.GameHub
	wg              sync.WaitGroup
}

func main() {
	var cfg config

	// Server Config
	cfg.version = "1.0.0"
	flag.IntVar(&cfg.port, "port", 8008, "http server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	// Database Config
	flag.StringVar(&cfg.db.dsn, "db-dsn", "", "DB connection string")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m",
		"PostgreSQL max connection idle time")

	// Limiter Config
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	// SMTP Config
	flag.StringVar(&cfg.smtp.host, "smtp-host", "sandbox.smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 2525, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "8af8bf145abbd9", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "13aa7a5c47a900", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "ScoreTable <no-reply@scoretable.com>",
		"SMTP sender")

	// CORS Config
	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(val string) error {
		origins := strings.Fields(val)
		if i := slices.Index(origins, "*"); i != -1 {
			return errors.New("cannot set CORS trusted origin to \"*\" with authorization header" +
				" in cross-origin requests")
		}
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})

	// Version
	displayVersion := flag.Bool("version", false, "Show API version and immediately exit")

	flag.Parse()

	if *displayVersion {
		fmt.Printf("Version: %s\n", cfg.version)
		os.Exit(0)
	}

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	defer db.Close()
	logger.PrintInfo("database connection pool established", nil)

	expvar.NewString("version").Set(cfg.version)
	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))
	expvar.Publish("database", expvar.Func(func() any {
		return db.Stats()
	}))
	expvar.Publish("timestamp", expvar.Func(func() any {
		return time.Now().Unix()
	}))

	app := &application{
		logger:          logger,
		config:          cfg,
		models:          data.NewModels(db),
		gamesInProgress: make(map[string]*data.GameHub),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password,
			cfg.smtp.sender),
	}

	go func() {
		for {
			fmt.Printf("%+v\n", app.gamesInProgress)
			time.Sleep(3 * time.Second)
		}
	}()

	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
