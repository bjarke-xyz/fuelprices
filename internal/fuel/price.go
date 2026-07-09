package fuel

import "time"

// Copenhagen is the timezone that defines what "today" means: prices are
// published against Danish calendar days.
var Copenhagen = mustLoadLocation("Europe/Copenhagen")

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}
	return loc
}

// dateLayout is the wire format of the OK.dk "dato" field and of the date
// column: a naive timestamp that has always been interpreted as UTC. Every row
// in the database is stored as UTC midnight, so any Day value must be too.
const dateLayout = "2006-01-02T15:04:05"

// Day returns the calendar day containing t (in Copenhagen), as UTC midnight.
func Day(t time.Time) time.Time {
	y, m, d := t.In(Copenhagen).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// ParseDay parses a yyyy-MM-dd query parameter, falling back to def when the
// value is absent or malformed.
func ParseDay(s string, def time.Time) time.Time {
	if s == "" {
		return def
	}
	t, err := time.ParseInLocation("2006-01-02", s, time.UTC)
	if err != nil {
		return def
	}
	return t
}

// AddDays offsets a day by n whole days. Days are UTC midnight, so this needs
// no daylight-saving handling.
func AddDays(t time.Time, n int) time.Time {
	return t.AddDate(0, 0, n)
}

// PreviousPrice records the price a day had before it was revised, and when we
// noticed the revision.
type PreviousPrice struct {
	DetectionTimestamp time.Time `json:"detectionTimestamp"`
	Price              float64   `json:"price"`
}

// Price is the price of one fuel type on one day.
type Price struct {
	FuelType   Type            `json:"fueltype"`
	Date       time.Time       `json:"date"`
	Price      float64         `json:"price"`
	PrevPrices []PreviousPrice `json:"prevPrices"`
}

// DayPrices holds the prices surrounding a requested day. Today is never nil
// when DayPrices is non-nil.
type DayPrices struct {
	Today     *Price `json:"today"`
	Yesterday *Price `json:"yesterday"`
	Tomorrow  *Price `json:"tomorrow"`
}
