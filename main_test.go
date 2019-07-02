package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
)

type decodeRequestPayloadFunc = func(_ context.Context, r *http.Request) (request interface{}, err error)

type emptyResponseWriter struct {
	http.ResponseWriter
}

func (emptyResponseWriter) Header() http.Header {
	return nil
}
func (emptyResponseWriter) Write([]byte) (int, error) {
	return 0, nil
}
func (emptyResponseWriter) WriteHeader(_ int) {
}

func testDecoder(
	method,
	path string,
	queryPairs []string,
	decodeRequestPayload decodeRequestPayloadFunc,
	endpoint endpoint.Endpoint,
	url string,
	bodyBytes []byte,
) *httptest.ResponseRecorder {
	router := mux.NewRouter()

	req := httptest.NewRequest(method, url, bytes.NewBuffer(bodyBytes))

	responseRecorder := httptest.NewRecorder()

	router.Methods(method).Path(path).Queries(queryPairs...).Handler(
		httptransport.NewServer(
			endpoint,
			decodeRequestPayload,
			func(_ context.Context, w http.ResponseWriter, response interface{}) error {
				return nil
			},
		))

	router.ServeHTTP(responseRecorder, req)

	return responseRecorder
}

type productRequest struct {
	CategoryName string   `json:"category_name"`
	ProjectID    int      `json:"project_id"`
	Query1       int8     `json:"query1"`
	Query2       string   `json:"query2"`
	Array1       []int32  `json:"array1"`
	Array2       []string `json:"array2"`
}

// TODO: primitive types
type BasicTypesRequest struct {
	I8   int8    `json:"i8"`
	I16  int16   `json:"i16"`
	I32  int32   `json:"i32"`
	I64  int64   `json:"i64"`
	UI8  uint8   `json:"ui8"`
	UI16 uint16  `json:"ui16"`
	UI32 uint32  `json:"ui32"`
	UI64 uint64  `json:"ui64"`
	I    int     `json:"i"`
	UI   uint    `json:"ui"`
	F32  float32 `json:"f32"`
	F64  float64 `json:"f64"`
	STR  string  `json:"str"`
	B    bool    `json:"b"`

	// NO_TAG string
	//

	// bool

	// PtrI8   *int8    `json:"ptr_i8"`
	// PtrI16  *int16   `json:"ptr_i16"`
	// PtrI32  *int32   `json:"ptr_i32"`
	// PtrI64  *int64   `json:"ptr_i64"`
	// PtrUI8  *uint8   `json:"ptr_ui8"`
	// PtrUI16 *uint16  `json:"ptr_ui16"`
	// PtrUI32 *uint32  `json:"ptr_ui32"`
	// PtrUI64 *uint64  `json:"ptr_ui64"`
	// PtrI    *int     `json:"ptr_i"`
	// PtrUI   *uint    `json:"ptr_ui"`
	// PtrF32  *float32 `json:"ptr_f32"`
	// PtrF64  *float64 `json:"ptr_f64"`
	// PtrSTR  *string  `json:"ptr_str"`
}

// do not test all variations of array and map because for body it implemented via standard json package
type BasicTypesRequestForBody struct {
	BasicTypesRequest

	Array []int          `json:"array"`
	Map   map[string]int `json:"map"`
}

func makeTestProductRequestEndpoint(t *testing.T, expected productRequest) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (response interface{}, err error) {
		actual := *request.(*productRequest)

		assert.Equal(t, expected, actual)
		return nil, nil
	}
}

func makeTestBasicTypesRequestEndpoint(t *testing.T, expected BasicTypesRequest) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (response interface{}, err error) {
		actual := *request.(*BasicTypesRequest)
		assert.Equal(t, expected, actual)
		return nil, nil
	}
}

func makeTestBodyBasicTypesRequestEndpoint(t *testing.T, expected BasicTypesRequestForBody) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (response interface{}, err error) {
		actual := *request.(*BasicTypesRequestForBody)

		assert.Equal(t, expected, actual)
		return nil, nil
	}
}

