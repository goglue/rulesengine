package rulesengine

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ────────────────────────────────────────────────────────────────────────────
// toTime helper
// ────────────────────────────────────────────────────────────────────────────

func TestToTime(t *testing.T) {
	t.Run("time.Time input returned unchanged", func(t *testing.T) {
		in := time.Date(2023, 6, 15, 10, 0, 0, 0, time.UTC)
		out, err := toTime(in)
		require.NoError(t, err)
		assert.True(t, out.Equal(in))
	})

	t.Run("non-nil *time.Time returns dereferenced value", func(t *testing.T) {
		in := time.Date(2023, 6, 15, 10, 0, 0, 0, time.UTC)
		out, err := toTime(&in)
		require.NoError(t, err)
		assert.True(t, out.Equal(in))
	})

	t.Run("nil *time.Time returns error", func(t *testing.T) {
		var ptr *time.Time
		_, err := toTime(ptr)
		require.Error(t, err)
	})

	t.Run("RFC3339 string parses correctly", func(t *testing.T) {
		out, err := toTime("2023-06-15T10:00:00Z")
		require.NoError(t, err)
		assert.Equal(t, 2023, out.Year())
		assert.Equal(t, time.June, out.Month())
		assert.Equal(t, 15, out.Day())
		assert.Equal(t, 10, out.Hour())
	})

	t.Run("date-only string YYYY-MM-DD parses to midnight UTC", func(t *testing.T) {
		out, err := toTime("2023-06-15")
		require.NoError(t, err)
		assert.Equal(t, 2023, out.Year())
		assert.Equal(t, time.June, out.Month())
		assert.Equal(t, 15, out.Day())
		assert.Equal(t, 0, out.Hour())
		assert.Equal(t, 0, out.Minute())
		assert.Equal(t, 0, out.Second())
	})

	t.Run("invalid string returns error", func(t *testing.T) {
		_, err := toTime("not-a-date")
		require.Error(t, err)
	})

	t.Run("unsupported type int returns error", func(t *testing.T) {
		_, err := toTime(12345)
		require.Error(t, err)
	})

	t.Run("unsupported type bool returns error", func(t *testing.T) {
		_, err := toTime(true)
		require.Error(t, err)
	})

	t.Run("RFC3339Nano string parses correctly", func(t *testing.T) {
		out, err := toTime("2023-06-15T10:00:00.123456789Z")
		require.NoError(t, err)
		assert.Equal(t, 2023, out.Year())
		assert.Equal(t, time.June, out.Month())
	})
}

// ────────────────────────────────────────────────────────────────────────────
// parseRelativeTime helper
// ────────────────────────────────────────────────────────────────────────────

func TestParseRelativeTime(t *testing.T) {
	// Use a fixed reference "now" so tests are deterministic.
	ref := time.Date(2026, 3, 18, 15, 30, 0, 0, time.UTC)

	t.Run("now returns the reference time", func(t *testing.T) {
		out, err := parseRelativeTime("now", ref)
		require.NoError(t, err)
		assert.True(t, out.Equal(ref))
	})

	t.Run("today returns midnight on the reference date", func(t *testing.T) {
		out, err := parseRelativeTime("today", ref)
		require.NoError(t, err)
		assert.Equal(t, ref.Year(), out.Year())
		assert.Equal(t, ref.Month(), out.Month())
		assert.Equal(t, ref.Day(), out.Day())
		assert.Equal(t, 0, out.Hour())
		assert.Equal(t, 0, out.Minute())
	})

	t.Run("thisMonth returns first of current month at midnight", func(t *testing.T) {
		out, err := parseRelativeTime("thisMonth", ref)
		require.NoError(t, err)
		assert.Equal(t, ref.Year(), out.Year())
		assert.Equal(t, ref.Month(), out.Month())
		assert.Equal(t, 1, out.Day())
		assert.Equal(t, 0, out.Hour())
	})

	t.Run("thisYear returns Jan 1 of current year at midnight", func(t *testing.T) {
		out, err := parseRelativeTime("thisYear", ref)
		require.NoError(t, err)
		assert.Equal(t, ref.Year(), out.Year())
		assert.Equal(t, time.January, out.Month())
		assert.Equal(t, 1, out.Day())
		assert.Equal(t, 0, out.Hour())
	})

	t.Run("now-12mo is approximately 12 months before reference", func(t *testing.T) {
		out, err := parseRelativeTime("now-12mo", ref)
		require.NoError(t, err)
		expected := ref.AddDate(0, -12, 0)
		assert.True(t, out.Equal(expected), "got %v, want %v", out, expected)
	})

	t.Run("now+1y is approximately 1 year after reference", func(t *testing.T) {
		out, err := parseRelativeTime("now+1y", ref)
		require.NoError(t, err)
		expected := ref.AddDate(1, 0, 0)
		assert.True(t, out.Equal(expected))
	})

	t.Run("thisYear-1y is Jan 1 of previous year", func(t *testing.T) {
		out, err := parseRelativeTime("thisYear-1y", ref)
		require.NoError(t, err)
		assert.Equal(t, ref.Year()-1, out.Year())
		assert.Equal(t, time.January, out.Month())
		assert.Equal(t, 1, out.Day())
	})

	t.Run("thisYear+1y is Jan 1 of next year", func(t *testing.T) {
		out, err := parseRelativeTime("thisYear+1y", ref)
		require.NoError(t, err)
		assert.Equal(t, ref.Year()+1, out.Year())
		assert.Equal(t, time.January, out.Month())
		assert.Equal(t, 1, out.Day())
	})

	t.Run("now+30d is 30 days ahead", func(t *testing.T) {
		out, err := parseRelativeTime("now+30d", ref)
		require.NoError(t, err)
		expected := ref.AddDate(0, 0, 30)
		assert.True(t, out.Equal(expected))
	})

	t.Run("invalid string returns error", func(t *testing.T) {
		_, err := parseRelativeTime("not-relative", ref)
		require.Error(t, err)
	})

	t.Run("empty string returns error", func(t *testing.T) {
		_, err := parseRelativeTime("", ref)
		require.Error(t, err)
	})

	t.Run("plain number returns error", func(t *testing.T) {
		_, err := parseRelativeTime("12345", ref)
		require.Error(t, err)
	})
}

