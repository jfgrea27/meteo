package db

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/jfgrea27/meteo/internal/weather"
)

type mockPgDB struct {
	execContextFunc     func(ctx context.Context, query string, args ...any) (sql.Result, error)
	queryRowContextFunc func(ctx context.Context, query string, args ...any) *sql.Row
	queryContextFunc    func(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

func (m *mockPgDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return m.execContextFunc(ctx, query, args...)
}

func (m *mockPgDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return m.queryRowContextFunc(ctx, query, args...)
}

func (m *mockPgDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return m.queryContextFunc(ctx, query, args...)
}

func (m *mockPgDB) Close() error {
	return nil
}

type mockResult struct{}

func (m mockResult) LastInsertId() (int64, error) { return 0, nil }
func (m mockResult) RowsAffected() (int64, error) { return 1, nil }

func newTestPgDatabase(mock *mockPgDB) *PgDatabase {
	return &PgDatabase{
		db:  mock,
		log: slog.Default().With("service", "postgres"),
	}
}

func TestSaveWeatherEntry(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var capturedArgs []any
		mock := &mockPgDB{
			execContextFunc: func(ctx context.Context, query string, args ...any) (sql.Result, error) {
				capturedArgs = args
				return mockResult{}, nil
			},
		}

		db := newTestPgDatabase(mock)
		entry := &weather.WeatherEntry{
			Time:        time.Unix(1700000000, 0),
			City:        "London",
			Temperature: 18.0,
			Pressure:    1015.0,
			Humidity:    65.0,
			WindSpeed:   6.0,
			UV:          3.0,
			Description: "few clouds",
		}

		err := db.SaveWeatherEntry(context.Background(), entry)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(capturedArgs) != 8 {
			t.Fatalf("expected 8 args, got %d", len(capturedArgs))
		}
		if capturedArgs[1] != "London" {
			t.Errorf("city arg = %v, want London", capturedArgs[1])
		}
	})

	t.Run("exec error", func(t *testing.T) {
		execErr := errors.New("insert failed")
		mock := &mockPgDB{
			execContextFunc: func(ctx context.Context, query string, args ...any) (sql.Result, error) {
				return nil, execErr
			},
		}

		db := newTestPgDatabase(mock)
		entry := &weather.WeatherEntry{
			Time: time.Unix(1700000000, 0),
			City: "London",
		}

		err := db.SaveWeatherEntry(context.Background(), entry)
		if !errors.Is(err, execErr) {
			t.Fatalf("expected exec error, got %v", err)
		}
	})
}

func TestGetHistoricalWeather(t *testing.T) {
	t.Run("query error", func(t *testing.T) {
		queryErr := errors.New("query failed")
		mock := &mockPgDB{
			queryContextFunc: func(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
				return nil, queryErr
			},
		}

		db := newTestPgDatabase(mock)
		_, err := db.GetHistoricalWeather(context.Background(), "London", time.Now().Add(-time.Hour), time.Now())
		if !errors.Is(err, queryErr) {
			t.Fatalf("expected query error, got %v", err)
		}
	})
}

func TestClose(t *testing.T) {
	mock := &mockPgDB{}
	db := newTestPgDatabase(mock)
	err := db.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
