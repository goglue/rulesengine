package rulesengine

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	durationRegex     = regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)(ns|us|µs|ms|s|m|h|d|w|mo|y)`)
	relativeTimeRegex = regexp.MustCompile(`(?i)^\s*(now|today|thisday|thismonth|thisyear)\s*(?:([+-])\s*(\d+)\s*([a-z]+)?)?\s*$`)
)

func compareEqual(a, b any) bool {
	return a == b
}

func toString(v any) string {
	switch s := v.(type) {
	case string:
		return s
	case *string:
		return *s
	case fmt.Stringer:
		return s.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func isNumeric(v any) bool {
	switch v.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float64, float32:
		return true
	case *int, *int8, *int16, *int32, *int64, *uint, *uint8, *uint16, *uint32, *uint64, *float64, *float32:
		return true
	}
	return false
}

func toFloat(v any) (float64, error) {
	switch val := v.(type) {
	case int:
		return float64(val), nil
	case int8:
		return float64(val), nil
	case int16:
		return float64(val), nil
	case int32:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case uint:
		return float64(val), nil
	case uint8:
		return float64(val), nil
	case uint16:
		return float64(val), nil
	case uint32:
		return float64(val), nil
	case uint64:
		return float64(val), nil
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case string:
		f, err := strconv.ParseFloat(val, 64)
		if nil != err {
			return 0, newError(errNumeric, val)
		}
		return f, nil
	case *int:
		if val == nil {
			return 0, fmt.Errorf("nil pointer")
		}
		return float64(*val), nil
	case *int8:
		return float64(*val), nil
	case *int16:
		return float64(*val), nil
	case *int32:
		return float64(*val), nil
	case *int64:
		return float64(*val), nil
	case *uint:
		return float64(*val), nil
	case *uint8:
		return float64(*val), nil
	case *uint16:
		return float64(*val), nil
	case *uint32:
		return float64(*val), nil
	case *uint64:
		return float64(*val), nil
	case *float32:
		return float64(*val), nil
	case *float64:
		return *val, nil
	case *string:
		if val == nil {
			return 0, fmt.Errorf("nil string pointer")
		}
		f, err := strconv.ParseFloat(*val, 64)
		if err != nil {
			return 0, newError(errNumeric, v)
		}
		return f, nil
	}

	return 0, newError(errNumeric, v)
}

func compareNumeric(a, b any, op Operator) (bool, error) {
	af, err := toFloat(a)
	if err != nil {
		return false, err
	}
	bf, err := toFloat(b)
	if err != nil {
		return false, err
	}
	switch op {
	case Gt:
		return af > bf, nil
	case Gte:
		return af >= bf, nil
	case Lt:
		return af < bf, nil
	case Lte:
		return af <= bf, nil
	}
	return false, nil
}

func inList(value any, list any) (bool, error) {
	l := reflect.ValueOf(list)
	if l.Kind() != reflect.Slice {
		return false, newError(errType, l.Kind())
	}
	for i := 0; i < l.Len(); i++ {
		if compareEqual(value, l.Index(i).Interface()) {
			return true, nil
		}
	}
	return false, nil
}

func anyInList(actual any, list any) (bool, error) {
	l := reflect.ValueOf(list)
	if l.Kind() != reflect.Slice {
		return false, newError(errType, l.Kind())
	}

	m := make(map[any]struct{}, l.Len())
	for i := 0; i < l.Len(); i++ {
		m[l.Index(i).Interface()] = struct{}{}
	}

	in := reflect.ValueOf(actual)
	if in.Kind() != reflect.Slice {
		return false, newError(errType, in.Kind())
	}

	for i := 0; i < in.Len(); i++ {
		if _, ok := m[in.Index(i).Interface()]; ok {
			return true, nil
		}
	}

	return false, nil
}

func isBetween(val, rangeVal any) (bool, error) {
	vals, ok := rangeVal.([]any)
	if !ok || len(vals) != 2 {
		return false, newError(errType, rangeVal)
	}
	v, err := toFloat(val)
	if err != nil {
		return false, err
	}
	min, err := toFloat(vals[0])
	if err != nil {
		return false, err
	}
	max, err := toFloat(vals[1])
	if err != nil {
		return false, err
	}
	return v >= min && v <= max, nil
}

func compareLength(val any, target any, op Operator) (bool, error) {
	length := 0
	switch v := val.(type) {
	case string:
		length = len(v)
	case []any:
		length = len(v)
	default:
		arr, ok := toInterfaceSlice(val)
		if !ok {
			return false, newError(errType, target)
		}
		length = len(arr)
	}
	floatVal, err := toFloat(target)
	if nil != err {
		return false, err
	}
	expected := int(floatVal)
	switch op {
	case LengthEq:
		return length == expected, nil
	case LengthGt:
		return length > expected, nil
	case LengthLt:
		return length < expected, nil
	}
	return false, nil
}

func compareTime(a, b any, op Operator) (bool, error) {
	at, aok := a.(time.Time)
	if !aok {
		return false, newError(errType, a)
	}
	bt, err := resolveExpectedTime(b, time.Now())
	if err != nil {
		return false, err
	}
	switch op {
	case Before:
		return at.Before(bt), nil
	case After:
		return at.After(bt), nil
	}
	return false, nil
}

func isTimeBetween(val any, rangeVal any) (bool, error) {
	v, ok := val.(time.Time)
	if !ok {
		return false, newError(errType, val)
	}
	now := time.Now()
	start, end, err := normalizeTimeRange(rangeVal, now)
	if err != nil {
		return false, err
	}
	return (v.After(start) || v.Equal(start)) &&
		(v.Before(end) || v.Equal(end)), nil
}

func isWithinTime(val any, duration any, op Operator) (bool, error) {
	t, ok := val.(time.Time)
	if !ok {
		return false, newError(errType, val)
	}
	durStr, ok := duration.(string)
	if !ok {
		return false, newError(errType, duration)
	}

	dur, err := parseFlexibleDuration(durStr)
	if err != nil {
		return false, newError(errType, durStr)
	}

	now := time.Now()
	switch op {
	case WithinLast:
		return t.After(now.Add(-dur)), nil
	case WithinNext:
		return t.Before(now.Add(dur)), nil
	}
	return false, nil
}

func resolveExpectedTime(expected any, now time.Time) (time.Time, error) {
	switch v := expected.(type) {
	case time.Time:
		return v, nil
	case *time.Time:
		if v == nil {
			return time.Time{}, newError(errType, expected)
		}
		return *v, nil
	case string:
		t, err := parseRelativeTime(v, now)
		if err != nil {
			return time.Time{}, newError(errType, expected)
		}
		return t, nil
	default:
		return time.Time{}, newError(errType, expected)
	}
}

func normalizeTimeRange(rangeVal any, now time.Time) (time.Time, time.Time, error) {
	switch r := rangeVal.(type) {
	case []time.Time:
		if len(r) != 2 {
			return time.Time{}, time.Time{}, newError(errType, rangeVal)
		}
		return r[0], r[1], nil
	case []any:
		if len(r) != 2 {
			return time.Time{}, time.Time{}, newError(errType, rangeVal)
		}
		start, err := resolveExpectedTime(r[0], now)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		end, err := resolveExpectedTime(r[1], now)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		return start, end, nil
	default:
		return time.Time{}, time.Time{}, newError(errType, rangeVal)
	}
}

func parseRelativeTime(input string, now time.Time) (time.Time, error) {
	match := relativeTimeRegex.FindStringSubmatch(input)
	if match == nil {
		return time.Time{}, fmt.Errorf("invalid relative time: %s", input)
	}

	base := strings.ToLower(match[1])
	sign := match[2]
	valueStr := match[3]
	unit := strings.ToLower(match[4])

	baseTime, err := relativeBaseTime(base, now)
	if err != nil {
		return time.Time{}, err
	}

	if valueStr == "" {
		if sign != "" || unit != "" {
			return time.Time{}, fmt.Errorf("invalid relative time: %s", input)
		}
		return baseTime, nil
	}

	if unit == "" {
		unit = defaultUnitForBase(base)
		if unit == "" {
			return time.Time{}, fmt.Errorf("missing unit for relative time: %s", input)
		}
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid relative time number: %s", valueStr)
	}
	if sign == "-" {
		value = -value
	}

	switch unit {
	case "y", "yr", "yrs", "year", "years":
		return baseTime.AddDate(value, 0, 0), nil
	case "mo", "mon", "month", "months":
		return baseTime.AddDate(0, value, 0), nil
	case "w", "week", "weeks":
		return baseTime.AddDate(0, 0, 7*value), nil
	case "d", "day", "days":
		return baseTime.AddDate(0, 0, value), nil
	case "h", "hr", "hrs", "hour", "hours":
		return baseTime.Add(time.Duration(value) * time.Hour), nil
	case "m", "min", "mins", "minute", "minutes":
		return baseTime.Add(time.Duration(value) * time.Minute), nil
	case "s", "sec", "secs", "second", "seconds":
		return baseTime.Add(time.Duration(value) * time.Second), nil
	case "ms", "millisecond", "milliseconds":
		return baseTime.Add(time.Duration(value) * time.Millisecond), nil
	case "us", "µs", "microsecond", "microseconds":
		return baseTime.Add(time.Duration(value) * time.Microsecond), nil
	case "ns", "nanosecond", "nanoseconds":
		return baseTime.Add(time.Duration(value) * time.Nanosecond), nil
	default:
		return time.Time{}, fmt.Errorf("unknown relative time unit: %s", unit)
	}
}

func relativeBaseTime(base string, now time.Time) (time.Time, error) {
	switch base {
	case "now":
		return now, nil
	case "today", "thisday":
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()), nil
	case "thismonth":
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()), nil
	case "thisyear":
		return time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, now.Location()), nil
	default:
		return time.Time{}, fmt.Errorf("unknown relative time base: %s", base)
	}
}

func defaultUnitForBase(base string) string {
	switch base {
	case "thisyear":
		return "y"
	case "thismonth":
		return "mo"
	case "today", "thisday":
		return "d"
	default:
		return ""
	}
}

func toInterfaceSlice(input any) ([]any, bool) {
	switch v := input.(type) {
	case []any:
		return v, true
	case []string:
		out := make([]any, len(v))
		for i, val := range v {
			out[i] = val
		}
		return out, true
	case []bool:
		out := make([]any, len(v))
		for i, val := range v {
			out[i] = val
		}
		return out, true

	// Signed ints
	case []int:
		out := make([]any, len(v))
		for i, val := range v {
			out[i] = val
		}
		return out, true
	case []int8:
		out := make([]any, len(v))
		for i, val := range v {
			out[i] = val
		}
		return out, true
	case []int16:
		out := make([]any, len(v))
		for i, val := range v {
			out[i] = val
		}
		return out, true
	case []int32:
		out := make([]any, len(v))
		for i, val := range v {
			out[i] = val
		}
		return out, true
	case []int64:
		out := make([]any, len(v))
		for i, val := range v {
			out[i] = val
		}
		return out, true

	// Unsigned ints
	case []uint:
		out := make([]any, len(v))
		for i, val := range v {
			out[i] = val
		}
		return out, true
	case []uint8:
		out := make([]any, len(v))
		for i, val := range v {
			out[i] = val
		}
		return out, true
	case []uint16:
		out := make([]any, len(v))
		for i, val := range v {
			out[i] = val
		}
		return out, true
	case []uint32:
		out := make([]any, len(v))
		for i, val := range v {
			out[i] = val
		}
		return out, true
	case []uint64:
		out := make([]any, len(v))
		for i, val := range v {
			out[i] = val
		}
		return out, true

	// Floating point
	case []float32:
		out := make([]any, len(v))
		for i, val := range v {
			out[i] = val
		}
		return out, true
	case []float64:
		out := make([]any, len(v))
		for i, val := range v {
			out[i] = val
		}
		return out, true

	default:
		// Fallback using reflect for custom types
		return toInterfaceSliceReflect(input)
	}
}

func toInterfaceSliceReflect(input any) ([]any, bool) {
	val := reflect.ValueOf(input)
	if val.Kind() != reflect.Slice {
		return nil, false
	}
	result := make([]any, val.Len())
	for i := 0; i < val.Len(); i++ {
		result[i] = val.Index(i).Interface()
	}
	return result, true
}

// parseFlexibleDuration parses durations like "5h", "2d", "3w", "1mo", "1.5y"
func parseFlexibleDuration(s string) (time.Duration, error) {
	matches := durationRegex.FindAllStringSubmatch(s, -1)
	if matches == nil {
		return 0, errors.New("invalid duration: " + s)
	}

	var total time.Duration
	for _, m := range matches {
		valueStr, unit := m[1], m[2]
		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid number %q: %v", valueStr, err)
		}

		switch unit {
		case "ns":
			total += time.Duration(value)
		case "us", "µs":
			total += time.Duration(value * float64(time.Microsecond))
		case "ms":
			total += time.Duration(value * float64(time.Millisecond))
		case "s":
			total += time.Duration(value * float64(time.Second))
		case "m":
			total += time.Duration(value * float64(time.Minute))
		case "h":
			total += time.Duration(value * float64(time.Hour))
		case "d":
			total += time.Duration(value * float64(24*time.Hour))
		case "w":
			total += time.Duration(value * float64(7*24*time.Hour))
		case "mo":
			// Approximate month = 30 days
			total += time.Duration(value * float64(30*24*time.Hour))
		case "y":
			// Approximate year = 365 days
			total += time.Duration(value * float64(365*24*time.Hour))
		default:
			return 0, fmt.Errorf("unknown unit: %s", unit)
		}
	}

	return total, nil
}
