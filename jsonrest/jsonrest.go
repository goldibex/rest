// Package jsonrest provides a bootstrapped JSON-REST service implementation.
package jsonrest

import (
  "rest"
  "os"
  "encoding/json"
)

var (
  // JSONCodec is a REST codec set for JSON requests and responses.
  // By default it only allows request bodies of up to a megabyte.
  JSONCodec rest.Codec = rest.Codec{
    Accepts: "application/json",
    MaxSize: 1<<10, // 1 megabyte
    Marshal: json.Marshal,
  }
)

// NewJSONEndpoint returns a *rest.Endpoint configured to use JSON.
func NewJSONEndpoint(name string) *rest.Endpoint {
  return &rest.Endpoint{
    GetCollection: rest.UnimplementedCollectionHandler,
    PostCollection: rest.UnimplementedCollectionHandler,

    Get: rest.UnimplementedHandler,
    Head: rest.UnimplementedHandler,
    Put: rest.UnimplementedHandler,
    Post: rest.UnimplementedHandler,
    Delete: rest.UnimplementedHandler,

    Codec: JSONCodec,
    Name: name,
    StatusCodeLookup: map[error]int{},
    Logger: rest.IOLogger{os.Stdout},
  }
}
