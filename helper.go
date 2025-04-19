package rulesengine

import (
	"fmt"
	"github.com/goglue/rulesengine/rules"
	"reflect"
	"strconv"
	"time"
)

func compareEqual(a, b interface{}) bool {
	return a == b
}

func toString(v interface{}) string {
	switch s := v.(type) {
	case string:
		return s
	case fmt.Stringer:
		return s.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func isNumeric(v interface{}) bool {
	switch v.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float64, float32:
		return true
	}
	return false
}

func toFloat(v interface{}) float64 {
	switch val := v.(type) {
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case float64:
		return val
	case float32:
		return float64(val)
	case string:
		f, _ := strconv.ParseFloat(val, 64)
		return f
	}
	return 0
}

func compareNumeric(a, b interface{}, op rules.Operator) bool {
	af := toFloat(a)
	bf := toFloat(b)
	switch op {
	case rules.Gt:
		return af > bf
	case rules.Gte:
		return af >= bf
	case rules.Lt:
		return af < bf
	case rules.Lte:
		return af <= bf
	}
	return false
}

func inList(value interface{}, list interface{}) bool {
	l := reflect.ValueOf(list)
	if l.Kind() != reflect.Slice {
		return false
	}
	for i := 0; i < l.Len(); i++ {
		if compareEqual(value, l.Index(i).Interface()) {
			return true
		}
	}
	return false
}

func isBetween(val, rangeVal interface{}) bool {
	vals, ok := rangeVal.([]interface{})
	if !ok || len(vals) != 2 {
		return false
	}
	v := toFloat(val)
	min := toFloat(vals[0])
	max := toFloat(vals[1])
	return v >= min && v <= max
}

func compareLength(val interface{}, target interface{}, op rules.Operator) bool {
	length := 0
	switch v := val.(type) {
	case string:
		length = len(v)
	case []interface{}:
		length = len(v)
	default:
		return false
	}
	expected := int(toFloat(target))
	switch op {
	case rules.LengthEq:
		return length == expected
	case rules.LengthGt:
		return length > expected
	case rules.LengthLt:
		return length < expected
	}
	return false
}

func compareTime(a, b interface{}, op rules.Operator) bool {
	at, aok := a.(time.Time)
	bt, bok := b.(time.Time)
	if !aok || !bok {
		return false
	}
	switch op {
	case rules.Before:
		return at.Before(bt)
	case rules.After:
		return at.After(bt)
	}
	return false
}

func isTimeBetween(val interface{}, rangeVal interface{}) bool {
	v, ok := val.(time.Time)
	if !ok {
		return false
	}
	r, ok := rangeVal.([]time.Time)
	if !ok || len(r) != 2 {
		return false
	}
	return (v.After(r[0]) || v.Equal(r[0])) && (v.Before(r[1]) || v.Equal(r[1]))
}

func isWithinTime(val interface{}, duration interface{}, op rules.Operator) bool {
	t, ok := val.(time.Time)
	if !ok {
		return false
	}
	dur, ok := duration.(time.Duration)
	if !ok {
		return false
	}

	now := time.Now()
	switch op {
	case rules.WithinLast:
		return t.After(now.Add(-dur))
	case rules.WithinNext:
		return t.Before(now.Add(dur))
	}
	return false
}
