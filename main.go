package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"
	"time"

	"github.com/gorilla/mux"
)

// param
// query
// body

// omitempty
// fixed size array
// decoder.RegisterConverter(time.Time{}, timeConverter)

type ConverterFunc func(string) reflect.Value

type Decoder struct {
	// custom handlers (e.g bool)
	customTypeConverters map[reflect.Type]ConverterFunc
}

func NewDecoder() *Decoder {
	return &Decoder{}
}

//registerconverter
func (d *Decoder) Decode(r *http.Request, payload interface{}) error {
	typ := reflect.TypeOf(payload).Elem()
	value := reflect.ValueOf(payload).Elem()

	params := mux.Vars(r)
	query := r.URL.Query()

	for i := 0; i < typ.NumField(); i++ {
		fieldType := typ.Field(i)
		fieldValue := value.Field(i)
		jsonTag := fieldType.Tag.Get("json")

		if jsonTag == "" {
			continue
		}

		if jsonTag[0] == '-' {
			continue
		}

		paramStringValue, ok := params[jsonTag]

		if ok {
			err := LiteralStore(paramStringValue, fieldValue)
			if err != nil {
				return err
			}
		}

		queryValue, ok := query[jsonTag]
		if ok {
			if len(queryValue) == 1 {
				err := LiteralStore(queryValue[0], fieldValue)
				if err != nil {
					return err
				}
			}
		}
	}

	bytes, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if len(bytes) > 0 {
		err = json.Unmarshal(bytes, payload)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Decoder) RegisterConverter(value interface{}, converterFunc ConverterFunc) {
	d.customTypeConverters[reflect.TypeOf(value)] = converterFunc
}

type KEK struct {
	T time.Time `json:"time"`
}

func main() {
}
