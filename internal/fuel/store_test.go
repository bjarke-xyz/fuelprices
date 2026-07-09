package fuel

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()

	db, err := sql.Open("sqlite", "file:"+filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	schema, err := os.ReadFile(filepath.Join("..", "..", "migrations", "01_create_fuelprices.sql"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(string(schema)); err != nil {
		t.Fatal(err)
	}
	return NewStore(db)
}

func day(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := time.ParseInLocation(time.DateOnly, s, time.UTC)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// Every one of the 17,926 rows inherited from the previous deployment stores
// its date as UTC midnight with a Z suffix. Writing anything else would give a
// day two primary keys and break the date range queries, which compare strings.
func TestUpsertPricesStoresUTCMidnight(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	err := store.UpsertPrices(ctx, []Price{{
		FuelType: Diesel,
		Date:     day(t, "2026-07-10"),
		Price:    15.89,
	}})
	if err != nil {
		t.Fatal(err)
	}

	var date, prevPrices string
	err = store.db.QueryRow(`SELECT date, prevPrices FROM fuelprices`).Scan(&date, &prevPrices)
	if err != nil {
		t.Fatal(err)
	}
	if want := "2026-07-10T00:00:00Z"; date != want {
		t.Errorf("stored date = %q, want %q", date, want)
	}
	if want := "[]"; prevPrices != want {
		t.Errorf("stored prevPrices = %q, want %q", prevPrices, want)
	}
}

// A day parsed from a query parameter must produce the same stored string as a
// day parsed from the OK.dk feed, or the two would never match.
func TestParseDayMatchesFeedParsing(t *testing.T) {
	fromQuery := ParseDay("2026-07-10", time.Time{})
	fromFeed, err := time.ParseInLocation(dateLayout, "2026-07-10T00:00:00", time.UTC)
	if err != nil {
		t.Fatal(err)
	}
	if !fromQuery.Equal(fromFeed) {
		t.Errorf("query day %v != feed day %v", fromQuery, fromFeed)
	}
	if got, want := fromQuery.Format(storeLayout), "2026-07-10T00:00:00Z"; got != want {
		t.Errorf("stored form = %q, want %q", got, want)
	}
}

// Day resolves the calendar day in Copenhagen, not UTC: just after Danish
// midnight in summer it is still the previous day in UTC.
func TestDayUsesCopenhagenCalendar(t *testing.T) {
	justAfterDanishMidnight := time.Date(2026, 7, 9, 22, 30, 0, 0, time.UTC) // 00:30 CEST on the 10th
	if got, want := Day(justAfterDanishMidnight).Format(time.DateOnly), "2026-07-10"; got != want {
		t.Errorf("Day() = %v, want %v", got, want)
	}
}

func TestUpsertPricesOverwritesExistingDay(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	date := day(t, "2026-07-10")

	if err := store.UpsertPrices(ctx, []Price{{FuelType: Diesel, Date: date, Price: 15.89}}); err != nil {
		t.Fatal(err)
	}
	revised := Price{
		FuelType:   Diesel,
		Date:       date,
		Price:      16.09,
		PrevPrices: []PreviousPrice{{DetectionTimestamp: time.Now().UTC(), Price: 15.89}},
	}
	if err := store.UpsertPrices(ctx, []Price{revised}); err != nil {
		t.Fatal(err)
	}

	prices, err := store.Prices(ctx, Diesel)
	if err != nil {
		t.Fatal(err)
	}
	if len(prices) != 1 {
		t.Fatalf("got %d rows, want 1 (the day should be updated, not duplicated)", len(prices))
	}
	if prices[0].Price != 16.09 {
		t.Errorf("price = %v, want 16.09", prices[0].Price)
	}
	if len(prices[0].PrevPrices) != 1 || prices[0].PrevPrices[0].Price != 15.89 {
		t.Errorf("prevPrices = %+v, want the superseded 15.89", prices[0].PrevPrices)
	}
}

func TestUpsertPricesChunksLargeBatches(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	// More rows than fit in one statement's bound-variable budget.
	var prices []Price
	start := day(t, "1993-01-01")
	for i := range upsertChunk*3 + 7 {
		prices = append(prices, Price{FuelType: Diesel, Date: AddDays(start, i), Price: float64(i) / 100})
	}
	if err := store.UpsertPrices(ctx, prices); err != nil {
		t.Fatal(err)
	}

	got, err := store.Prices(ctx, Diesel)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(prices) {
		t.Errorf("stored %d rows, want %d", len(got), len(prices))
	}
}

func TestPricesForDay(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	// Spans a month boundary, which the day-of-month comparison in the
	// TypeScript version got wrong.
	prices := []Price{
		{FuelType: Diesel, Date: day(t, "2026-06-30"), Price: 15.79},
		{FuelType: Diesel, Date: day(t, "2026-07-01"), Price: 15.89},
		{FuelType: Diesel, Date: day(t, "2026-07-02"), Price: 15.99},
		{FuelType: Octane100, Date: day(t, "2026-07-01"), Price: 18.09},
	}
	if err := store.UpsertPrices(ctx, prices); err != nil {
		t.Fatal(err)
	}

	got, err := store.PricesForDay(ctx, Diesel, day(t, "2026-07-01"))
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("PricesForDay = nil, want prices")
	}
	if got.Today.Price != 15.89 || got.Yesterday.Price != 15.79 || got.Tomorrow.Price != 15.99 {
		t.Errorf("got today=%v yesterday=%v tomorrow=%v; want 15.89/15.79/15.99",
			got.Today.Price, got.Yesterday.Price, got.Tomorrow.Price)
	}
}

func TestPricesForDayWithoutTodayIsNil(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	// Yesterday exists, today does not.
	err := store.UpsertPrices(ctx, []Price{{FuelType: Diesel, Date: day(t, "2026-06-30"), Price: 15.79}})
	if err != nil {
		t.Fatal(err)
	}

	got, err := store.PricesForDay(ctx, Diesel, day(t, "2026-07-01"))
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Errorf("PricesForDay = %+v, want nil when today is missing", got)
	}
}
