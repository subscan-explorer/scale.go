package types

func toInterfaceSlice[T any](values []T) []interface{} {
	out := make([]interface{}, len(values))
	for i, v := range values {
		out[i] = v
	}
	return out
}

func asInterfaceSlice(value interface{}) ([]interface{}, bool) {
	switch v := value.(type) {
	case []interface{}:
		return v, true
	case []string:
		return toInterfaceSlice(v), true
	case []int:
		return toInterfaceSlice(v), true
	case []int8:
		return toInterfaceSlice(v), true
	case []int16:
		return toInterfaceSlice(v), true
	case []int32:
		return toInterfaceSlice(v), true
	case []int64:
		return toInterfaceSlice(v), true
	case []uint:
		return toInterfaceSlice(v), true
	case []uint16:
		return toInterfaceSlice(v), true
	case []uint32:
		return toInterfaceSlice(v), true
	case []uint64:
		return toInterfaceSlice(v), true
	case []byte:
		return toInterfaceSlice(v), true
	case []bool:
		return toInterfaceSlice(v), true
	default:
		return nil, false
	}
}
