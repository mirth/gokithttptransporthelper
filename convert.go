package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

var numberType = reflect.TypeOf(json.Number(""))

const useNumber bool = false


func LiteralStore(s string, v reflect.Value) error {
	if len(s) == 0 {
		return fmt.Errorf("Empty string given for %v", v.Type())
	}

	{
		vAddr := v
		if v.Kind() != reflect.Ptr && v.Type().Name() != "" && v.CanAddr() {
			// haveAddr = true
			vAddr = v.Addr()
		}

		if vAddr.Type().NumMethod() > 0 && vAddr.CanInterface() {
			if u, ok := vAddr.Interface().(json.Unmarshaler); ok {
				return u.UnmarshalJSON([]byte(s))
			}
			// if !decodingNull {
			// 	if u, ok := v.Interface().(encoding.TextUnmarshaler); ok {
			// 		return nil, u, reflect.Value{}
			// 	}
			// }
		}
	}

	switch v.Kind() {
	case reflect.String:
		v.SetString(s)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil || v.OverflowInt(n) {
			return fmt.Errorf("Failed to parse [%s] in to %v", s, v.Type())
		}
		v.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		n, err := strconv.ParseUint(s, 10, 64)
		if err != nil || v.OverflowUint(n) {
			return fmt.Errorf("Failed to parse [%s] in to %v", s, v.Type())
		}
		v.SetUint(n)
	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(s, v.Type().Bits())
		if err != nil || v.OverflowFloat(n) {
			return fmt.Errorf("Failed to parse [%s] in to %v", s, v.Type())
		}
		v.SetFloat(n)
	case reflect.Bool:
		n, err := strconv.ParseBool(s)
		if err != nil {
			return fmt.Errorf("Failed to parse [%s] in to %v", s, v.Type())
		}
		v.SetBool(n)
	default:
		return fmt.Errorf("Unsupported type %v", v.Type())
	}

	return nil
}
