package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	httptransport "github.com/go-kit/kit/transport/http"

	"github.com/gorilla/mux"
)

type UserInfoRequest struct {
	UserID int64 `json:"user_id"`
}

// func decodeValue(s string, kt reflect.Type) reflect.Value {
// 	var kv reflect.Value

// 	switch kt.Kind() {
// 	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
// 		n, err := strconv.ParseInt(s, 10, 64)
// 		if err != nil || reflect.Zero(kt).OverflowInt(n) {
// 			// d.saveError(&UnmarshalTypeError{Value: "number " + s, Type: kt, Offset: int64(start + 1)})
// 			// break
// 		}
// 		kv = reflect.ValueOf(n).Convert(kt)
// 		// v.SetInt(n)
// 	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
// 		n, err := strconv.ParseUint(s, 10, 64)
// 		if err != nil || reflect.Zero(kt).OverflowUint(n) {
// 			// d.saveError(&UnmarshalTypeError{Value: "number " + s, Type: kt, Offset: int64(start + 1)})
// 			// break
// 		}
// 		kv = reflect.ValueOf(n).Convert(kt)
// 	case reflect.String:
// 		kv = reflect.ValueOf(s).Convert(kt)
// 	default:
// 		panic("json: Unexpected key type") // should never occur
// 	}

// 	return kv
// }

// param
// query
// body

// omitempty
func MakeRequestDecoder(payloadMaker func() interface{}) httptransport.DecodeRequestFunc {
	return func(_ context.Context, r *http.Request) (request interface{}, err error) {
		payload := payloadMaker()

		typ := reflect.TypeOf(payload).Elem()
		value := reflect.ValueOf(payload).Elem()

		params := mux.Vars(r)
		query := r.URL.Query()
		fmt.Println(params)
		for i := 0; i < typ.NumField(); i++ {
			fieldType := typ.Field(i)
			fieldValue := value.Field(i)
			jsonTag := fieldType.Tag.Get("json")

			paramStringValue, ok := params[jsonTag]
			fmt.Println("paramStringValue", paramStringValue)
			if ok {
				LiteralStore([]byte(paramStringValue), fieldValue, false)
			}

			queryValue, ok := query[jsonTag]
			if ok {
				if len(queryValue) == 1 {
					LiteralStore([]byte(queryValue[0]), fieldValue, false)
				} else {
					// TODO
				}
			}

			err := json.NewDecoder(r.Body).Decode(payload)
			if err != nil {
				// TODO
			}
		}

		return payload, nil
	}
}

// type Expected interface {

// }
// func (x interface{}) kek() {

// }

type Any interface{}

type AnyValue struct {
	Any
}

func (av AnyValue) Dump() {
	fmt.Printf("%#v\n", av.Any)
}

func NewExpected(value interface{}) {

}

func main() {
	// makeRequestDecoder(func() interface{} {
	// 	return &UserInfoRequest{}
	// })
	AnyValue{42}.Dump()
}
