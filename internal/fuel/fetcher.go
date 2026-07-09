package fuel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"time"
)

const okURL = "https://www.ok.dk/privat/produkter/ok-kort/prisudvikling/getProduktHistorik"

type okPriceHistoryItem struct {
	Dato string  `json:"dato"`
	Pris float64 `json:"pris"`
}

type okPriceHistoryResponse struct {
	Historik []okPriceHistoryItem `json:"historik"`
}

// Fetcher pulls the published price history from OK.dk and records new and
// revised prices.
type Fetcher struct {
	store  *Store
	client *http.Client
}

func NewFetcher(store *Store) *Fetcher {
	return &Fetcher{store: store, client: &http.Client{Timeout: 30 * time.Second}}
}

// Fetch updates every fuel type. A failure for one type is logged and does not
// stop the others.
func (f *Fetcher) Fetch(ctx context.Context) {
	for _, fuelType := range AllTypes {
		if err := f.fetchType(ctx, fuelType); err != nil {
			slog.Error("ok failed", "fuelType", fuelType, "error", err)
		}
	}
}

func (f *Fetcher) fetchType(ctx context.Context, fuelType Type) error {
	resp, err := f.fetchHistory(ctx, fuelType)
	if err != nil {
		return err
	}
	prices, err := f.newPrices(ctx, fuelType, resp)
	if err != nil {
		return err
	}
	slog.Info("OK data job", "fuelType", fuelType, "newPrices", len(prices))
	return f.store.UpsertPrices(ctx, prices)
}

func (f *Fetcher) fetchHistory(ctx context.Context, fuelType Type) (*okPriceHistoryResponse, error) {
	body, err := json.Marshal(map[string]any{
		"varenr":    fuelType.okItemNumber(),
		"pumpepris": "true",
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, okURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("OK returned %d", resp.StatusCode)
	}

	var history okPriceHistoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&history); err != nil {
		return nil, err
	}
	return &history, nil
}

// newPrices returns the prices that need writing: days we have never seen, and
// days whose published price has changed since we last looked. For a changed
// price the superseded value is appended to prevPrices.
func (f *Fetcher) newPrices(ctx context.Context, fuelType Type, resp *okPriceHistoryResponse) ([]Price, error) {
	stored, err := f.store.Prices(ctx, fuelType)
	if err != nil {
		return nil, err
	}
	// Keyed by calendar day rather than time.Time, whose equality also compares
	// the monotonic reading and location pointer.
	byDate := make(map[string]Price, len(stored))
	for _, p := range stored {
		byDate[p.Date.Format(time.DateOnly)] = p
	}

	now := time.Now().UTC()
	var prices []Price
	for _, okPrice := range resp.Historik {
		// OK publishes naive timestamps; they have always been read as UTC.
		date, err := time.ParseInLocation(dateLayout, okPrice.Dato, time.UTC)
		if err != nil {
			slog.Error("skipping unparseable date", "dato", okPrice.Dato, "error", err)
			continue
		}

		current, seen := byDate[date.Format(time.DateOnly)]
		switch {
		case !seen:
			prices = append(prices, Price{FuelType: fuelType, Date: date, Price: okPrice.Pris})
		case current.Price != okPrice.Pris:
			prevPrices := append(slices.Clip(current.PrevPrices), PreviousPrice{
				DetectionTimestamp: now,
				Price:              current.Price,
			})
			prices = append(prices, Price{
				FuelType:   fuelType,
				Date:       date,
				Price:      okPrice.Pris,
				PrevPrices: prevPrices,
			})
		}
	}
	return prices, nil
}
