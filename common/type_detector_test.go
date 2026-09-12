package common

import (
	"testing"
	"time"
)

func TestDetectClickHouseType(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		// Bool类型
		{"bool true", true, CKTypeBool},
		{"bool false", false, CKTypeBool},
		{"string true", "true", CKTypeBool},
		{"string false", "false", CKTypeBool},

		// Int64类型
		{"int", 123, CKTypeInt64},
		{"int64", int64(123), CKTypeInt64},
		{"string int", "123", CKTypeInt64},
		{"negative int", -456, CKTypeInt64},

		// Float64类型
		{"float64", 123.45, CKTypeFloat64},
		{"string float", "123.45", CKTypeFloat64},
		{"negative float", -456.78, CKTypeFloat64},

		// DateTime类型
		{"time.Time", time.Now(), CKTypeDateTime},
		{"timestamp ms", "1609459200000", CKTypeDateTime},
		{"ISO8601", "2024-01-01T00:00:00Z", CKTypeDateTime},
		{"date string", "2024-01-01", CKTypeDateTime},

		// String类型
		{"empty string", "", CKTypeString},
		{"plain string", "hello", CKTypeString},
		{"nil value", nil, CKTypeString},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectClickHouseType(tt.value)
			if result != tt.expected {
				t.Errorf("DetectClickHouseType(%v) = %s, want %s", tt.value, result, tt.expected)
			}
		})
	}
}

func TestDetectClickHouseTypeFromSamples(t *testing.T) {
	tests := []struct {
		name     string
		samples  []interface{}
		expected string
	}{
		{
			name:     "all integers",
			samples:  []interface{}{1, 2, 3, 4, 5},
			expected: CKTypeInt64,
		},
		{
			name:     "all floats",
			samples:  []interface{}{1.1, 2.2, 3.3},
			expected: CKTypeFloat64,
		},
		{
			name:     "mixed int and float",
			samples:  []interface{}{1, 2.5, 3, 4.7},
			expected: CKTypeFloat64, // Float64可以包含Int64
		},
		{
			name:     "all bools",
			samples:  []interface{}{true, false, true},
			expected: CKTypeBool,
		},
		{
			name:     "mixed types with string",
			samples:  []interface{}{1, 2, "hello"},
			expected: CKTypeString, // 包含String则降级为String
		},
		{
			name:     "all strings",
			samples:  []interface{}{"a", "b", "c"},
			expected: CKTypeString,
		},
		{
			name:     "empty samples",
			samples:  []interface{}{},
			expected: CKTypeString,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectClickHouseTypeFromSamples(tt.samples)
			if result != tt.expected {
				t.Errorf("DetectClickHouseTypeFromSamples() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestConvertToClickHouseValue(t *testing.T) {
	tests := []struct {
		name       string
		value      interface{}
		targetType string
		wantErr    bool
	}{
		// Bool转换
		{"bool to bool", true, CKTypeBool, false},
		{"string to bool", "true", CKTypeBool, false},
		{"int to bool", 1, CKTypeBool, false},

		// Int64转换
		{"int to int64", 123, CKTypeInt64, false},
		{"string to int64", "123", CKTypeInt64, false},
		{"float to int64", 123.45, CKTypeInt64, false},

		// Float64转换
		{"float to float64", 123.45, CKTypeFloat64, false},
		{"string to float64", "123.45", CKTypeFloat64, false},
		{"int to float64", 123, CKTypeFloat64, false},

		// String转换
		{"any to string", 123, CKTypeString, false},
		{"nil to string", nil, CKTypeString, false},

		// 错误情况
		{"invalid string to int", "abc", CKTypeInt64, true},
		{"invalid string to float", "xyz", CKTypeFloat64, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ConvertToClickHouseValue(tt.value, tt.targetType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToClickHouseValue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetClickHouseColumnDefinition(t *testing.T) {
	tests := []struct {
		name       string
		columnName string
		ckType     string
		want       string
	}{
		{
			name:       "String column",
			columnName: "name",
			ckType:     CKTypeString,
			want:       "`name` String DEFAULT ''",
		},
		{
			name:       "Int64 column",
			columnName: "age",
			ckType:     CKTypeInt64,
			want:       "`age` Nullable(Int64)",
		},
		{
			name:       "Float64 column",
			columnName: "score",
			ckType:     CKTypeFloat64,
			want:       "`score` Nullable(Float64)",
		},
		{
			name:       "Bool column",
			columnName: "active",
			ckType:     CKTypeBool,
			want:       "`active` Nullable(Bool)",
		},
		{
			name:       "DateTime column",
			columnName: "created_at",
			ckType:     CKTypeDateTime,
			want:       "`created_at` Nullable(DateTime)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetClickHouseColumnDefinition(tt.columnName, tt.ckType)
			if result != tt.want {
				t.Errorf("GetClickHouseColumnDefinition() = %s, want %s", result, tt.want)
			}
		})
	}
}

func BenchmarkDetectClickHouseType(b *testing.B) {
	samples := []interface{}{
		123,
		"456",
		123.45,
		"true",
		"2024-01-01T00:00:00Z",
		"hello world",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, sample := range samples {
			DetectClickHouseType(sample)
		}
	}
}

func BenchmarkDetectClickHouseTypeFromSamples(b *testing.B) {
	samples := []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DetectClickHouseTypeFromSamples(samples)
	}
}
