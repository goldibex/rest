package jsonrest

import (
  "rest"
  "os"
  "encoding/json"
)

var (
  // jsonrest.Codec is a REST codec set up for JSON requests and responses.
  // By default it only allows request bodies of up to a megabyte.
  Codec rest.Codec = rest.Codec{
    Accepts: "application/json",
    MaxSize: 1<<10, // 1 megabyte
    Marshal: json.Marshal,
  }
)

// NewJSONEndpoint returns a *rest.Endpoint configured to use JSON.
func NewEndpoint(name string) *rest.Endpoint {
  return &rest.Endpoint{
    GetCollection: rest.UnimplementedCollectionHandler,
    PostCollection: rest.UnimplementedCollectionHandler,

    Get: rest.UnimplementedHandler,
    Head: rest.UnimplementedHandler,
    Put: rest.UnimplementedHandler,
    Post: rest.UnimplementedHandler,
    Delete: rest.UnimplementedHandler,

    Codec: Codec,
    Name: name,
    StatusCodeLookup: map[error]int{},
    Logger: rest.IOLogger{os.Stdout},
  }
}
