package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	_ "github.com/spinelli/encurtador-links/docs"
	"github.com/spinelli/encurtador-links/features/redirect"
	"github.com/spinelli/encurtador-links/features/shorten"
	"github.com/spinelli/encurtador-links/features/stats"
	"github.com/spinelli/encurtador-links/shared/config"
	"github.com/spinelli/encurtador-links/shared/encoder"
	"github.com/spinelli/encurtador-links/shared/idgen"
	"github.com/spinelli/encurtador-links/shared/metrics"
	"github.com/spinelli/encurtador-links/shared/middleware"
	"github.com/spinelli/encurtador-links/shared/repository"
)

// @title Encurtador de URLs API
// @version 1.0
// @description API para encurtar URLs usando Redis + PostgreSQL
// @host localhost:8080
// @BasePath /
func main() {
	cfg := config.Load()
	ctx := context.Background()

	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("REDIS_URL inválida: %v", err)
	}
	rdb := redis.NewClient(opt)
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("falha ao conectar no Redis: %v", err)
	}

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("falha ao conectar no PostgreSQL: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("falha ao pingar PostgreSQL: %v", err)
	}

	dbRead, err := sql.Open("postgres", cfg.DatabaseReadURL)
	if err != nil {
		log.Fatalf("falha ao conectar no PostgreSQL (read): %v", err)
	}
	defer dbRead.Close()

	repo := repository.NewPostgres(db)
	repoRead := repository.NewPostgres(dbRead)
	if err := repo.Migrate(); err != nil {
		log.Fatalf("falha ao rodar migration: %v", err)
	}

	idGenerator := idgen.New(rdb, cfg.RedisStartOffset)
	if err := idGenerator.Init(ctx); err != nil {
		log.Fatalf("falha ao inicializar gerador de IDs: %v", err)
	}

	enc, err := encoder.New(cfg.HashSalt, cfg.HashPepper, cfg.HashMinLength)
	if err != nil {
		log.Fatalf("falha ao criar encoder: %v", err)
	}

	shortenService := shorten.NewService(cfg.BaseURL, idGenerator, enc, repo)
	shortenHandler := shorten.NewHandler(shortenService)

	redirectService := redirect.NewService(repoRead, rdb)
	redirectHandler := redirect.NewHandler(redirectService)

	statsHandler := stats.NewHandler()

	mux := http.NewServeMux()
	mux.Handle("/shorten", shortenHandler)
	mux.Handle("/stats", statsHandler)
	mux.Handle("/swagger/", httpSwagger.WrapHandler)
	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/", redirectHandler)

	handler := middleware.CORS(cfg.AllowedOrigins)(metrics.Middleware(mux))

	fmt.Printf("Servidor rodando em %s\n", cfg.BaseURL)
	fmt.Printf("Swagger: %s/swagger/index.html\n", cfg.BaseURL)
	log.Fatal(http.ListenAndServe(cfg.ServerPort, handler))
}
