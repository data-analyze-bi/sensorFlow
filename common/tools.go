package common

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func GetUnixTime() int64 {
	return time.Now().Unix()
}

func Int64ToStr(id int64) string {
	return fmt.Sprintf("%d", id)
}

func If(condition bool, trueVal, falseVal interface{}) interface{} {
	if condition {
		return trueVal
	}
	return falseVal
}

func RemoveDollarSign(data map[string]interface{}) map[string]interface{} {
	newData := make(map[string]interface{})
	for key, value := range data {
		newKey := strings.TrimPrefix(key, "$")
		switch v := value.(type) {
		case map[string]interface{}:
			newData[newKey] = RemoveDollarSign(v)
		case []interface{}:
			newArray := make([]interface{}, len(v))
			for i, item := range v {
				if m, ok := item.(map[string]interface{}); ok {
					newArray[i] = RemoveDollarSign(m)
				} else {
					newArray[i] = item
				}
			}
			newData[newKey] = newArray
		default:
			newData[newKey] = value
		}
	}
	return newData
}

func FlattenJSON(data map[string]interface{}) map[string]interface{} {
	flatMap := make(map[string]interface{})
	for k, v := range data {
		if m, ok := v.(map[string]interface{}); ok {
			subFlat := FlattenJSON(m)
			for subk, subv := range subFlat {
				flatMap[subk] = subv
			}
			continue
		}
		flatMap[k] = v
	}
	return flatMap
}

func ValuesToStrings(m map[string]interface{}) (map[string]string, error) {
	result := make(map[string]string)

	for k, v := range m {
		if v == nil {
			result[k] = ""
			continue
		}
		switch value := v.(type) {
		case bool:
			result[k] = strconv.FormatBool(value)
		case int:
			result[k] = strconv.Itoa(value)
		case int8:
			result[k] = strconv.FormatInt(int64(value), 10)
		case int16:
			result[k] = strconv.FormatInt(int64(value), 10)
		case int32:
			result[k] = strconv.FormatInt(int64(value), 10)
		case int64:
			result[k] = strconv.FormatInt(value, 10)
		case uint:
			result[k] = strconv.FormatUint(uint64(value), 10)
		case uint8:
			result[k] = strconv.FormatUint(uint64(value), 10)
		case uint16:
			result[k] = strconv.FormatUint(uint64(value), 10)
		case uint32:
			result[k] = strconv.FormatUint(uint64(value), 10)
		case uint64:
			result[k] = strconv.FormatUint(value, 10)
		case float32:
			result[k] = strconv.FormatFloat(float64(value), 'f', -1, 32)
		case float64:
			result[k] = strconv.FormatFloat(value, 'f', -1, 64)
		case string:
			result[k] = value
		case time.Time:
			result[k] = value.Format(time.RFC3339)
		case map[string]interface{}, []interface{}:
			bytes, err := json.Marshal(value)
			if err != nil {
				return nil, fmt.Errorf("marshal complex type for key %q failed: %w", k, err)
			}
			result[k] = string(bytes)
		default:
			bytes, err := json.Marshal(value)
			if err != nil {
				return nil, fmt.Errorf("unsupported type for key %q: %T", k, value)
			}
			result[k] = string(bytes)
		}
	}

	return result, nil
}
