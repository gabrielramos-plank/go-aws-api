package server

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go-aws-api/internal/items"
)

type fakeRepo struct {
	items []items.Item
}

func (f *fakeRepo) List(context.Context) ([]items.Item, error) {
	return f.items, nil
}

func (f *fakeRepo) Create(_ context.Context, name string) (items.Item, error) {
	it := items.Item{ID: int64(len(f.items) + 1), Name: name, CreatedAt: time.Now()}
	f.items = append(f.items, it)
	return it, nil
}

type fakePinger struct {
	err error
}

func (f fakePinger) Ping(context.Context) error {
	return f.err
}

func newTestServer() http.Handler {
	return newTestServerWithDB(fakePinger{})
}

func newTestServerWithDB(db Pinger) http.Handler {
	return New(slog.New(slog.NewTextHandler(io.Discard, nil)), &fakeRepo{}, db)
}

func TestHealth(t *testing.T) {
	rec := httptest.NewRecorder()
	newTestServer().ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/health", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestReady(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"database up", nil, http.StatusOK},
		{"database down", errors.New("connection refused"), http.StatusServiceUnavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/ready", nil)
			newTestServerWithDB(fakePinger{err: tt.err}).ServeHTTP(rec, req)

			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tt.want, rec.Body)
			}
		})
	}
}

func TestHello(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  string
	}{
		{"no name", "", `{"message":"Hello World"}`},
		{"empty name", "?name=", `{"message":"Hello World"}`},
		{"with name", "?name=Gabriel", `{"message":"Hello World, Gabriel"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			newTestServer().ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/hello"+tt.query, nil))

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
			}
			if got := strings.TrimSpace(rec.Body.String()); got != tt.want {
				t.Fatalf("body = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestCreateItem(t *testing.T) {
	tests := []struct {
		name string
		body string
		want int
	}{
		{"valid", `{"name":"first"}`, http.StatusCreated},
		{"empty name", `{"name":"  "}`, http.StatusBadRequest},
		{"invalid json", `{`, http.StatusBadRequest},
		{"unknown field", `{"name":"x","extra":1}`, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/items", strings.NewReader(tt.body))
			newTestServer().ServeHTTP(rec, req)

			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tt.want, rec.Body)
			}
		})
	}
}

func TestListItemsEmptyReturnsArray(t *testing.T) {
	rec := httptest.NewRecorder()
	newTestServer().ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/items", nil))

	if got := strings.TrimSpace(rec.Body.String()); got != "[]" {
		t.Fatalf("body = %q, want []", got)
	}
}
