package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
)

type UserRequest struct {
	UserID int `json:"user_id"`
}

type decodeRequestPayloadFunc = func(_ context.Context, r *http.Request) (request interface{}, err error)

type emptyResponseWriter struct {
	http.ResponseWriter
}

// Write([]byte) (int, error)
// WriteHeader(statusCode int)
func (emptyResponseWriter) Header() http.Header {
	return nil
}
func (emptyResponseWriter) Write([]byte) (int, error) {
	return 0, nil
}
func (emptyResponseWriter) WriteHeader(_ int) {
}

func testDecoder(path, method string,
	decodeRequestPayload decodeRequestPayloadFunc,
	endpoint endpoint.Endpoint,
	url string,
) {
	router := mux.NewRouter()

	req := httptest.NewRequest(method, url, nil)

	router.Methods(method).Path(path).Handler( //.Queries(queryPairs...)
		httptransport.NewServer(
			endpoint,
			decodeRequestPayload,
			func(_ context.Context, w http.ResponseWriter, response interface{}) error {
				return nil
			},
		))
	// fmt.Println("URL: ", req.URL)
	router.ServeHTTP(emptyResponseWriter{}, req)
	// w http.ResponseWriter, req *http.Request
}

func makeEndpoint(t *testing.T) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*UserRequest)
		t.Log(req.UserID)

		return nil, nil
	}
}

func TestMakeRequestDecoder(t *testing.T) {
	// body := ""
	fmt.Println("AZAZAZA")

	testDecoder(
		"/users/{user_id}",
		"GET",
		MakeRequestDecoder(func() interface{} {
			return &UserRequest{}
		}),
		makeEndpoint(t),
		"https://example.com/users/123",
	)
	//

}
