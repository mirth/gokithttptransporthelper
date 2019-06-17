package main

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"

	httptransport "github.com/go-kit/kit/transport/http"

	"github.com/gorilla/mux"
)

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

func main() {

}