func makeRequestDecoder(payloadMaker func() interface{}) httptransport.DecodeRequestFunc {
	decoder := NewDecoder()
	return func(_ context.Context, r *http.Request) (request interface{}, err error) {
		payload := payloadMaker()
		err = decoder.Decode(r, payload)
		return payload, err
	}
}

func TestSimpleRequest(t *testing.T) {
	productRequestDecoder := makeRequestDecoder(func() interface{} {
		return &productRequest{}
	})

	{
		product := productRequest{
			CategoryName: "test_category",
			ProjectID:    123123,
			Query1:       124,
			Query2:       "q2",
		}

		testDecoder(
			"GET",
			"/products/{project_id}/{category_name}",
			[]string{
				"query1", "{query1}",
				"query2", "{query2}",
			},
			productRequestDecoder,
			makeTestProductRequestEndpoint(t, product),
			"http://example.com/products/123123/test_category?query1=124&query2=q2",
			[]byte("{}"),
		)
	}

	{
		product := productRequest{
			CategoryName: "test_category",
			ProjectID:    123123,
			Query1:       124,
			Query2:       "q2",
			Array1:       []int32{1, 2, 333},
			Array2:       []string{"a", "bb", "c"},
		}

		testDecoder(
			"POST",
			"/products",
			[]string{},
			productRequestDecoder,
			makeTestProductRequestEndpoint(t, product),
			"http://example.com/products",
			[]byte(
				`{
				"category_name": "test_category",
				"project_id": 123123,
				"query1": 124,
				"query2": "q2",
				"array1": [1, 2, 333],
				"array2": ["a", "bb", "c"]
			}`),
		)
	}
}

var BasicTypesRequestDecoder = makeRequestDecoder(func() interface{} {
	return &BasicTypesRequest{}
})

var bodyBasicTypesRequestDecoder = makeRequestDecoder(func() interface{} {
	return &BasicTypesRequestForBody{}
})

func TestBasicTypes(t *testing.T) {
	expectedBasicTypesRequest := BasicTypesRequest{
		I8:   int8(1),
		I16:  int16(2),
		I32:  int32(3),
		I64:  int64(4),
		UI8:  uint8(5),
		UI16: uint16(6),
		UI32: uint32(7),
		UI64: uint64(8),
		I:    int(9),
		UI:   uint(10),
		F32:  float32(0.11),
		F64:  float64(0.12),
		STR:  string("thirteen"),
		B:    true,
	}

	{
		testDecoder(
			"GET",
			"/{i8}/{i16}/{i32}/{i64}/{ui8}/{ui16}/{ui32}/{ui64}/{i}/{ui}/{f32}/{f64}/{str}/{ptr_i8}/{b}",
			[]string{},
			BasicTypesRequestDecoder,
			makeTestBasicTypesRequestEndpoint(t, expectedBasicTypesRequest),
			"http://example.com/1/2/3/4/5/6/7/8/9/10/0.11/0.12/thirteen/14/1",
			[]byte("{}"),
		)
	}

	{
		responseRecorder := testDecoder(
			"GET",
			"/",
			[]string{
				"i8", "{i8}",
				"i16", "{i16}",
				"i32", "{i32}",
				"i64", "{i64}",
				"ui8", "{ui8}",
				"ui16", "{ui16}",
				"ui32", "{ui32}",
				"ui64", "{ui64}",
				"i", "{i}",
				"ui", "{ui}",
				"f32", "{f32}",
				"f64", "{f64}",
				"str", "{str}",
			},
			BasicTypesRequestDecoder,
			makeTestBasicTypesRequestEndpoint(t, expectedBasicTypesRequest),
			"http://example.com/?i8=1&i16=2&i32=3&i64=4&ui8=5&ui16=6&ui32=7&ui64=8&i=9&ui=10&f32=0.11&f64=0.12&str=thirteen&b=1",
			[]byte("{}"),
		)

		assert.Equal(t, responseRecorder.Code, 200)
	}

	{
		expectedBodyBasicTypesRequest := BasicTypesRequestForBody{
			expectedBasicTypesRequest,
			[]int{1, 2, 3},
			map[string]int{"one": 4, "two": 5, "three": 6},
		}
		// fmt.Println("expectedBodyBasicTypesRequest", expectedBodyBasicTypesRequest)

		testDecoder(
			"GET",
			"/",
			[]string{},
			bodyBasicTypesRequestDecoder,
			makeTestBodyBasicTypesRequestEndpoint(t, expectedBodyBasicTypesRequest),
			"http://example.com/",
			[]byte(`{
				"i8": 1,
				"i16": 2,
				"i32": 3,
				"i64": 4,
				"ui8": 5,
				"ui16": 6,
				"ui32": 7,
				"ui64": 8,
				"i": 9,
				"ui": 10,
				"f32": 0.11,
				"f64": 0.12,
				"str": "thirteen",
				"b": true,
				"array": [1, 2, 3],
				"map": {
					"one": 4,
					"two": 5,
					"three": 6
				}
			}`),
		)
	}
}

