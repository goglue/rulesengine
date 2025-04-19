package rulesengine

import (
	"fmt"
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

func compareNumeric(a, b interface{}, op Operator) bool {
	af := toFloat(a)
	bf := toFloat(b)
	switch op {
	case Gt:
		return af > bf
	case Gte:
		return af >= bf
	case Lt:
		return af < bf
	case Lte:
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

func compareLength(val interface{}, target interface{}, op Operator) bool {
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
	case LengthEq:
		return length == expected
	case LengthGt:
		return length > expected
	case LengthLt:
		return length < expected
	}
	return false
}

func compareTime(a, b interface{}, op Operator) bool {
	at, aok := a.(time.Time)
	bt, bok := b.(time.Time)
	if !aok || !bok {
		return false
	}
	switch op {
	case Before:
		return at.Before(bt)
	case After:
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

func isWithinTime(val interface{}, duration interface{}, op Operator) bool {
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
	case WithinLast:
		return t.After(now.Add(-dur))
	case WithinNext:
		return t.Before(now.Add(dur))
	}
	return false
}
