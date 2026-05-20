package inventory

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Equipment struct {
	ID            int64
	Code          string
	Name          string
	Category      string
	QuantityTotal int
	Status        string
	CreatedAt     time.Time
}

type Kit struct {
	ID          int64
	Code        string
	Name        string
	Description string
	CreatedAt   time.Time
}

type KitItem struct {
	KitID        int64
	EquipmentID  int64
	EquipmentName string
	EquipmentCode string
	Quantity     int
}

type Reservation struct {
	ID             int64
	EquipmentID    int64
	ServiceOrderID int64
	ReservedFrom   time.Time
	ReservedUntil  time.Time
	Quantity       int
	Status         string
	CreatedAt      time.Time
}

type Availability struct {
	EquipmentID   int64
	EquipmentCode string
	EquipmentName string
	TotalQty      int
	ReservedQty   int
	AvailableQty  int
}

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) Init(ctx context.Context) error {
	if _, err := s.pool.Exec(ctx, schema); err != nil {
		return fmt.Errorf("create inventory schema: %w", err)
	}
	return nil
}

// Equipment

func (s *Store) ListEquipment(ctx context.Context) ([]Equipment, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, code, name, category, quantity_total, status, created_at
		FROM equipment
		ORDER BY category, name
	`)
	if err != nil {
		return nil, fmt.Errorf("list equipment: %w", err)
	}
	defer rows.Close()

	var items []Equipment
	for rows.Next() {
		var e Equipment
		if err := rows.Scan(&e.ID, &e.Code, &e.Name, &e.Category, &e.QuantityTotal, &e.Status, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan equipment: %w", err)
		}
		items = append(items, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate equipment: %w", err)
	}
	return items, nil
}

func (s *Store) CreateEquipment(ctx context.Context, code, name, category string, qty int) error {
	if code == "" || name == "" {
		return fmt.Errorf("codigo e nome sao obrigatorios")
	}
	if qty <= 0 {
		qty = 1
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO equipment (code, name, category, quantity_total)
		VALUES ($1, $2, $3, $4)
	`, code, name, category, qty)
	if err != nil {
		return fmt.Errorf("create equipment: %w", err)
	}
	return nil
}

func (s *Store) GetEquipment(ctx context.Context, id int64) (Equipment, error) {
	var e Equipment
	err := s.pool.QueryRow(ctx, `
		SELECT id, code, name, category, quantity_total, status, created_at
		FROM equipment WHERE id = $1
	`, id).Scan(&e.ID, &e.Code, &e.Name, &e.Category, &e.QuantityTotal, &e.Status, &e.CreatedAt)
	if err != nil {
		return e, fmt.Errorf("get equipment: %w", err)
	}
	return e, nil
}

// Kits

func (s *Store) ListKits(ctx context.Context) ([]Kit, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, code, name, description, created_at
		FROM kits
		ORDER BY name
	`)
	if err != nil {
		return nil, fmt.Errorf("list kits: %w", err)
	}
	defer rows.Close()

	var items []Kit
	for rows.Next() {
		var k Kit
		if err := rows.Scan(&k.ID, &k.Code, &k.Name, &k.Description, &k.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan kit: %w", err)
		}
		items = append(items, k)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate kits: %w", err)
	}
	return items, nil
}

func (s *Store) CreateKit(ctx context.Context, code, name, description string) (int64, error) {
	if code == "" || name == "" {
		return 0, fmt.Errorf("codigo e nome sao obrigatorios")
	}
	var id int64
	err := s.pool.QueryRow(ctx, `
		INSERT INTO kits (code, name, description)
		VALUES ($1, $2, $3)
		RETURNING id
	`, code, name, description).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create kit: %w", err)
	}
	return id, nil
}

func (s *Store) AddKitItem(ctx context.Context, kitID, equipmentID int64, qty int) error {
	if qty <= 0 {
		return fmt.Errorf("quantidade deve ser maior que zero")
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO kit_items (kit_id, equipment_id, quantity)
		VALUES ($1, $2, $3)
		ON CONFLICT (kit_id, equipment_id) DO UPDATE SET quantity = EXCLUDED.quantity
	`, kitID, equipmentID, qty)
	if err != nil {
		return fmt.Errorf("add kit item: %w", err)
	}
	return nil
}