func TestMalformedInputData(t *testing.T) {
	failNowEndpoint := func(_ context.Context, request interface{}) (response interface{}, err error) {
		assert.FailNow(t, "Must be not called")
		return nil, nil
	}

	// overflow
	{
		responseRecorder := testDecoder(
			"GET",
			"/{ui8}",
			[]string{},
			BasicTypesRequestDecoder,
			failNowEndpoint,
			"http://example.com/1111",
			[]byte("{}"),
		)
		assert.Equal(t, responseRecorder.Code, 500)
	}
	{
		responseRecorder := testDecoder(
			"GET",
			"/{i8}",
			[]string{},
			BasicTypesRequestDecoder,
			failNowEndpoint,
			"http://example.com/1111",
			[]byte("{}"),
		)
		assert.Equal(t, responseRecorder.Code, 500)
	}
	{
		responseRecorder := testDecoder(
			"GET",
			"/{f32}",
			[]string{},
			BasicTypesRequestDecoder,
			failNowEndpoint,
			"http://example.com/3.40282346638528859811704183484516925440e+39",
			[]byte("{}"),
		)
		assert.Equal(t, responseRecorder.Code, 500)
	}

	// failed to parse
	{
		responseRecorder := testDecoder(
			"GET",
			"/{i}",
			[]string{},
			BasicTypesRequestDecoder,
			failNowEndpoint,
			"http://example.com/kek",
			[]byte("{}"),
		)
		assert.Equal(t, responseRecorder.Code, 500)
	}
	{
		responseRecorder := testDecoder(
			"GET",
			"/{ui}",
			[]string{},
			BasicTypesRequestDecoder,
			failNowEndpoint,
			"http://example.com/kek",
			[]byte("{}"),
		)
		assert.Equal(t, responseRecorder.Code, 500)
	}
	{
		responseRecorder := testDecoder(
			"GET",
			"/{f64}",
			[]string{},
			BasicTypesRequestDecoder,
			failNowEndpoint,
			"http://example.com/kek",
			[]byte("{}"),
		)
		assert.Equal(t, responseRecorder.Code, 500)
	}
	{
		responseRecorder := testDecoder(
			"GET",
			"/",
			[]string{},
			BasicTypesRequestDecoder,
			failNowEndpoint,
			"http://example.com/",
			[]byte("{kek}"),
		)
		assert.Equal(t, responseRecorder.Code, 500)
	}

	{
		testDecoder(
			"GET",
			"/{i}",
			[]string{
				"i", "{i}",
			},
			BasicTypesRequestDecoder,
			failNowEndpoint,
			"http://example.com/kek",
			[]byte("{}"),
		)
		// assert.Equal(t, responseRecorder.Code, 500)
	}

	{
		testDecoder(
			"GET",
			"/{i}",
			[]string{
				"i", "{i}",
			},
			BasicTypesRequestDecoder,
			failNowEndpoint,
			"http://example.com/kek",
			[]byte(`{ "i": 42}`),
		)
		// assert.Equal(t, responseRecorder.Code, 500)
	}
}

