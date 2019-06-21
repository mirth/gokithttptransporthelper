package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

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
) {
	router := mux.NewRouter()

	body := bytes.NewBuffer(bodyBytes)
	req := httptest.NewRequest(method, url, body)

	router.Methods(method).Path(path).Queries(queryPairs...).Handler(
		httptransport.NewServer(
			endpoint,
			decodeRequestPayload,
			func(_ context.Context, w http.ResponseWriter, response interface{}) error {
				return nil
			},
		))

	router.ServeHTTP(emptyResponseWriter{}, req)
}

type productRequest struct {
	CategoryName string   `json:"category_name"`
	ProjectID    int      `json:"project_id"`
	Query1       int8     `json:"query1"`
	Query2       string   `json:"query2"`
	Array1       []int32  `json:"array1"`
	Array2       []string `json:"array2"`
}

type typesRequest struct {
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
}

func makeTestProductRequestEndpoint(t *testing.T, gt productRequest) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (response interface{}, err error) {
		req := *request.(*productRequest)

		assert.Equal(t, req, gt)
		return nil, nil
	}
}

func makeTestTypesRequestEndpoint(t *testing.T, gt typesRequest) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (response interface{}, err error) {
		req := *request.(*typesRequest)

		assert.Equal(t, req, gt)
		return nil, nil
	}
}

func makeRequestDecoder(payloadMaker func() interface{}) httptransport.DecodeRequestFunc {
	decoder := NewDecoder()
	return func(_ context.Context, r *http.Request) (request interface{}, err error) {
		payload := payloadMaker()
		decoder.Decode(r, payload)
		return payload, nil
	}
}

func TestMakeRequestDecoder(t *testing.T) {
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

	typesRequestDecoder := makeRequestDecoder(func() interface{} {
		return &typesRequest{}
	})

	expectedTypesRequest := typesRequest{
		I8:   1,
		I16:  2,
		I32:  3,
		I64:  4,
		UI8:  5,
		UI16: 6,
		UI32: 7,
		UI64: 8,
		I:    9,
		UI:   10,
		F32:  0.11,
		F64:  0.12,
		STR:  "thirteen",
	}

	{
		testDecoder(
			"GET",
			"/{i8}/{i16}/{i32}/{i64}/{ui8}/{ui16}/{ui32}/{ui64}/{i}/{ui}/{f32}/{f64}/{str}",
			[]string{},
			typesRequestDecoder,
			makeTestTypesRequestEndpoint(t, expectedTypesRequest),
			"http://example.com/1/2/3/4/5/6/7/8/9/10/0.11/0.12/thirteen",
			[]byte("{}"),
		)
	}

	{
		testDecoder(
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
			typesRequestDecoder,
			makeTestTypesRequestEndpoint(t, expectedTypesRequest),
			"http://example.com/?i8=1&i16=2&i32=3&i64=4&ui8=5&ui16=6&ui32=7&ui64=8&i=9&ui=10&f32=0.11&f64=0.12&str=thirteen",
			[]byte("{}"),
		)
	}

	{
		testDecoder(
			"GET",
			"/",
			[]string{},
			typesRequestDecoder,
			makeTestTypesRequestEndpoint(t, expectedTypesRequest),
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
				"str": "thirteen"
			}`),
		)
	}
	// all types in all scopes
	// overflows/forbidden/empty
	// scope intersection
}
