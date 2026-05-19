package todo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("todo not found")

type Item struct {
	ID    int64
	Title string
	Done  bool
}

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) Init(ctx context.Context) error {
	schema := `
CREATE TABLE IF NOT EXISTS todos (
	id BIGSERIAL PRIMARY KEY,
	title TEXT NOT NULL,
	done BOOLEAN NOT NULL DEFAULT FALSE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`
	if _, err := s.pool.Exec(ctx, schema); err != nil {
		return fmt.Errorf("create schema: %w", err)
	}

	var count int64
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM todos`).Scan(&count); err != nil {
		return fmt.Errorf("count todos: %w", err)
	}

	if count == 0 {
		defaults := []Item{
			{Title: "Write the first HTMX component", Done: false},
			{Title: "Persist tasks in PostgreSQL", Done: true},
		}
		for _, item := range defaults {
			if _, err := s.pool.Exec(ctx, `
INSERT INTO todos (title, done)
VALUES ($1, $2)
`, item.Title, item.Done); err != nil {
				return fmt.Errorf("seed todo: %w", err)
			}
		}
	}

	return nil
}

func (s *Store) List(ctx context.Context) ([]Item, error) {
	rows, err := s.pool.Query(ctx, `
SELECT id, title, done
FROM todos
ORDER BY done ASC, created_at ASC, id ASC
`)
	if err != nil {
		return nil, fmt.Errorf("list todos: %w", err)
	}
	defer rows.Close()

	items := make([]Item, 0)
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.Title, &item.Done); err != nil {
			return nil, fmt.Errorf("scan todo: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate todos: %w", err)
	}

	return items, nil
}

func (s *Store) Create(ctx context.Context, title string) (Item, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Item{}, fmt.Errorf("title is required")
	}

	var item Item
	if err := s.pool.QueryRow(ctx, `
INSERT INTO todos (title, done)
VALUES ($1, FALSE)
RETURNING id, title, done
`, title).Scan(&item.ID, &item.Title, &item.Done); err != nil {
		return Item{}, fmt.Errorf("create todo: %w", err)
	}

	return item, nil
}

func (s *Store) Toggle(ctx context.Context, id int64) (Item, error) {
	var item Item
	err := s.pool.QueryRow(ctx, `
UPDATE todos
SET done = NOT done
WHERE id = $1
RETURNING id, title, done
`, id).Scan(&item.ID, &item.Title, &item.Done)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Item{}, ErrNotFound
		}
		return Item{}, fmt.Errorf("toggle todo: %w", err)
	}

	return item, nil
}

func (s *Store) Delete(ctx context.Context, id int64) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM todos WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete todo: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
