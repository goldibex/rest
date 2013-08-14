package jsonrest

import (
  "testing"

  "net/http"
  "net/http/httptest"
  "encoding/json"
  "bytes"
  "strings"
)

type testT struct {
  Yams string `json:"yams"`
  HasYams bool `json:"has_yams"`
  YamCount int `json:"yam_count"`
}

func TestNewJSONEndpoint(t *testing.T) {
  e := NewJSONEndpoint("yams")
  handler := e.Handler()

  // with valid Accept
  w := httptest.NewRecorder()
  r, _ := http.NewRequest("POST", "http://example.com/yams/1", nil)
  r.Header.Set("Accept", "application/json")
  r.Header.Set("Content-Type", "application/json")
  handler.ServeHTTP(w, r)

  if w.Code != http.StatusNotImplemented {
    t.Errorf("With valid Accept: expected http return code %d, got %d",
    http.StatusNotImplemented, w.Code)
  }

  // with invalid Accept
  w = httptest.NewRecorder()
  r, _ = http.NewRequest("POST", "http://example.com/yams/1", nil)
  r.Header.Set("Accept", "application/yams")
  r.Header.Set("Content-Type", "application/yams")
  handler.ServeHTTP(w, r)

  if w.Code != http.StatusNotAcceptable {
    t.Errorf("With invalid Accept: expected http return code %d, got %d",
    http.StatusNotAcceptable, w.Code)
  }

  // now prepare for an actual test
  e.Post = func(r *http.Request, id string, body []byte) (interface{}, error) {
    var yamObject testT
    var expectedYamObject testT = testT {
      Yams: "YAMSYAMSYAMS",
    }
    if err := json.Unmarshal(body, &yamObject); err != nil {
      t.Fatalf("testPostHandler: Expected no error unmarshaling JSON, got %s (body %s)", err, string(body))
    } else if yamObject != expectedYamObject {
      t.Errorf("testPostHandler: expected yamObject to be %+v, got %+v", expectedYamObject, yamObject)
    }

    // fiddle the yam object some
    yamObject.YamCount = strings.Count(yamObject.Yams, "YAMS")
    yamObject.HasYams = yamObject.YamCount > 0
    return yamObject, nil
  }


  // encode test object
  testRequestObject := testT{
    Yams: "YAMSYAMSYAMS",
  }
  data, err := json.Marshal(&testRequestObject); if err != nil {
    t.Fatalf("Error creating test request body via json.Marshal: %s", err)
  }

  // with an actual JSON request body via POST
  w = httptest.NewRecorder()
  t.Logf("data: %s", string(data))
  r, _ = http.NewRequest("POST", "http://example.com/yams/1", bytes.NewBuffer(data))
  r.Header.Set("Accept", "application/json")
  r.Header.Set("Content-Type", "application/json")
  handler.ServeHTTP(w, r)

  if w.Code != http.StatusOK {
    t.Errorf("With a real live JSON body: expected http return code %d, got %d",
    http.StatusOK, w.Code)
  }

  // we also expect a slightly mutated version of our request object
  var testResponseObject testT
  dec := json.NewDecoder(w.Body)
  if err := dec.Decode(&testResponseObject); err != nil {
    t.Fatalf("Error decoding JSON object from response body: %s", err)
  }
  if !testResponseObject.HasYams {
    t.Errorf("Expected testResponseObject.HasYams to be true, got false") 
  }
  if testResponseObject.YamCount != 3 {
    t.Errorf("Expected testResponseObject.YamCount to equal %d, got %d", 3, testResponseObject.YamCount)
  }
}