func (s *Store) ListKitItems(ctx context.Context, kitID int64) ([]KitItem, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT ki.kit_id, ki.equipment_id, e.name, e.code, ki.quantity
		FROM kit_items ki
		JOIN equipment e ON e.id = ki.equipment_id
		WHERE ki.kit_id = $1
		ORDER BY e.name
	`, kitID)
	if err != nil {
		return nil, fmt.Errorf("list kit items: %w", err)
	}
	defer rows.Close()

	var items []KitItem
	for rows.Next() {
		var ki KitItem
		if err := rows.Scan(&ki.KitID, &ki.EquipmentID, &ki.EquipmentName, &ki.EquipmentCode, &ki.Quantity); err != nil {
			return nil, fmt.Errorf("scan kit item: %w", err)
		}
		items = append(items, ki)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate kit items: %w", err)
	}
	return items, nil
}

// Reservations

func (s *Store) CreateReservation(ctx context.Context, equipmentID, soID int64, fromDate, untilDate time.Time, qty int) error {
	if qty <= 0 {
		return fmt.Errorf("quantidade deve ser maior que zero")
	}
	if untilDate.Before(fromDate) {
		return fmt.Errorf("data final nao pode ser anterior a data inicial")
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO equipment_reservations (equipment_id, service_order_id, reserved_from, reserved_until, quantity)
		VALUES ($1, $2, $3, $4, $5)
	`, equipmentID, soID, fromDate, untilDate, qty)
	if err != nil {
		return fmt.Errorf("create reservation: %w", err)
	}
	return nil
}

func (s *Store) UpdateReservationStatus(ctx context.Context, id int64, status string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE equipment_reservations SET status = $2 WHERE id = $1
	`, id, status)
	if err != nil {
		return fmt.Errorf("update reservation status: %w", err)
	}
	return nil
}

func (s *Store) UpdateReservationsBySO(ctx context.Context, soID int64, status string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE equipment_reservations SET status = $2 WHERE service_order_id = $1
	`, soID, status)
	if err != nil {
		return fmt.Errorf("update reservations by so: %w", err)
	}
	return nil
}

func (s *Store) ListReservationsBySO(ctx context.Context, soID int64) ([]Reservation, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, equipment_id, service_order_id, reserved_from, reserved_until, quantity, status, created_at
		FROM equipment_reservations
		WHERE service_order_id = $1
		ORDER BY reserved_from
	`, soID)
	if err != nil {
		return nil, fmt.Errorf("list reservations: %w", err)
	}
	defer rows.Close()

	var items []Reservation
	for rows.Next() {
		var r Reservation
		if err := rows.Scan(&r.ID, &r.EquipmentID, &r.ServiceOrderID, &r.ReservedFrom, &r.ReservedUntil, &r.Quantity, &r.Status, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan reservation: %w", err)
		}
		items = append(items, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate reservations: %w", err)
	}
	return items, nil
}

// Availability

func (s *Store) IsAvailable(ctx context.Context, equipmentID int64, fromDate, untilDate time.Time, qty int) (bool, error) {
	var reserved int
	if err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(r.quantity), 0)
		FROM equipment_reservations r
		WHERE r.equipment_id = $1
			AND r.status IN ('reservado', 'em_campo', 'aguardando_retorno')
			AND r.reserved_from <= $3
			AND r.reserved_until >= $2
	`, equipmentID, fromDate, untilDate).Scan(&reserved); err != nil {
		return false, fmt.Errorf("check availability: %w", err)
	}
	var total int
	if err := s.pool.QueryRow(ctx, `SELECT quantity_total FROM equipment WHERE id = $1`, equipmentID).Scan(&total); err != nil {
		return false, fmt.Errorf("check total: %w", err)
	}
	return (total - reserved) >= qty, nil
}

func (s *Store) CheckAvailability(ctx context.Context, fromDate, untilDate time.Time) ([]Availability, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			e.id,
			e.code,
			e.name,
			e.quantity_total,
			COALESCE(SUM(r.quantity), 0) AS reserved
		FROM equipment e
		LEFT JOIN equipment_reservations r
			ON r.equipment_id = e.id
			AND r.status IN ('reservado', 'em_campo')
			AND r.reserved_from <= $2
			AND r.reserved_until >= $1
		WHERE e.status = 'disponivel'
		GROUP BY e.id, e.code, e.name, e.quantity_total
		ORDER BY e.category, e.name
	`, fromDate, untilDate)
	if err != nil {
		return nil, fmt.Errorf("check availability: %w", err)
	}
	defer rows.Close()

	var items []Availability
	for rows.Next() {
		var a Availability
		if err := rows.Scan(&a.EquipmentID, &a.EquipmentCode, &a.EquipmentName, &a.TotalQty, &a.ReservedQty); err != nil {
			return nil, fmt.Errorf("scan availability: %w", err)
		}
		a.AvailableQty = a.TotalQty - a.ReservedQty
		if a.AvailableQty < 0 {
			a.AvailableQty = 0
		}
		items = append(items, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate availability: %w", err)
	}
	return items, nil
}
