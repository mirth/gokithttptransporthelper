package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"

	httptransport "github.com/go-kit/kit/transport/http"

	"github.com/gorilla/mux"
)

// param
// query
// body

// omitempty
// fixed size array
// decoder.RegisterConverter(time.Time{}, timeConverter)

func MakeRequestDecoder(payloadMaker func() interface{}) httptransport.DecodeRequestFunc {
	return func(_ context.Context, r *http.Request) (request interface{}, err error) {
		payload := payloadMaker()

		typ := reflect.TypeOf(payload).Elem()
		value := reflect.ValueOf(payload).Elem()

		params := mux.Vars(r)
		query := r.URL.Query()

		for i := 0; i < typ.NumField(); i++ {
			fieldType := typ.Field(i)
			fieldValue := value.Field(i)
			jsonTag := fieldType.Tag.Get("json")

			paramStringValue, ok := params[jsonTag]

			if ok {
				LiteralStore(paramStringValue, fieldValue)
			}

			queryValue, ok := query[jsonTag]
			if ok {
				if len(queryValue) == 1 {
					LiteralStore(queryValue[0], fieldValue)
				} else {
					// decodeQueryArray(fieldType, fieldValue, queryValue)
				}
			}
		}

		bytes, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if len(bytes) > 0 {
			err = json.Unmarshal(bytes, payload) //NewDecoder(r.Body).Decode(payload)
			if err != nil {
				panic(err.Error())
				// TODO
			}
		}

		return payload, nil
	}
}

// func decodeQueryArray(fieldType reflect.StructField, fieldValue reflect.Value, queryValue []string) {
// 	array := reflect.MakeSlice(fieldValue.Type(), 0, 0)
// 	for _, stringValue := range queryValue {
// 		value := reflect.New(fieldType.Type.Elem())
// 		LiteralStore(stringValue, value)
// 		array = reflect.Append(array, value)
// 	}
// 	fieldValue.Set(array)
// }

func main() {

}
