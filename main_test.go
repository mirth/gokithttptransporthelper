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
	CategoryName string `json:"category_name"`
	ProjectID    int    `json:"project_id"`
	Query1       int8   `json:"query1"`
	Query2       string `json:"query2"`
}

func makeTestProductRequestEndpoint(t *testing.T, gt productRequest) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (response interface{}, err error) {
		req := *request.(*productRequest)

		assert.Equal(t, req, gt)
		return nil, nil
	}
}

func TestMakeRequestDecoder(t *testing.T) {
	produtTestingEndpoint := makeTestProductRequestEndpoint(t, productRequest{
		CategoryName: "test_category",
		ProjectID:    123123,
		Query1:       124,
		Query2:       "q2",
	})
	productRequestDecoder := MakeRequestDecoder(func() interface{} {
		return &productRequest{}
	})

	testDecoder(
		"GET",
		"/products/{project_id}/{category_name}",
		[]string{
			"query1", "{query1}",
			"query2", "{query2}",
		},
		productRequestDecoder,
		produtTestingEndpoint,
		"http://example.com/products/123123/test_category?query1=124&query2=q2",
		[]byte{},
	)

	testDecoder(
		"POST",
		"/products",
		[]string{},
		productRequestDecoder,
		produtTestingEndpoint,
		"http://example.com/products",
		[]byte(
			`{
				"category_name": "test_category",
				"project_id": 123123,
				"query1": 124,
				"query2": "q2"
			}`),
	)

	// all types in all scopes
	// scope intersection
}
