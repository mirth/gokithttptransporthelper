package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/gorilla/mux"
)

// param
// query
// body

// omitempty
// fixed size array
// decoder.RegisterConverter(time.Time{}, timeConverter)

type Decoder struct {
}

func NewDecoder() *Decoder {
	return &Decoder{}
}

func (d *Decoder) Decode(r *http.Request, payload interface{}) error {
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
		// fmt.Println("bytes", string(bytes))
		err = json.Unmarshal(bytes, payload) //NewDecoder(r.Body).Decode(payload)
		if err != nil {
			panic(err.Error())
			// TODO
		}
	}

	return nil
}

func main() {

}
