package common

import (
	"fmt"
	"strconv"
	"time"
)

// ClickHouse 类型定义
const (
	CKTypeString   = "String"
	CKTypeInt64    = "Int64"
	CKTypeFloat64  = "Float64"
	CKTypeBool     = "Bool"
	CKTypeDateTime = "DateTime"
)

// DetectClickHouseType 从数据样本中推断ClickHouse类型
func DetectClickHouseType(value interface{}) string {
	if value == nil {
		return CKTypeString
	}

	switch v := value.(type) {
	case bool:
		return CKTypeBool
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return CKTypeInt64
	case float32, float64:
		return CKTypeFloat64
	case string:
		return detectStringType(v)
	case time.Time:
		return CKTypeDateTime
	default:
		return CKTypeString
	}
}

// detectStringType 尝试从字符串推断类型
func detectStringType(s string) string {
	if s == "" {
		return CKTypeString
	}

	// Bool
	if s == "true" || s == "false" || s == "True" || s == "False" {
		return CKTypeBool
	}

	// 时间戳（毫秒）
	if ts, err := strconv.ParseInt(s, 10, 64); err == nil {
		if ts > 946684800000 && ts < 4102444800000 {
			return CKTypeDateTime
		}
	}

	// Int64
	if _, err := strconv.ParseInt(s, 10, 64); err == nil {
		return CKTypeInt64
	}

	// Float64
	if _, err := strconv.ParseFloat(s, 64); err == nil {
		return CKTypeFloat64
	}

	// ISO时间格式
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if _, err := time.Parse(layout, s); err == nil {
			return CKTypeDateTime
		}
	}

	return CKTypeString
}

// DetectClickHouseTypeFromSamples 从多个样本中推断最合适的类型
func DetectClickHouseTypeFromSamples(samples []interface{}) string {
	if len(samples) == 0 {
		return CKTypeString
	}

	typeCount := make(map[string]int)
	for _, sample := range samples {
		t := DetectClickHouseType(sample)
		typeCount[t]++
	}

	// 如果所有样本类型一致
	if len(typeCount) == 1 {
		for t := range typeCount {
			return t
		}
	}

	// 类型优先级：String > DateTime > Float64 > Int64 > Bool
	if typeCount[CKTypeString] > 0 {
		return CKTypeString
	}
	if typeCount[CKTypeDateTime] > 0 {
		return CKTypeString
	}
	if typeCount[CKTypeFloat64] > 0 {
		return CKTypeFloat64
	}
	if typeCount[CKTypeInt64] > 0 {
		return CKTypeInt64
	}
	if typeCount[CKTypeBool] > 0 {
		return CKTypeBool
	}

	return CKTypeString
}

// ConvertToClickHouseValue 将值转换为ClickHouse期望的类型
func ConvertToClickHouseValue(value interface{}, targetType string) (interface{}, error) {
	if value == nil {
		return getDefaultValue(targetType), nil
	}

	switch targetType {
	case CKTypeBool:
		return convertToBool(value)
	case CKTypeInt64:
		return convertToInt64(value)
	case CKTypeFloat64:
		return convertToFloat64(value)
	case CKTypeDateTime:
		return convertToDateTime(value)
	case CKTypeString:
		return convertToString(value)
	default:
		return convertToString(value)
	}
}

func convertToBool(value interface{}) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		return strconv.ParseBool(v)
	case int, int8, int16, int32, int64:
		return v != 0, nil
	default:
		return false, fmt.Errorf("cannot convert %T to bool", value)
	}
}

func convertToInt64(value interface{}) (int64, error) {
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	case float64:
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to int64", value)
	}
}

func convertToFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case int, int64:
		i64, _ := convertToInt64(v)
		return float64(i64), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", value)
	}
}

func convertToDateTime(value interface{}) (time.Time, error) {
	switch v := value.(type) {
	case time.Time:
		return v, nil
	case string:
		layouts := []string{
			time.RFC3339,
			"2006-01-02T15:04:05Z07:00",
			"2006-01-02 15:04:05",
			"2006-01-02",
		}
		for _, layout := range layouts {
			if t, err := time.Parse(layout, v); err == nil {
				return t, nil
			}
		}
		if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
			return time.Unix(ts/1000, (ts%1000)*1e6), nil
		}
		return time.Time{}, fmt.Errorf("cannot parse string to datetime: %s", v)
	case int64:
		return time.Unix(v/1000, (v%1000)*1e6), nil
	default:
		return time.Time{}, fmt.Errorf("cannot convert %T to datetime", value)
	}
}

func convertToString(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	default:
		return fmt.Sprintf("%v", value), nil
	}
}

func getDefaultValue(ckType string) interface{} {
	switch ckType {
	case CKTypeBool:
		return false
	case CKTypeInt64:
		return int64(0)
	case CKTypeFloat64:
		return float64(0)
	case CKTypeDateTime:
		return time.Unix(0, 0)
	case CKTypeString:
		return ""
	default:
		return ""
	}
}

// GetClickHouseColumnDefinition 生成ClickHouse列定义
func GetClickHouseColumnDefinition(columnName, ckType string) string {
	switch ckType {
	case CKTypeString:
		return fmt.Sprintf("`%s` String DEFAULT ''", columnName)
	case CKTypeBool:
		return fmt.Sprintf("`%s` Nullable(Bool)", columnName)
	case CKTypeInt64:
		return fmt.Sprintf("`%s` Nullable(Int64)", columnName)
	case CKTypeFloat64:
		return fmt.Sprintf("`%s` Nullable(Float64)", columnName)
	case CKTypeDateTime:
		return fmt.Sprintf("`%s` Nullable(DateTime)", columnName)
	default:
		return fmt.Sprintf("`%s` String DEFAULT ''", columnName)
	}
}
