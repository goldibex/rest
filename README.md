rest
====

A simple, powerful, convention-oriented REST API helper for go's net/http package.

Installation
------------

```go
import (
  "github.com/goldibex/rest"
)
```

Overview
--------

Package rest provides a simple, powerful, convention-oriented REST API helper for Go's net/http package.
The package itself handles nearly all the work associated with responding appropriately
to an HTTP/REST request, and returns an http.Handler (or Gorilla mux.Router) object ready for use.

Given a set of functions to handle REST methods ("GET", "POST", and the like),
rest invokes these functions as properly addressed HTTP requests come in. It returns
correct HTTP error codes in common conditions, such as the wrong method being
specified.

Most users will probably want to start with the associated package rest/jsonrest,
which provides a fast and easy starting point for a REST/JSON system.
