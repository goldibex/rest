package rest

import (
	"testing"

	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
)

func tryEndpoint(e *Endpoint, method, accept, url string) int {
	w := httptest.NewRecorder()
	handler := e.Handler()

	r, _ := http.NewRequest(method, url, nil)
	r.Header.Set("Accept", accept)
	handler.ServeHTTP(w, r)
	return w.Code
}

func newFalseEndpoint(name string) *Endpoint {
	return &Endpoint{
		GetCollection:  UnimplementedCollectionHandler,
		PostCollection: UnimplementedCollectionHandler,

		Get:    UnimplementedHandler,
		Head:   UnimplementedHandler,
		Put:    UnimplementedHandler,
		Post:   UnimplementedHandler,
		Delete: UnimplementedHandler,

		Codec: Codec{
			Accepts: "application/yams",
			MaxSize: 1 << 10,
			Marshal: func(v interface{}) ([]byte, error) {
				return []byte("YAMSYAMSYAMS"), nil
			},
		},
		Name: name,
		StatusCodeLookup: map[error]int{
			nil:               http.StatusOK,
			ErrNotImplemented: http.StatusNotImplemented,
		},
		Logger: IOLogger{os.Stdout},
	}
}

func TestCollectionDefaults(t *testing.T) {
	e := newFalseEndpoint("yams")
	// GET collection
	if statusCode := tryEndpoint(e, "GET", "application/yams",
		"http://example.com/yams"); statusCode != http.StatusNotImplemented {
		t.Errorf("Collection GET: Expected http return code %d, got %d",
			http.StatusNotImplemented, statusCode)
	}

	// POST collection
	if statusCode := tryEndpoint(e, "POST", "application/yams",
		"http://example.com/yams"); statusCode != http.StatusNotImplemented {
		t.Errorf("Collection POST: Expected http return code %d, got %d",
			http.StatusNotImplemented, statusCode)
	}

	// POST collection, but with wrong accept header
	if statusCode := tryEndpoint(e, "POST", "application/xml",
		"http://example.com/yams"); statusCode != http.StatusNotAcceptable {
		t.Errorf("Collection POST, wrong Accept: Expected http return code %d, got %d",
			http.StatusNotAcceptable, statusCode)
	}

	// PUT collection (wrong method)
	if statusCode := tryEndpoint(e, "PUT", "application/yams",
		"http://example.com/yams"); statusCode != http.StatusMethodNotAllowed {
		t.Errorf("Collection PUT: Expected http return code %d, got %d",
			http.StatusMethodNotAllowed, statusCode)
	}

}

func TestItemDefaults(t *testing.T) {
	e := newFalseEndpoint("yams")

	// GET, HEAD, POST, PUT, DELETE item
	for _, method := range []string{"GET", "HEAD", "POST", "PUT", "DELETE"} {
		if statusCode := tryEndpoint(e, method, "application/yams",
			"http://example.com/yams/1"); statusCode != http.StatusNotImplemented {
			t.Errorf("Item %s: Expected http return code %d, got %d",
				method, http.StatusNotImplemented, statusCode)
		}
	}

	// POST item (wrong accept header)
	if statusCode := tryEndpoint(e, "POST", "application/xml",
		"http://example.com/yams/1"); statusCode != http.StatusNotAcceptable {
		t.Errorf("Item POST, wrong Accept: Expected http return code %d, got %d",
			http.StatusNotAcceptable, statusCode)
	}

	// OPTIONS, TRACE, CONNECT item (wrong methods)
	for _, method := range []string{"OPTIONS", "TRACE", "CONNECT"} {
		if statusCode := tryEndpoint(e, method, "application/yams",
			"http://example.com/yams/1"); statusCode != http.StatusMethodNotAllowed {
			t.Errorf("Item %s: Expected http return code %d, got %d",
				method, http.StatusMethodNotAllowed, statusCode)
		}
	}
}

func TestSizeLimit(t *testing.T) {
	e := newFalseEndpoint("yams")

	// a megabyte plus a byte of random data is needed for this exercise
	// we'll just take that from /dev/random
	devRandom, err := os.Open("/dev/random")
	if err != nil {
		t.Fatalf("Couldn't open /dev/random: %s", err)
	}
	data := make([]byte, (1<<10)+1)
	n, err := devRandom.Read(data)
	if n < (1<<10)+1 || err != nil {
		t.Fatalf("Error reading data from /dev/random for test: got %d bytes with error %s", err)
	}

	handler := e.Handler()
	w := httptest.NewRecorder()

	r, _ := http.NewRequest("POST", "http://example.com/yams/1", bytes.NewBuffer(data))
	r.Header.Set("Accept", "application/yams")
	r.Header.Set("Content-Type", "application/yams")
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Size limit test: expected http return code %d, got %d",
			http.StatusRequestEntityTooLarge, w.Code)
	}
}
