package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/lib/pq"

	"github.com/jfgrea27/meteo/internal/weather"
)

type pgDB interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	Close() error
}

type PgDatabase struct {
	db  pgDB
	log *slog.Logger
}

func ConstructPgDatabase(log *slog.Logger, connStr, schema string) *PgDatabase {
	if connStr == "" {
		panic("DATABASE_URL is required for postgres provider")
	}

	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(fmt.Sprintf("failed to open postgres connection: %v", err))
	}

	if err := conn.Ping(); err != nil {
		panic(fmt.Sprintf("failed to ping postgres: %v", err))
	}

	if schema != "" {
		if _, err := conn.Exec(fmt.Sprintf("SET search_path TO %s", schema)); err != nil {
			panic(fmt.Sprintf("failed to set search_path to %s: %v", schema, err))
		}
		log.Info("connected to postgres", "schema", schema)
	} else {
		log.Info("connected to postgres")
	}

	return &PgDatabase{db: conn, log: log.With("service", "postgres")}
}

func (p *PgDatabase) SaveWeatherEntry(ctx context.Context, entry *weather.WeatherEntry) error {
	p.log.Debug("inserting weather entry", "city", entry.City, "time", entry.Time)

	_, err := p.db.ExecContext(ctx,
		`INSERT INTO weather (time, city, temperature, pressure, humidity, wind_speed, uv, description)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		entry.Time, string(entry.City), entry.Temperature, entry.Pressure,
		entry.Humidity, entry.WindSpeed, entry.UV, entry.Description,
	)
	if err != nil {
		p.log.Error("failed to insert weather entry", "city", entry.City, "error", err)
		return fmt.Errorf("failed to insert weather entry: %w", err)
	}

	p.log.Info("saved weather entry", "city", entry.City, "time", entry.Time)
	return nil
}

func (p *PgDatabase) GetCurrentWeather(ctx context.Context, city weather.City) (*weather.WeatherEntry, error) {
	p.log.Debug("fetching current weather", "city", city)

	row := p.db.QueryRowContext(ctx,
		`SELECT time, city, temperature, pressure, humidity, wind_speed, uv, description
		 FROM weather
		 WHERE city = $1
		 ORDER BY time DESC
		 LIMIT 1`,
		string(city),
	)

	var entry weather.WeatherEntry
	var cityStr string
	err := row.Scan(
		&entry.Time, &cityStr, &entry.Temperature, &entry.Pressure,
		&entry.Humidity, &entry.WindSpeed, &entry.UV, &entry.Description,
	)
	if err != nil {
		p.log.Error("failed to get current weather", "city", city, "error", err)
		return nil, fmt.Errorf("failed to get current weather for %s: %w", city, err)
	}
	entry.City = weather.City(cityStr)

	p.log.Info("fetched current weather", "city", city, "time", entry.Time)
	return &entry, nil
}

func (p *PgDatabase) GetHistoricalWeather(ctx context.Context, city weather.City, from, to time.Time) ([]weather.WeatherEntry, error) {
	p.log.Debug("fetching historical weather", "city", city, "from", from, "to", to)

	rows, err := p.db.QueryContext(ctx,
		`SELECT time, city, temperature, pressure, humidity, wind_speed, uv, description
		 FROM weather
		 WHERE city = $1 AND time >= $2 AND time <= $3
		 ORDER BY time ASC`,
		string(city), from, to,
	)
	if err != nil {
		p.log.Error("failed to query historical weather", "city", city, "error", err)
		return nil, fmt.Errorf("failed to query historical weather for %s: %w", city, err)
	}
	defer rows.Close()

	var entries []weather.WeatherEntry
	for rows.Next() {
		var entry weather.WeatherEntry
		var cityStr string
		if err := rows.Scan(
			&entry.Time, &cityStr, &entry.Temperature, &entry.Pressure,
			&entry.Humidity, &entry.WindSpeed, &entry.UV, &entry.Description,
		); err != nil {
			return nil, fmt.Errorf("failed to scan weather entry: %w", err)
		}
		entry.City = weather.City(cityStr)
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating weather rows: %w", err)
	}

	p.log.Info("fetched historical weather", "city", city, "entries", len(entries))
	return entries, nil
}

func (p *PgDatabase) Close() error {
	return p.db.Close()
}
