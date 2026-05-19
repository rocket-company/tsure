package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"tsure/apps/web/internal/orders"
)

//go:embed templates/*.html templates/partials/*.html public/*
var embeddedAssets embed.FS

type pageData struct {
	Title           string
	Today           string
	DefaultDate     string
	DefaultVehicle  string
	DefaultCrewSize int
}

type dashboardData struct {
	Orders  []orders.ServiceOrder
	Summary orders.DashboardSummary
}

type app struct {
	store     *orders.Store
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

	store := orders.NewStore(pool)
	if err := store.Init(ctx); err != nil {
		log.Fatalf("init store: %v", err)
	}

	tmpl := template.Must(template.New("").Funcs(template.FuncMap{
		"formatMoney":     formatMoneyBRL,
		"formatDate":      formatDateBR,
		"statusLabel":     orders.StatusLabel,
		"nextStatusLabel": orders.NextStatusLabel,
	}).ParseFS(
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
	mux.HandleFunc("/orders", a.handleOrders)
	mux.HandleFunc("/orders/", a.handleOrderByID)

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
		dbURL = "postgres://tsure:tsure@127.0.0.1:5432/tsure?sslmode=disable"
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

	now := time.Now()
	a.render(w, "index", pageData{
		Title:           "tsure | ERP de leasing para eventos",
		Today:           now.Format("02 Jan 2006"),
		DefaultDate:     now.Add(48 * time.Hour).Format("2006-01-02"),
		DefaultVehicle:  "Van Operacional - Placa TST-1001",
		DefaultCrewSize: 4,
	})
}

func (a *app) handleOrders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.renderOrdersPanel(w, r)
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form data", http.StatusBadRequest)
			return
		}

		crewSize, err := strconv.Atoi(strings.TrimSpace(r.FormValue("crewSize")))
		if err != nil {
			http.Error(w, "equipe invalida", http.StatusBadRequest)
			return
		}

		if err := a.store.Create(
			r.Context(),
			r.FormValue("customerName"),
			r.FormValue("eventName"),
			r.FormValue("eventCity"),
			r.FormValue("eventDate"),
			r.FormValue("vehicleLabel"),
			crewSize,
		); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		a.renderOrdersPanel(w, r)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func (a *app) handleOrderByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(strings.TrimPrefix(r.URL.Path, "/orders/"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodPut:
		if err := a.store.AdvanceStatus(r.Context(), id); err != nil {
			if errors.Is(err, orders.ErrNotFound) {
				http.NotFound(w, r)
				return
			}
			a.internalError(w, err)
			return
		}

		a.renderOrdersPanel(w, r)
	case http.MethodDelete:
		if err := a.store.Delete(r.Context(), id); err != nil {
			if errors.Is(err, orders.ErrNotFound) {
				http.NotFound(w, r)
				return
			}
			a.internalError(w, err)
			return
		}

		a.renderOrdersPanel(w, r)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func (a *app) renderOrdersPanel(w http.ResponseWriter, r *http.Request) {
	data, err := a.dashboard(r.Context())
	if err != nil {
		a.internalError(w, err)
		return
	}
	a.render(w, "orders-panel", data)
}

func (a *app) dashboard(ctx context.Context) (dashboardData, error) {
	items, err := a.store.List(ctx)
	if err != nil {
		return dashboardData{}, err
	}

	return dashboardData{
		Orders:  items,
		Summary: orders.BuildSummary(items),
	}, nil
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

func formatDateBR(value time.Time) string {
	return value.Format("02/01/2006")
}

func formatMoneyBRL(value float64) string {
	reais := int64(math.Round(value * 100))
	inteiro := reais / 100
	centavos := reais % 100
	if centavos < 0 {
		centavos = -centavos
	}

	intPart := strconv.FormatInt(inteiro, 10)
	if len(intPart) > 3 {
		var groups []string
		for len(intPart) > 3 {
			groups = append([]string{intPart[len(intPart)-3:]}, groups...)
			intPart = intPart[:len(intPart)-3]
		}
		groups = append([]string{intPart}, groups...)
		intPart = strings.Join(groups, ".")
	}

	return fmt.Sprintf("R$ %s,%02d", intPart, centavos)
}
