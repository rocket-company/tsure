package orders

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("service order not found")

var statusFlow = []string{
	"orcamento",
	"agendado",
	"em_execucao",
	"aguardando_retorno",
	"finalizado",
}

type ServiceOrder struct {
	ID               int64
	Code             string
	CustomerName     string
	EventName        string
	EventCity        string
	EventDate        time.Time
	Status           string
	CrewSize         int
	VehicleLabel     string
	InventorySummary string
	TotalAmount      float64
	BalanceDue       float64
}

type DashboardSummary struct {
	TotalOrders      int
	OrdersInField    int
	VehiclesReserved int
	OpenReceivables  float64
}

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) Init(ctx context.Context) error {
	schema := `
CREATE TABLE IF NOT EXISTS service_orders (
	id BIGSERIAL PRIMARY KEY,
	code TEXT NOT NULL UNIQUE,
	customer_name TEXT NOT NULL,
	event_name TEXT NOT NULL,
	event_city TEXT NOT NULL,
	event_date DATE NOT NULL,
	status TEXT NOT NULL CHECK (
		status IN (
			'orcamento',
			'agendado',
			'em_execucao',
			'aguardando_retorno',
			'finalizado'
		)
	),
	crew_size INTEGER NOT NULL DEFAULT 1 CHECK (crew_size > 0),
	vehicle_label TEXT NOT NULL,
	inventory_summary TEXT NOT NULL,
	total_amount NUMERIC(12, 2) NOT NULL DEFAULT 0 CHECK (total_amount >= 0),
	balance_due NUMERIC(12, 2) NOT NULL DEFAULT 0 CHECK (balance_due >= 0),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`
	if _, err := s.pool.Exec(ctx, schema); err != nil {
		return fmt.Errorf("create schema: %w", err)
	}

	var count int64
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM service_orders`).Scan(&count); err != nil {
		return fmt.Errorf("count service orders: %w", err)
	}

	if count == 0 {
		defaults := []ServiceOrder{
			{
				Code:             "OS-2026-001",
				CustomerName:     "Casa Aurora Eventos",
				EventName:        "Feira de Noivas Primavera",
				EventCity:        "Cuiaba",
				EventDate:        time.Now().Add(48 * time.Hour),
				Status:           "agendado",
				CrewSize:         6,
				VehicleLabel:     "VUC 3/4 - Placa TSR-2041",
				InventorySummary: "Palco modular, 120 cadeiras Tiffany, kit iluminacao cenario",
				TotalAmount:      18500,
				BalanceDue:       9250,
			},
			{
				Code:             "OS-2026-002",
				CustomerName:     "Grupo Pantanal Experience",
				EventName:        "Convencao de Franqueados",
				EventCity:        "Varzea Grande",
				EventDate:        time.Now(),
				Status:           "em_execucao",
				CrewSize:         10,
				VehicleLabel:     "Truck BaU - Placa TSR-8830",
				InventorySummary: "Painel LED P3, praticaveis, sonorizacao completa",
				TotalAmount:      42750,
				BalanceDue:       0,
			},
			{
				Code:             "OS-2026-003",
				CustomerName:     "Instituto Terra Viva",
				EventName:        "Mutirao de Saude Corporativa",
				EventCity:        "Rondonopolis",
				EventDate:        time.Now().Add(120 * time.Hour),
				Status:           "orcamento",
				CrewSize:         4,
				VehicleLabel:     "Van Operacional - Placa TSR-1108",
				InventorySummary: "Tendas 10x10, climatizadores, mobiliario lounge",
				TotalAmount:      13200,
				BalanceDue:       13200,
			},
		}

		for _, item := range defaults {
			if _, err := s.pool.Exec(ctx, `
INSERT INTO service_orders (
	code,
	customer_name,
	event_name,
	event_city,
	event_date,
	status,
	crew_size,
	vehicle_label,
	inventory_summary,
	total_amount,
	balance_due
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
`,
				item.Code,
				item.CustomerName,
				item.EventName,
				item.EventCity,
				item.EventDate,
				item.Status,
				item.CrewSize,
				item.VehicleLabel,
				item.InventorySummary,
				item.TotalAmount,
				item.BalanceDue,
			); err != nil {
				return fmt.Errorf("seed service order: %w", err)
			}
		}
	}

	return nil
}

func (s *Store) List(ctx context.Context) ([]ServiceOrder, error) {
	rows, err := s.pool.Query(ctx, `
SELECT
	id,
	code,
	customer_name,
	event_name,
	event_city,
	event_date,
	status,
	crew_size,
	vehicle_label,
	inventory_summary,
	total_amount,
	balance_due
FROM service_orders
ORDER BY event_date ASC, created_at ASC, id ASC
`)
	if err != nil {
		return nil, fmt.Errorf("list service orders: %w", err)
	}
	defer rows.Close()

	items := make([]ServiceOrder, 0)
	for rows.Next() {
		var item ServiceOrder
		if err := rows.Scan(
			&item.ID,
			&item.Code,
			&item.CustomerName,
			&item.EventName,
			&item.EventCity,
			&item.EventDate,
			&item.Status,
			&item.CrewSize,
			&item.VehicleLabel,
			&item.InventorySummary,
			&item.TotalAmount,
			&item.BalanceDue,
		); err != nil {
			return nil, fmt.Errorf("scan service order: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate service orders: %w", err)
	}

	return items, nil
}

func (s *Store) Create(ctx context.Context, customerName, eventName, eventCity, eventDateRaw, vehicleLabel string, crewSize int) error {
	customerName = strings.TrimSpace(customerName)
	eventName = strings.TrimSpace(eventName)
	eventCity = strings.TrimSpace(eventCity)
	eventDateRaw = strings.TrimSpace(eventDateRaw)
	vehicleLabel = strings.TrimSpace(vehicleLabel)

	if customerName == "" {
		return fmt.Errorf("cliente e obrigatorio")
	}
	if eventName == "" {
		return fmt.Errorf("evento e obrigatorio")
	}
	if eventCity == "" {
		return fmt.Errorf("cidade e obrigatoria")
	}
	if vehicleLabel == "" {
		return fmt.Errorf("veiculo e obrigatorio")
	}
	if crewSize <= 0 {
		return fmt.Errorf("equipe deve ser maior que zero")
	}

	eventDate, err := time.Parse("2006-01-02", eventDateRaw)
	if err != nil {
		return fmt.Errorf("data do evento invalida")
	}

	code, err := s.nextCode(ctx, eventDate.Year())
	if err != nil {
		return fmt.Errorf("generate code: %w", err)
	}

	totalAmount := float64(crewSize) * 2750
	balanceDue := totalAmount * 0.5

	if _, err := s.pool.Exec(ctx, `
INSERT INTO service_orders (
	code,
	customer_name,
	event_name,
	event_city,
	event_date,
	status,
	crew_size,
	vehicle_label,
	inventory_summary,
	total_amount,
	balance_due
)
VALUES ($1, $2, $3, $4, $5, 'orcamento', $6, $7, $8, $9, $10)
`,
		code,
		customerName,
		eventName,
		eventCity,
		eventDate,
		crewSize,
		vehicleLabel,
		"A definir no planejamento operacional",
		totalAmount,
		balanceDue,
	); err != nil {
		return fmt.Errorf("create service order: %w", err)
	}

	return nil
}

func (s *Store) AdvanceStatus(ctx context.Context, id int64) error {
	var currentStatus string
	if err := s.pool.QueryRow(ctx, `SELECT status FROM service_orders WHERE id = $1`, id).Scan(&currentStatus); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("load service order status: %w", err)
	}

	nextStatus := currentStatus
	for index, status := range statusFlow {
		if status == currentStatus {
			nextStatus = statusFlow[(index+1)%len(statusFlow)]
			break
		}
	}

	tag, err := s.pool.Exec(ctx, `
UPDATE service_orders
SET status = $2, updated_at = NOW()
WHERE id = $1
`, id, nextStatus)
	if err != nil {
		return fmt.Errorf("advance service order status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *Store) Delete(ctx context.Context, id int64) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM service_orders WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete service order: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func BuildSummary(items []ServiceOrder) DashboardSummary {
	summary := DashboardSummary{
		TotalOrders: len(items),
	}

	vehicles := make(map[string]struct{})
	for _, item := range items {
		if item.Status == "em_execucao" || item.Status == "aguardando_retorno" {
			summary.OrdersInField++
		}
		if item.Status == "agendado" || item.Status == "em_execucao" || item.Status == "aguardando_retorno" {
			vehicles[item.VehicleLabel] = struct{}{}
		}
		summary.OpenReceivables += item.BalanceDue
	}

	summary.VehiclesReserved = len(vehicles)
	summary.OpenReceivables = math.Round(summary.OpenReceivables*100) / 100
	return summary
}

func StatusLabel(status string) string {
	switch status {
	case "orcamento":
		return "Orcamento"
	case "agendado":
		return "Agendado"
	case "em_execucao":
		return "Em execucao"
	case "aguardando_retorno":
		return "Aguardando retorno"
	case "finalizado":
		return "Finalizado"
	default:
		return status
	}
}

func NextStatusLabel(status string) string {
	for index, current := range statusFlow {
		if current == status {
			return StatusLabel(statusFlow[(index+1)%len(statusFlow)])
		}
	}
	return "Atualizar status"
}

func (s *Store) nextCode(ctx context.Context, year int) (string, error) {
	var count int
	if err := s.pool.QueryRow(ctx, `
SELECT COUNT(*)
FROM service_orders
WHERE code LIKE $1
`, fmt.Sprintf("OS-%d-%%", year)).Scan(&count); err != nil {
		return "", err
	}

	return fmt.Sprintf("OS-%d-%03d", year, count+1), nil
}
