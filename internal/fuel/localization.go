package fuel

import (
	"fmt"
	"strconv"
	"strings"
)

// Language is the language of the /api/prices message.
type Language string

const (
	Danish  Language = "da"
	English Language = "en"
)

// ParseLanguage maps a query-string value to a language, defaulting to English.
func ParseLanguage(s string) Language {
	if strings.ToLower(s) == "da" {
		return Danish
	}
	return English
}

// ErrorText is the message returned when no prices exist for the requested day.
func ErrorText(l Language) string {
	if l == Danish {
		return "Der blev ikke fundet priser for den dato"
	}
	return "No prices were found for that date"
}

// Text renders the spoken-word summary of a day's prices.
func Text(l Language, prices DayPrices, fuelType Type) string {
	if l == Danish {
		return textDanish(prices, fuelType)
	}
	return textEnglish(prices, fuelType)
}

func textEnglish(prices DayPrices, fuelType Type) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Today, the price of %s is %s kroner.",
		fuelType.localized(English), formatPrice(prices.Today.Price))

	if prices.Yesterday != nil && prices.Yesterday.Price > 0 {
		fmt.Fprintf(&b, " Yesterday the price was %s: %s kroner.",
			diffText(English, prices.Today.Price, prices.Yesterday.Price),
			formatPrice(prices.Yesterday.Price))
	}
	if prices.Tomorrow != nil && prices.Tomorrow.Price > 0 {
		fmt.Fprintf(&b, " Tomorrow the price will be %s: %s kroner.",
			diffText(English, prices.Today.Price, prices.Tomorrow.Price),
			formatPrice(prices.Tomorrow.Price))
	}
	return b.String()
}

func textDanish(prices DayPrices, fuelType Type) string {
	kroner, orer, ok := kronerAndOrer(prices.Today.Price)
	if !ok {
		return ErrorText(Danish)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s koster %s kroner og %s ører i dag.",
		fuelType.localized(Danish), kroner, orer)

	if prices.Yesterday != nil && prices.Yesterday.Price > 0 {
		kroner, orer, ok := kronerAndOrer(prices.Yesterday.Price)
		if !ok {
			return ErrorText(Danish)
		}
		fmt.Fprintf(&b, " I går var prisen %s: %s kroner og %s ører.",
			diffText(Danish, prices.Today.Price, prices.Yesterday.Price), kroner, orer)
	}
	if prices.Tomorrow != nil && prices.Tomorrow.Price > 0 {
		kroner, orer, ok := kronerAndOrer(prices.Tomorrow.Price)
		if !ok {
			return ErrorText(Danish)
		}
		fmt.Fprintf(&b, " I morgen vil prisen være %s: %s kroner og %s ører.",
			diffText(Danish, prices.Today.Price, prices.Tomorrow.Price), kroner, orer)
	}
	return b.String()
}

func diffText(l Language, today, otherDay float64) string {
	if l == Danish {
		switch {
		case otherDay > today:
			return "højere"
		case otherDay < today:
			return "lavere"
		default:
			return "den samme"
		}
	}
	switch {
	case otherDay > today:
		return "higher"
	case otherDay < today:
		return "lower"
	default:
		return "the same"
	}
}

// formatPrice renders a price the way JavaScript's Number.toString would: the
// shortest representation that round-trips.
func formatPrice(p float64) string {
	return strconv.FormatFloat(p, 'f', -1, 64)
}

// kronerAndOrer splits a price into its whole and fractional parts. It reports
// false for a price with no fractional part, which is how the original
// behaved: 4 has no decimal point and yields the error message, while 4.03
// yields ("4", "3") — the leading zero is stripped rather than the fraction
// being scaled, so 4.03 is spoken as "3 ører".
func kronerAndOrer(p float64) (kroner, orer string, ok bool) {
	whole, frac, found := strings.Cut(formatPrice(p), ".")
	if !found {
		return "", "", false
	}
	if _, err := strconv.Atoi(whole); err != nil {
		return "", "", false
	}
	return whole, strings.TrimLeft(frac, "0"), true
}
