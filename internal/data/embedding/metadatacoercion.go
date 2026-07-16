package embedding

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

// equals checks if two values are equal
func equals(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}

// compare compares two values and returns:
// -1 if a < b
//
//	0 if a == b
//	1 if a > b
func compare(a, b interface{}) int {
	aVal := reflect.ValueOf(a)
	bVal := reflect.ValueOf(b)

	// Handle nil values
	if !aVal.IsValid() || !bVal.IsValid() {
		if !aVal.IsValid() && !bVal.IsValid() {
			return 0
		}
		if !aVal.IsValid() {
			return -1
		}
		return 1
	}

	// Handle different types
	aType := aVal.Type()
	bType := bVal.Type()

	// Try to convert to comparable types
	switch {
	case isNumeric(aType) && isNumeric(bType):
		// Convert both to float64 for comparison
		aFloat := Float64(a)
		bFloat := Float64(b)
		if aFloat < bFloat {
			return -1
		} else if aFloat > bFloat {
			return 1
		}
		return 0

	case isString(aType) && isString(bType):
		// Compare as strings
		aStr := toString(a)
		bStr := toString(b)
		return strings.Compare(aStr, bStr)

	case isTime(a) && isTime(b):
		// Compare as time.Time
		aTime := toTime(a)
		bTime := toTime(b)
		if aTime.Before(bTime) {
			return -1
		} else if aTime.After(bTime) {
			return 1
		}
		return 0

	default:
		// For incomparable types, compare string representations
		aStr := fmt.Sprintf("%v", a)
		bStr := fmt.Sprintf("%v", b)
		return strings.Compare(aStr, bStr)
	}
}

// contains checks if a contains b
func contains(a, b interface{}) bool {
	aStr := toString(a)
	bStr := toString(b)
	return strings.Contains(aStr, bStr)
}

// valueIn checks if a is in the collection b
func valueIn(a, b interface{}) bool {
	// If b is not a collection, compare directly
	bVal := reflect.ValueOf(b)
	if bVal.Kind() != reflect.Slice && bVal.Kind() != reflect.Array {
		return equals(a, b)
	}

	// Check if a is in the collection b
	for i := 0; i < bVal.Len(); i++ {
		if equals(a, bVal.Index(i).Interface()) {
			return true
		}
	}
	return false
}

// isNumeric checks if a type is numeric
func isNumeric(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

// isString checks if a type is a string
func isString(t reflect.Type) bool {
	return t.Kind() == reflect.String
}

// isTime checks if a value is a time.Time
func isTime(v interface{}) bool {
	_, ok := v.(time.Time)
	return ok
}

// Float64 converts a metadata value to float64, returning zero when conversion fails.
func Float64(v interface{}) float64 {
	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(val.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(val.Uint())
	case reflect.Float32, reflect.Float64:
		return val.Float()
	case reflect.String:
		var f float64
		_, err := fmt.Sscanf(val.String(), "%f", &f)
		if err != nil {
			// Log error and return 0 or handle accordingly
			return 0
		}
		return f
	default:
		return 0
	}
}

// toString converts a value to string
func toString(v interface{}) string {
	return fmt.Sprintf("%v", v)
}

// DefaultTimeFormats provides a list of common time formats for parsing
var DefaultTimeFormats = []string{
	time.RFC3339,
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006-01-02",
}

// toTime converts a value to time.Time
func toTime(v interface{}) time.Time {
	if t, ok := v.(time.Time); ok {
		return t
	}
	if s, ok := v.(string); ok {
		// Try to parse common time formats
		for _, format := range DefaultTimeFormats {
			if t, err := time.Parse(format, s); err == nil {
				return t
			}
		}
	}
	return time.Time{}
}
