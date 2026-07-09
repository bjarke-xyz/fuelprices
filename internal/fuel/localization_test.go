package fuel

import "testing"

func dayPrices(today float64, neighbours ...float64) DayPrices {
	p := DayPrices{Today: &Price{Price: today}}
	if len(neighbours) > 0 {
		p.Yesterday = &Price{Price: neighbours[0]}
	}
	if len(neighbours) > 1 {
		p.Tomorrow = &Price{Price: neighbours[1]}
	}
	return p
}

func TestText(t *testing.T) {
	tests := []struct {
		name     string
		lang     Language
		fuelType Type
		prices   DayPrices
		want     string
	}{
		{
			name: "english today only", lang: English, fuelType: Diesel,
			prices: dayPrices(15.89),
			want:   "Today, the price of Diesel is 15.89 kroner.",
		},
		{
			name: "english all three", lang: English, fuelType: Unleaded95,
			prices: dayPrices(15.89, 15.79, 15.99),
			want: "Today, the price of Unleaed octane 95 is 15.89 kroner." +
				" Yesterday the price was lower: 15.79 kroner." +
				" Tomorrow the price will be higher: 15.99 kroner.",
		},
		{
			name: "english unchanged neighbours", lang: English, fuelType: Octane100,
			prices: dayPrices(18.09, 18.09, 18.09),
			want: "Today, the price of Octane 100 is 18.09 kroner." +
				" Yesterday the price was the same: 18.09 kroner." +
				" Tomorrow the price will be the same: 18.09 kroner.",
		},
		{
			// A zero-priced neighbour is treated as absent.
			name: "english skips zero neighbour", lang: English, fuelType: Diesel,
			prices: dayPrices(15.89, 0, 0),
			want:   "Today, the price of Diesel is 15.89 kroner.",
		},
		{
			name: "danish today only", lang: Danish, fuelType: Unleaded95,
			prices: dayPrices(15.89),
			want:   "Blyfri oktan 95 koster 15 kroner og 89 ører i dag.",
		},
		{
			name: "danish all three", lang: Danish, fuelType: Diesel,
			prices: dayPrices(15.89, 15.99, 15.79),
			want: "Diesel koster 15 kroner og 89 ører i dag." +
				" I går var prisen højere: 15 kroner og 99 ører." +
				" I morgen vil prisen være lavere: 15 kroner og 79 ører.",
		},
		{
			// The leading zero is stripped rather than the fraction scaled, so
			// 4.03 is spoken as "3 ører". Faithful to the original.
			name: "danish strips leading zero from ore", lang: Danish, fuelType: Diesel,
			prices: dayPrices(4.03),
			want:   "Diesel koster 4 kroner og 3 ører i dag.",
		},
		{
			// 4.10 renders as "4.1", so a single ore digit likewise.
			name: "danish trailing zero is dropped by float formatting", lang: Danish, fuelType: Diesel,
			prices: dayPrices(4.10),
			want:   "Diesel koster 4 kroner og 1 ører i dag.",
		},
		{
			// A whole number has no decimal point and yields the error text.
			name: "danish whole number errors", lang: Danish, fuelType: Diesel,
			prices: dayPrices(4),
			want:   ErrorText(Danish),
		},
		{
			name: "danish whole-number neighbour errors", lang: Danish, fuelType: Diesel,
			prices: dayPrices(15.89, 16),
			want:   ErrorText(Danish),
		},
		{
			name: "english whole number is fine", lang: English, fuelType: Diesel,
			prices: dayPrices(4),
			want:   "Today, the price of Diesel is 4 kroner.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Text(tt.lang, tt.prices, tt.fuelType); got != tt.want {
				t.Errorf("Text() =\n %q\nwant\n %q", got, tt.want)
			}
		})
	}
}

func TestParseLanguage(t *testing.T) {
	for in, want := range map[string]Language{
		"da": Danish, "DA": Danish, "Da": Danish,
		"en": English, "": English, "fr": English, "danish": English,
	} {
		if got := ParseLanguage(in); got != want {
			t.Errorf("ParseLanguage(%q) = %q, want %q", in, got, want)
		}
	}
}
