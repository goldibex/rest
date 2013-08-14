rest
====

A simple, powerful, convention-oriented REST API helper for go's net/http package.

The package itself handles nearly all the work associated with responding appropriately
to an HTTP/REST request, and returns an http.Handler (or Gorilla mux.Router) object ready for use.

Given a set of functions to handle REST methods ("GET", "POST", and the like),
rest invokes these functions as properly addressed HTTP requests come in. It returns
correct HTTP error codes in common conditions, such as the wrong method being
specified.

Most users will probably want to start with the associated package rest/jsonrest,
which provides a fast and easy starting point for a REST/JSON system.

Installation
------------

```go
import (
  "net/http"
  
  "github.com/goldibex/rest"
)
```

Then, before you compile:

```sh
go get github.com/goldibex/rest
```

Usage
-----

App Engine:

```go
func init() {
  e := rest.NewEndpoint("yams")
  http.Handle("/yams", e.Handler() )}
}
```

Mere mortals:

```go
func main() {
  e := rest.NewEndpoint("yams")
  http.Handle("/yams", e.Handler() )}

  http.ListenAndServe(":8998", nil)
}
```

These code fragments will set up a new REST endpoint at "/yams" that answers GET and POST methods for
the collection, and GET, HEAD, POST, PUT, and DELETE methods for objects (that's to say items requested
as /yams/{id}, like /yams/firstyam).

This code produces a working REST endpoint. Let's try to hit it:

```sh
$ curl -i localhost:8998/yams
HTTP/1.1 406 Not Acceptable
Content-Type: text/plain; charset=utf-8
Content-Length: 1
Date: Thu, 1 Jan 1970 00:00:01 GMT

```

Whoops! We didn't set the "Accept" header to something it liked, so we got the expected HTTP 406 "Not Acceptable" response. We can see that it sends the response Content-Type as "text/plain", so let's try using that:

```sh
$ curl -i -H "Accept: text/plain" localhost:8998/yams                             
HTTP/1.1 501 Not Implemented
Content-Type: text/plain
Content-Length: 2
Date: Thu, 1 Jan 1970 00:00:01 GMT
[]
```

That's better, but we still get 501 "Not Implemented." This is because we haven't actually set any handler functions yet. We can alter the server's behavior by setting the returned ```*rest.Endpoint```'s handlers like so:

```go
	e.GetCollection = func(r *http.Request, body []byte) ([]interface{}, error) {
		yams := []string{
			"YAMS",
			"YAMS",
			"YAMS",
		}
		return yams, nil
	}
```

And the new response:

```
$ curl -i -H "Accept: text/plain" localhost:8998/yams
HTTP/1.1 200 OK
Content-Type: text/plain
Content-Length: 16
Date: Thu, 1 Jan 1970 00:00:01 GMT

[YAMS YAMS YAMS]
```

Now for GET requests to the collection ("/yams"), the ```*rest.Endpoint``` will return http.StatusOK because
the associated handler doesn't return an error. Moreover, ```*rest.Endpoint``` takes the object we return, serializes it using the function specified in ```e.Codec.Marshal```, and attaches it to the response body.

The basic codec's ```Marshal``` function uses Go's built-in object serialization format and returns it as "text/plain", which is useful for debugging purposes. But we might like something else, say, JSON:

```go
	e.Codec.Accepts = "application/json"
	e.Codec.Marshal = json.Marshal // make sure you import "encoding/json"
```

Now let's rerun the request and ask for JSON:

```sh
$ curl -i -H "Accept: application/json" localhost:8998/yams
HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 22
Date: Thu, 1 Jan 1970 00:00:01 GMT

["YAMS","YAMS","YAMS"]
```

And we can see that we now get the object serialized as JSON instead.

(NB: for JSON, don't do this yourself. Use the primitives in ```github.com/goldibex/rest/jsonrest``` instead.)

Happy RESTing!

License
-------

It's the MIT license (read LICENSE for more details). Hack away. Pull requests welcome!
