package fuel

import "testing"

func TestParseType(t *testing.T) {
	for in, want := range map[string]Type{
		"diesel": Diesel, "Diesel": Diesel, "DIESEL": Diesel,
		"octane100": Octane100, "Octane100": Octane100,
		"unleaded95": Unleaded95, "": Unleaded95, "petrol": Unleaded95,
	} {
		if got := ParseType(in); got != want {
			t.Errorf("ParseType(%q) = %q, want %q", in, got, want)
		}
	}
}

// The integer mapping is the on-disk encoding of the fueltype column and must
// not drift.
func TestTypeIntRoundTrip(t *testing.T) {
	want := map[Type]int{Unleaded95: 1, Octane100: 2, Diesel: 3}
	for fuelType, i := range want {
		if got := fuelType.Int(); got != i {
			t.Errorf("%q.Int() = %d, want %d", fuelType, got, i)
		}
		got, err := TypeFromInt(i)
		if err != nil {
			t.Fatalf("TypeFromInt(%d): %v", i, err)
		}
		if got != fuelType {
			t.Errorf("TypeFromInt(%d) = %q, want %q", i, got, fuelType)
		}
	}
	if _, err := TypeFromInt(4); err == nil {
		t.Error("TypeFromInt(4) = nil error, want an error")
	}
}

func TestOkItemNumber(t *testing.T) {
	want := map[Type]int{Unleaded95: 536, Octane100: 533, Diesel: 231}
	for fuelType, n := range want {
		if got := fuelType.okItemNumber(); got != n {
			t.Errorf("%q.okItemNumber() = %d, want %d", fuelType, got, n)
		}
	}
}
