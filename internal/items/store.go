package items

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Item struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) List(ctx context.Context) ([]Item, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, name, created_at FROM items ORDER BY id DESC LIMIT 100`)
	if err != nil {
		return nil, fmt.Errorf("query items: %w", err)
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[Item])
}

func (s *Store) Create(ctx context.Context, name string) (Item, error) {
	var it Item
	err := s.pool.QueryRow(ctx,
		`INSERT INTO items (name) VALUES ($1) RETURNING id, name, created_at`, name,
	).Scan(&it.ID, &it.Name, &it.CreatedAt)
	if err != nil {
		return Item{}, fmt.Errorf("insert item: %w", err)
	}
	return it, nil
}
