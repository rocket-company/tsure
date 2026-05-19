package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"htmx-go-pgsql-todo/apps/web/internal/todo"
)

//go:embed templates/*.html templates/partials/*.html public/*
var embeddedAssets embed.FS

type pageData struct {
	Title string
}

type app struct {
	store     *todo.Store
	templates *template.Template
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg := configFromEnv()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("ping postgres: %v", err)
	}

	store := todo.NewStore(pool)
	if err := store.Init(ctx); err != nil {
		log.Fatalf("init store: %v", err)
	}

	tmpl := template.Must(template.New("").ParseFS(
		embeddedAssets,
		"templates/*.html",
		"templates/partials/*.html",
	))

	a := &app{
		store:     store,
		templates: tmpl,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", a.handleIndex)
	mux.HandleFunc("/todos", a.handleTodos)
	mux.HandleFunc("/todos/", a.handleTodoByID)

	staticFS, err := fs.Sub(embeddedAssets, "public")
	if err != nil {
		log.Fatalf("load static assets: %v", err)
	}
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	srv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      logRequests(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("listening on http://%s", cfg.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("serve: %v", err)
	}
}

type config struct {
	Addr        string
	DatabaseURL string
}

func configFromEnv() config {
	addr := strings.TrimSpace(os.Getenv("ADDR"))
	if addr == "" {
		addr = "127.0.0.1:3456"
	}

	dbURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dbURL == "" {
		dbURL = "postgres://todos:todos@127.0.0.1:5432/todos?sslmode=disable"
	}

	return config{
		Addr:        addr,
		DatabaseURL: dbURL,
	}
}

func (a *app) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	a.render(w, "index", pageData{Title: "HTMX To Do"})
}

func (a *app) handleTodos(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		todos, err := a.store.List(r.Context())
		if err != nil {
			a.internalError(w, err)
			return
		}

		a.render(w, "list", struct {
			Todos []todo.Item
		}{Todos: todos})
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form data", http.StatusBadRequest)
			return
		}

		title := strings.TrimSpace(r.FormValue("newToDo"))
		item, err := a.store.Create(r.Context(), title)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		a.render(w, "todo", item)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func (a *app) handleTodoByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(strings.TrimPrefix(r.URL.Path, "/todos/"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodPut:
		item, err := a.store.Toggle(r.Context(), id)
		if err != nil {
			if errors.Is(err, todo.ErrNotFound) {
				http.NotFound(w, r)
				return
			}
			a.internalError(w, err)
			return
		}

		a.render(w, "todo", item)
	case http.MethodDelete:
		if err := a.store.Delete(r.Context(), id); err != nil {
			if errors.Is(err, todo.ErrNotFound) {
				http.NotFound(w, r)
				return
			}
			a.internalError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func (a *app) render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := a.templates.ExecuteTemplate(w, name, data); err != nil {
		a.internalError(w, err)
	}
}

func (a *app) internalError(w http.ResponseWriter, err error) {
	log.Printf("internal error: %v", err)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func parseID(raw string) (int64, error) {
	raw = strings.Trim(raw, "/")
	if raw == "" {
		return 0, fmt.Errorf("missing id")
	}
	id, err := strconv.ParseInt(path.Base(raw), 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid id")
	}
	return id, nil
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}