// ────────────────────────────────────────────────────────────────────────────
// toFloat helper
// ────────────────────────────────────────────────────────────────────────────

func TestToFloat(t *testing.T) {
	t.Run("int converts correctly", func(t *testing.T) {
		out, err := toFloat(42)
		require.NoError(t, err)
		assert.Equal(t, float64(42), out)
	})

	t.Run("int64 converts correctly", func(t *testing.T) {
		out, err := toFloat(int64(100))
		require.NoError(t, err)
		assert.Equal(t, float64(100), out)
	})

	t.Run("float64 returns as-is", func(t *testing.T) {
		out, err := toFloat(float64(3.14))
		require.NoError(t, err)
		assert.Equal(t, 3.14, out)
	})

	t.Run("float32 converts correctly", func(t *testing.T) {
		out, err := toFloat(float32(2.5))
		require.NoError(t, err)
		assert.InDelta(t, 2.5, out, 1e-6)
	})

	t.Run("uint32 converts correctly", func(t *testing.T) {
		out, err := toFloat(uint32(7))
		require.NoError(t, err)
		assert.Equal(t, float64(7), out)
	})

	t.Run("string numeric converts correctly", func(t *testing.T) {
		out, err := toFloat("42.5")
		require.NoError(t, err)
		assert.Equal(t, 42.5, out)
	})

	t.Run("invalid string returns error", func(t *testing.T) {
		_, err := toFloat("not-a-number")
		require.Error(t, err)
	})

	t.Run("nil *int returns error", func(t *testing.T) {
		var p *int
		_, err := toFloat(p)
		require.Error(t, err)
	})

	t.Run("non-nil *int converts correctly", func(t *testing.T) {
		v := 55
		out, err := toFloat(&v)
		require.NoError(t, err)
		assert.Equal(t, float64(55), out)
	})

	t.Run("unsupported type returns error", func(t *testing.T) {
		_, err := toFloat([]int{1, 2})
		require.Error(t, err)
	})

	t.Run("nil *string returns error", func(t *testing.T) {
		var p *string
		_, err := toFloat(p)
		require.Error(t, err)
	})
}

// ────────────────────────────────────────────────────────────────────────────
// toInterfaceSlice helper
// ────────────────────────────────────────────────────────────────────────────

func TestToInterfaceSlice(t *testing.T) {
	t.Run("[]any returned as-is", func(t *testing.T) {
		in := []any{"a", "b", "c"}
		out, ok := toInterfaceSlice(in)
		require.True(t, ok)
		assert.Equal(t, in, out)
	})

	t.Run("[]string converted to []any", func(t *testing.T) {
		out, ok := toInterfaceSlice([]string{"x", "y"})
		require.True(t, ok)
		require.Len(t, out, 2)
		assert.Equal(t, "x", out[0])
		assert.Equal(t, "y", out[1])
	})

	t.Run("[]int converted to []any", func(t *testing.T) {
		out, ok := toInterfaceSlice([]int{1, 2, 3})
		require.True(t, ok)
		require.Len(t, out, 3)
		assert.Equal(t, 1, out[0])
		assert.Equal(t, 3, out[2])
	})

	t.Run("[]float64 converted to []any", func(t *testing.T) {
		out, ok := toInterfaceSlice([]float64{1.1, 2.2})
		require.True(t, ok)
		require.Len(t, out, 2)
		assert.Equal(t, 1.1, out[0])
	})

	t.Run("non-slice returns false", func(t *testing.T) {
		_, ok := toInterfaceSlice("just-a-string")
		assert.False(t, ok)
	})

	t.Run("non-slice int returns false", func(t *testing.T) {
		_, ok := toInterfaceSlice(42)
		assert.False(t, ok)
	})

	t.Run("empty []any returns empty slice and true", func(t *testing.T) {
		out, ok := toInterfaceSlice([]any{})
		require.True(t, ok)
		assert.Empty(t, out)
	})

	t.Run("empty []string returns empty slice and true", func(t *testing.T) {
		out, ok := toInterfaceSlice([]string{})
		require.True(t, ok)
		assert.Empty(t, out)
	})

	t.Run("[]bool converted to []any", func(t *testing.T) {
		out, ok := toInterfaceSlice([]bool{true, false})
		require.True(t, ok)
		require.Len(t, out, 2)
		assert.Equal(t, true, out[0])
		assert.Equal(t, false, out[1])
	})

	t.Run("[]int64 converted to []any", func(t *testing.T) {
		out, ok := toInterfaceSlice([]int64{10, 20})
		require.True(t, ok)
		require.Len(t, out, 2)
		assert.Equal(t, int64(10), out[0])
	})

	t.Run("nil value returns false", func(t *testing.T) {
		_, ok := toInterfaceSlice(nil)
		assert.False(t, ok)
	})
}