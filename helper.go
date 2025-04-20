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

func toFloat(v interface{}) (float64, error) {
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
	}

	return 0, newError(errNumeric, v)
}

func compareNumeric(a, b interface{}, op Operator) (bool, error) {
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

func inList(value interface{}, list interface{}) (bool, error) {
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

func isBetween(val, rangeVal interface{}) (bool, error) {
	vals, ok := rangeVal.([]interface{})
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

func compareLength(val interface{}, target interface{}, op Operator) (bool, error) {
	length := 0
	switch v := val.(type) {
	case string:
		length = len(v)
	case []interface{}:
		length = len(v)
	default:
		return false, newError(errType, val)
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

func compareTime(a, b interface{}, op Operator) (bool, error) {
	at, aok := a.(time.Time)
	if !aok {
		return false, newError(errType, a)
	}
	bt, bok := b.(time.Time)
	if !bok {
		return false, newError(errType, b)
	}
	switch op {
	case Before:
		return at.Before(bt), nil
	case After:
		return at.After(bt), nil
	}
	return false, nil
}

func isTimeBetween(val interface{}, rangeVal interface{}) (bool, error) {
	v, ok := val.(time.Time)
	if !ok {
		return false, newError(errType, val)
	}
	r, ok := rangeVal.([]time.Time)
	if !ok || len(r) != 2 {
		return false, newError(errType, rangeVal)
	}
	return (v.After(r[0]) || v.Equal(r[0])) &&
		(v.Before(r[1]) || v.Equal(r[1])), nil
}

func isWithinTime(val interface{}, duration interface{}, op Operator) (bool, error) {
	t, ok := val.(time.Time)
	if !ok {
		return false, newError(errType, val)
	}
	dur, ok := duration.(time.Duration)
	if !ok {
		return false, newError(errType, duration)
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
