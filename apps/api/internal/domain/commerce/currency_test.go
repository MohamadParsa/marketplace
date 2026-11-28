package commerce

import "testing"

func TestGatewayCurrencyAndExactRateConversion(t *testing.T) {
	if SiteCurrency != CurrencyToman || CurrencyToman.DefaultGatewayRate() != "1" || CurrencyRial.DefaultGatewayRate() != "10" || Currency("USD").Valid() {
		t.Fatal("currency defaults")
	}
	for _, tc := range []struct {
		amount   Amount
		currency Currency
		rate     string
		want     int64
	}{
		{100000, CurrencyToman, "1", 10000},
		{100000, CurrencyRial, "10", 100000},
		{101, CurrencyRial, "10", 101},
		{100000, CurrencyRial, "10.25", 102500},
		{10000000, CurrencyRial, "0.000001", 1},
	} {
		got, err := GatewayAmount(tc.amount, tc.currency, tc.rate)
		if err != nil || got != tc.want {
			t.Fatalf("%+v: %d %v", tc, got, err)
		}
	}
	for _, rate := range []string{"", "0", "-1", "NaN", "1e1", "1/10", "1.0000001", "1000001", " 10", "0.000000"} {
		if _, err := NormalizeGatewayRate(rate); err == nil {
			t.Fatalf("accepted rate %q", rate)
		}
	}
	if got, err := NormalizeGatewayRate("0000010.250000"); err != nil || got != "10.25" {
		t.Fatal(got, err)
	}
	for _, tc := range []struct {
		amount   Amount
		currency Currency
		rate     string
	}{
		{101, CurrencyToman, "1"}, // no silent truncation of 10.1 toman
		{MaxAmount, CurrencyRial, "1000000"},
		{100, Currency("USD"), "1"},
		{0, CurrencyToman, "1"},
		{-10, CurrencyToman, "1"},
	} {
		if _, err := GatewayAmount(tc.amount, tc.currency, tc.rate); err == nil {
			t.Fatalf("accepted invalid conversion %+v", tc)
		}
	}
}
