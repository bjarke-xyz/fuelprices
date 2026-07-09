package fuel

import (
	"fmt"
	"strings"
)

// Type is a fuel type. Its string form is the name used in the JSON API.
type Type string

const (
	Unleaded95 Type = "Unleaded95"
	Octane100  Type = "Octane100"
	Diesel     Type = "Diesel"
)

// AllTypes is the set of fuel types the job fetches, in fetch order.
var AllTypes = []Type{Diesel, Octane100, Unleaded95}

// ParseType maps a query-string value to a fuel type, defaulting to Unleaded95.
func ParseType(s string) Type {
	switch strings.ToLower(s) {
	case "octane100":
		return Octane100
	case "diesel":
		return Diesel
	default:
		return Unleaded95
	}
}

// Int is the fueltype column value.
func (t Type) Int() int {
	switch t {
	case Unleaded95:
		return 1
	case Octane100:
		return 2
	case Diesel:
		return 3
	}
	return 0
}

// TypeFromInt maps a fueltype column value back to a fuel type.
func TypeFromInt(i int) (Type, error) {
	switch i {
	case 1:
		return Unleaded95, nil
	case 2:
		return Octane100, nil
	case 3:
		return Diesel, nil
	}
	return "", fmt.Errorf("unknown integer fuel type '%d'", i)
}

// okItemNumber is the "varenr" the OK.dk price history endpoint expects.
func (t Type) okItemNumber() int {
	switch t {
	case Octane100:
		return 533
	case Diesel:
		return 231
	default:
		return 536
	}
}

func (t Type) localized(l Language) string {
	if l == Danish {
		switch t {
		case Octane100:
			return "Oktan 100"
		case Diesel:
			return "Diesel"
		default:
			return "Blyfri oktan 95"
		}
	}
	switch t {
	case Octane100:
		return "Octane 100"
	case Diesel:
		return "Diesel"
	default:
		return "Unleaed octane 95"
	}
}
