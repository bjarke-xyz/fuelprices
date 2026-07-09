package fuel

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"
)

// storeLayout is how dates and detection timestamps are written. Every
// pre-existing row uses it; see ParseDay.
const storeLayout = time.RFC3339

// upsertChunk bounds how many rows go into one INSERT. SQLite caps the number
// of bound variables per statement and a first run inserts thousands of rows.
const upsertChunk = 200

// Store reads and writes fuel prices.
type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// PricesForDay returns the prices for the given day and its neighbours. It
// returns nil when there is no price for the day itself.
func (s *Store) PricesForDay(ctx context.Context, fuelType Type, day time.Time) (*DayPrices, error) {
	yesterday, tomorrow := AddDays(day, -1), AddDays(day, 1)

	prices, err := s.PricesBetween(ctx, fuelType, yesterday, tomorrow)
	if err != nil {
		return nil, err
	}

	var dayPrices DayPrices
	for i := range prices {
		p := &prices[i]
		switch {
		case p.Date.Equal(day):
			dayPrices.Today = p
		case p.Date.Equal(yesterday):
			dayPrices.Yesterday = p
		case p.Date.Equal(tomorrow):
			dayPrices.Tomorrow = p
		}
	}
	if dayPrices.Today == nil {
		return nil, nil
	}
	return &dayPrices, nil
}

// PricesBetween returns the prices in [from, to], ordered by date.
func (s *Store) PricesBetween(ctx context.Context, fuelType Type, from, to time.Time) ([]Price, error) {
	return s.query(ctx,
		`SELECT fueltype, date, price, prevPrices FROM fuelprices
		 WHERE fueltype = ? AND date >= ? AND date <= ? ORDER BY date`,
		fuelType.Int(), from.UTC().Format(storeLayout), to.UTC().Format(storeLayout))
}

// Prices returns every price for a fuel type.
func (s *Store) Prices(ctx context.Context, fuelType Type) ([]Price, error) {
	return s.query(ctx,
		`SELECT fueltype, date, price, prevPrices FROM fuelprices
		 WHERE fueltype = ? ORDER BY date`, fuelType.Int())
}

func (s *Store) query(ctx context.Context, q string, args ...any) ([]Price, error) {
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prices []Price
	for rows.Next() {
		var (
			fuelTypeInt int
			date        string
			price       Price
			prevPrices  string
		)
		if err := rows.Scan(&fuelTypeInt, &date, &price.Price, &prevPrices); err != nil {
			return nil, err
		}
		if price.FuelType, err = TypeFromInt(fuelTypeInt); err != nil {
			return nil, err
		}
		if price.Date, err = time.Parse(storeLayout, date); err != nil {
			return nil, fmt.Errorf("parse date %q: %w", date, err)
		}
		if err := json.Unmarshal([]byte(prevPrices), &price.PrevPrices); err != nil {
			return nil, fmt.Errorf("parse prevPrices %q: %w", prevPrices, err)
		}
		prices = append(prices, price)
	}
	return prices, rows.Err()
}

// UpsertPrices inserts the given prices, overwriting the price and prevPrices
// of any day already present.
func (s *Store) UpsertPrices(ctx context.Context, prices []Price) error {
	if len(prices) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for chunk := range slices.Chunk(prices, upsertChunk) {
		placeholders := strings.TrimSuffix(strings.Repeat("(?,?,?,?),", len(chunk)), ",")
		args := make([]any, 0, len(chunk)*4)
		for _, p := range chunk {
			prevPrices, err := json.Marshal(nonNil(p.PrevPrices))
			if err != nil {
				return err
			}
			args = append(args, p.FuelType.Int(), p.Date.UTC().Format(storeLayout), p.Price, string(prevPrices))
		}
		_, err := tx.ExecContext(ctx,
			`INSERT INTO fuelprices (fueltype, date, price, prevPrices) VALUES `+placeholders+
				` ON CONFLICT(fueltype, date) DO UPDATE SET
				   price = excluded.price, prevPrices = excluded.prevPrices`, args...)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

// nonNil keeps an empty prevPrices list marshalling to [] rather than null,
// matching every existing row.
func nonNil(p []PreviousPrice) []PreviousPrice {
	if p == nil {
		return []PreviousPrice{}
	}
	return p
}
