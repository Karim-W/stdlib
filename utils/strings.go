package utils

import (
	"errors"
	"strconv"
	"strings"
)

var ERR_UNSPORTED_TYPE = errors.New("unsupported type")

func StringBuilder(args ...any) (string, error) {
	builder := strings.Builder{}
	for _, arg := range args {
		switch arg.(type) {
		case string:
			builder.WriteString(arg.(string))
		case int:
			builder.WriteString(strconv.Itoa(arg.(int)))
		case int8:
			builder.WriteString(strconv.FormatInt(int64(arg.(int8)), 10))
		case int16:
			builder.WriteString(strconv.FormatInt(int64(arg.(int16)), 10))
		case int32:
			builder.WriteString(strconv.FormatInt(int64(arg.(int32)), 10))
		case int64:
			builder.WriteString(strconv.FormatInt(arg.(int64), 10))
		case uint:
			builder.WriteString(strconv.FormatUint(uint64(arg.(uint)), 10))
		case uint8:
			builder.WriteString(strconv.FormatUint(uint64(arg.(uint8)), 10))
		case uint16:
			builder.WriteString(strconv.FormatUint(uint64(arg.(uint16)), 10))
		case uint32:
			builder.WriteString(strconv.FormatUint(uint64(arg.(uint32)), 10))
		case uint64:
			builder.WriteString(strconv.FormatUint(arg.(uint64), 10))
		case float64:
			builder.WriteString(strconv.FormatFloat(arg.(float64), 'g', -1, 64))
		case float32:
			builder.WriteString(strconv.FormatFloat(float64(arg.(float32)), 'g', -1, 32))
		case bool:
			builder.WriteString(strconv.FormatBool(arg.(bool)))
		case []byte:
			builder.Write(arg.([]byte))
		case []rune:
			for i := 0; i < len(arg.([]rune)); i++ {
				builder.WriteRune(arg.([]rune)[i])
			}
		default:
			return "", ERR_UNSPORTED_TYPE
		}
	}
	return builder.String(), nil
}
