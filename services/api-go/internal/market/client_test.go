package market

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOnlineProvider_YahooQuoteVIX(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"chart": {
				"result": [{
					"meta": {
						"currency": "USD",
						"symbol": "^VIX",
						"exchangeTimezoneName": "America/Chicago",
						"regularMarketTime": 1795641301,
						"regularMarketPrice": 28.29,
						"regularMarketChangePercent": 7.322
					}
				}]
			}
		}`))
	}))
	defer srv.Close()

	seed := NewSeedProvider(t.TempDir())
	provider := NewOnlineProvider(seed)
	provider.yahooBaseURL = srv.URL + "/"

	quotes := provider.GetQuotesBatch(context.Background(), []string{"VIX"})
	if len(quotes) != 1 {
		t.Fatalf("expected 1 quote, got %d", len(quotes))
	}
	q := quotes[0]
	if q.Symbol != "VIX" || q.Name != "恐慌指数" || q.Category != "global" {
		t.Fatalf("unexpected quote identity: %+v", q)
	}
	if q.Price != 28.29 || q.ChangePct != 7.32 {
		t.Fatalf("unexpected price/changePct: price=%v changePct=%v", q.Price, q.ChangePct)
	}
	if q.AsOf != "2026-11-25" || q.Meta.Status != "ok" {
		t.Fatalf("expected as_of=2026-11-25 and status=ok, got as_of=%q status=%q", q.AsOf, q.Meta.Status)
	}
}

func TestOnlineProvider_EastmoneyCN10Y(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"success": true,
			"result": {
				"data": [
					{"SOLAR_DATE": "2026-11-25 00:00:00", "EMM00166466": 1.5914},
					{"SOLAR_DATE": "2026-11-24 00:00:00", "EMM00166466": 1.5853}
				]
			}
		}`))
	}))
	defer srv.Close()

	seed := NewSeedProvider(t.TempDir())
	provider := NewOnlineProvider(seed)
	provider.eastmoneyBondURL = srv.URL

	q := provider.GetQuote(context.Background(), "CN10Y")
	if q.Symbol != "CN10Y" || q.Price != 1.5914 || q.AsOf != "2026-11-25" || q.Meta.Status != "ok" {
		t.Fatalf("unexpected CN10Y quote: %+v", q)
	}
}
