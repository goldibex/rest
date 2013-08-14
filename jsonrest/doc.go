/*
Package jsonrest provides a bootstrapped JSON-REST service implementation for Go's net/http package.

Handling REST requests over HTTP is as simple as the following:
  e := jsonrest.NewEndpoint("yams")
  http.Handle("/yams", e.Handler())

In this state, the handler will return 501s for all REST actions, as none have
been implemented. But adding one is as simple as the following:

  e.Get = func(r *http.Request, id string, body []byte) (interface{}, error) {
    // return some object based on the id.
    if object := fancydb.Lookup(id); object != nil {
      return object, nil
    }
    return nil, fancydb.ErrFancyDBIsBusted
  })

The first return parameter will be serialized using json.Marshal and returned as
the response body. The error will be logged, but more importantly it determines
the HTTP response code. Unknown error objects will be sent on the response as
error code 500 (Internal Server Error). You can customize this behavior by
mapping error codes on e.StatusCodeLookup as follows:
  
  e.StatusCodeLookup[fancydb.ErrFancyDBIsBusted] = http.StatusServiceUnavailable

*/
package jsonrest