func TestScopesIntersection(t *testing.T) {
	{
		responseRecorder := testDecoder(
			"GET",
			"/{f64}/{str}/{ui16}",
			[]string{
				"f64", "{f64}",
				"str", "{str}",
			},
			BasicTypesRequestDecoder,
			func(_ context.Context, request interface{}) (response interface{}, err error) {
				actual := *request.(*BasicTypesRequest)

				assert.Equal(t, BasicTypesRequest{
					F64: 0.2,
					STR: "kek",
					UI16: 123,
				}, actual)

				return nil, nil
			},
			"http://example.com/0.1/abc/123?f64=0.2&str=lol",
			[]byte(`{ "str": "kek" }`),
		)
		assert.Equal(t, responseRecorder.Code, 200)
	}
}

func TestTime(t *testing.T) {
	type WithTime struct {
		BODYTIME time.Time `json:"body_time"`
		VARTIME time.Time `json:"var_time"`
		QUERYTIME time.Time `json:"query_time"`
	}

	{
		responseRecorder := testDecoder(
			"GET",
			"/{var_time}",
			[]string{"query_time", "{query_time}"},
			makeRequestDecoder(func() interface{} {
				return &WithTime{}
			}),
			func(_ context.Context, request interface{}) (response interface{}, err error) {
				actual := *request.(*WithTime)

				assert.Equal(t, WithTime{
					BODYTIME: time.Unix(1562086576, 0).UTC(),
					VARTIME: time.Unix(1562086577, 0).UTC(),
					QUERYTIME: time.Unix(1562086578, 0).UTC(),
				}, actual)

				return nil, nil
			},
			`http://example.com/"2019-07-02T16:56:17Z"?query_time="2019-07-02T16:56:18Z"`,
			[]byte(`{ "body_time": "2019-07-02T16:56:16Z" }`),
		)
		assert.Equal(t, 200, responseRecorder.Code)
	}
}


type MyBoolean bool

func (b *MyBoolean) UnmarshalJSON(data []byte) error {
	if string(data) == "✓" {
		*b = MyBoolean(true)
		return nil
	}

	*b = MyBoolean(false)

	return nil
}

type WithCustomUnmarshal struct {
	FOO string
}

func (x *WithCustomUnmarshal) UnmarshalJSON(data []byte) error {
	x.FOO = string(data) + "kek"
	return nil
}

func TestCustomUnmarshallers(t *testing.T) {
	type Custom1 struct {
		MYB MyBoolean `json:"my_bool"`
		C WithCustomUnmarshal `json:"c"`
	}

	{
		responseRecorder := testDecoder(
			"GET",
			"/{c}",
			[]string{"my_bool", "{my_bool}"},
			makeRequestDecoder(func() interface{} {
				return &Custom1{}
			}),
			func(_ context.Context, request interface{}) (response interface{}, err error) {
				actual := *request.(*Custom1)

				assert.Equal(t, Custom1{
					MYB: MyBoolean(true),
					C: WithCustomUnmarshal{ FOO: "123kek" },
				}, actual)

				return nil, nil
			},
			`http://example.com/123?my_bool=✓`,
			[]byte(`{}`),
		)
		assert.Equal(t, 200, responseRecorder.Code)
	}
}


// func TestNotUseField(t *testing.T) {
// 	type withDashField struct {
// 		Dash int `json:"-"`
// 	}

// 	responseRecorder := testDecoder(
// 		"GET",
// 		"/{-}",
// 		[]string{
// 			"–", "{-}",
// 		},
// 		BasicTypesRequestDecoder,
// 		func(_ context.Context, request interface{}) (response interface{}, err error) {
// 			actual := *request.(*withDashField)

// 			assert.Equal(t, withDashField{
// 				Dash: 0,
// 			}, actual)

// 			return nil, nil
// 		},
// 		"http://example.com/-345",
// 		[]byte(`{ "-": 123 }`),
// 	)
// 	assert.Equal(t, responseRecorder.Code, 200)
// }

	// all types in all scopes +
	// overflows/forbidden/empty +

	// scope intersection +
	// omitempty +
	// embedded +
	// pointer to embedded
	// body null
	// struct as field
	// res := testDecoder
	// time.Time
	// string -> bytes
	// bytes
	// pointers to field
	// custom with "json" and without unmarshaller