package rest

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
  "os"

	"github.com/gorilla/mux"
)

// Logger is an App Engine-compatible logging interface.
// You can assign an App Engine context object as a logger, as appengine.Context
// already implements rest.Logger.
type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warningf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Criticalf(format string, args ...interface{})
}

// IOLogger wraps any io.Writer to implement rest.Logger.
type IOLogger struct {
	io.Writer
}

// Debugf writes a debug-level log message to the underlying io.Writer.
func (i IOLogger) Debugf(format string, args ...interface{}) {
	fmt.Fprintf(i.Writer, "DEBUG: ")
	fmt.Fprintf(i.Writer, format, args...)
	fmt.Fprintf(i.Writer, "\n")
}

// Infof writes an info-level log message to the underlying io.Writer.
func (i IOLogger) Infof(format string, args ...interface{}) {
	fmt.Fprintf(i.Writer, "INFO: ")
	fmt.Fprintf(i.Writer, format, args...)
	fmt.Fprintf(i.Writer, "\n")

}

// Warningf writes a warning-level log message to the underlying io.Writer.
func (i IOLogger) Warningf(format string, args ...interface{}) {
	fmt.Fprintf(i.Writer, "WARNING: ")
	fmt.Fprintf(i.Writer, format, args...)
	fmt.Fprintf(i.Writer, "\n")
}

// Errorf writes an error-level log message to the underlying io.Writer.
func (i IOLogger) Errorf(format string, args ...interface{}) {
	fmt.Fprintf(i.Writer, "ERROR: ")
	fmt.Fprintf(i.Writer, format, args...)
	fmt.Fprintf(i.Writer, "\n")
}

// Criticalf writes a critical-level log message to the underlying io.Writer.
func (i IOLogger) Criticalf(format string, args ...interface{}) {
	fmt.Fprintf(i.Writer, "CRITICAL: ")
	fmt.Fprintf(i.Writer, format, args...)
	fmt.Fprintf(i.Writer, "\n")
}

// Codec describes a means of converting the request body to a valid Go struct
// and back again.
type Codec struct {
	// Accepts can be any valid MIME type, i.e. "application/yams."
	Accepts string
	// MaxSize specifies the upper limit on request body size. Any larger
	// request will be rejected with http.StatusRequestEntityTooLarge.
	MaxSize int64
	// Marshal is the function the package will call to encode a returned object
	// into a response body. For instance, jsonrest calls json.Marshal.
	Marshal func(v interface{}) ([]byte, error)
}

var (
	// ErrNotImplemented is the error that stub REST handlers always return. It
	// corresponds to http.StatusNotImplemented.
	ErrNotImplemented error = errors.New("Method not implemented")
  // ErrNotFound corresponds to http.StatusNotFound.
  ErrNotFound error = errors.New("Not found")
)

/*
type GetCollectionHandler func(r *http.Request, params url.Values) ([]interface{}, error)

func GetCollectionHandlerGen(name string) GetCollectionHandler {
  return func(r *http.Request, params url.Values) ([]interface{}, error) {
    // the basic handler has the following properties:
    // it will only get 10 values, up to a maximum of 100 via the count param.
    // it supports ordering on a single field via the order param.
    count, err := strconv.Atoi(params.Get("count")); if err != nil || count > 100 {
      count = 10
    }
    results := make([]interface{}, count)

    // run the datastore query
    _, err = datastore.NewQuery(name).
      Count(count).
      Order(params["order"]).
      GetAll(c, results)
    if err != nil {
      return nil, err
    }
    // return the results
    return results, nil
  }
}
*/

// CollectionHandler is the function signature for handlers of REST collections.
type CollectionHandler func(r *http.Request, body []byte) (interface{}, error)

// UnimplementedCollectionHandler is a stub function to fill out an Endpoint that
// only needs to implement certain REST methods on collections but not others. It always
// returns ErrNotImplemented.
func UnimplementedCollectionHandler(r *http.Request, body []byte) (interface{}, error) {
	return nil, ErrNotImplemented
}

// Handler is the function signature for handlers of REST objects.
type Handler func(r *http.Request, id string, body []byte) (interface{}, error)

// UnimplementedHandler is a stub function to fill out an Endpoint that
// only needs to implement certain REST methods on single objects but not others. It always
// returns ErrNotImplemented.
func UnimplementedHandler(r *http.Request, id string, body []byte) (interface{}, error) {
	return nil, ErrNotImplemented
}

// Endpoint is the description of a RESTful API that rest uses to handle HTTP requests.
type Endpoint struct {
	GetCollection  CollectionHandler
	PostCollection CollectionHandler

	Head   Handler
	Get    Handler
	Put    Handler
	Post   Handler
	Delete Handler

	Codec Codec
	// Name will be used to set the HTTP URL handlers for this REST object. For
	// instance, if Name is "yams", then Endpoint.Handler will return an http.Handler
	// that responds to "/yams" for collection actions and "/yams/{id}" for object actions.
  // the "id" URL parameter will be passed through to the relevant method handler,
  // so for instance a request to /yams/sweetpotato will have "sweetpotato" in 
  // the id argument.
	Name string
	// StatusCodeLookup maps error object to HTTP status codes.
	// NB: if rest can't look up an error an Endpoint returns in this map, it will
	// send http.StatusInternalServerError, except in the case of a true nil, for
	// which it will send http.StatusOK.
	StatusCodeLookup map[error]int
	// Logger is the Logger object rest uses to record events in the REST lifecycle.
	Logger Logger
  
  // If not nil, rest.Endpoint will call this method to get a logger for the
  // specific request rather than use the default logger. Useful for App Engine apps.
  RequestLogger func(r *http.Request) Logger
}

// NewEndpoint returns a initialized endpoint ready for use. Note that all requests
// will return 501 (Not Implemented) until proper handlers are set.
func NewEndpoint(name string) *Endpoint {
  return &Endpoint{
    GetCollection: UnimplementedCollectionHandler,
    PostCollection: UnimplementedCollectionHandler,

    Head: UnimplementedHandler,
    Get: UnimplementedHandler,
    Put: UnimplementedHandler,
    Post: UnimplementedHandler,
    Delete: UnimplementedHandler,

    Name: name,
    Codec: Codec {
      Accepts: "text/plain",
      MaxSize: 1<<10, // 1 megabyte
      Marshal: func(v interface{}) ([]byte, error) {
        return []byte(fmt.Sprintf("%+v", v)), nil
      },
    },
    StatusCodeLookup: map[error]int{
      ErrNotFound: http.StatusNotFound,
    },
    Logger: IOLogger{os.Stdout},
  }
}

func (e *Endpoint) handlerGen() func(w http.ResponseWriter, r *http.Request) {
  return func(w http.ResponseWriter, r *http.Request) {
		var (
			rv   interface{}
			data []byte
			err  error

      statusCode int
      ok bool
      log Logger
		)
    // get the logger
    if e.RequestLogger != nil {
      log = e.RequestLogger(r)
    } else {
      log = e.Logger
    }

		// we return the content type set by the codec
		w.Header().Set("Content-Type", e.Codec.Accepts)

		// recover the object id (mux stashes it away for us)
		id := mux.Vars(r)["id"]
		if id != "" {
			log.Debugf("id: %s", id)
		}

		// decode body phase
		// respect size limit
		if r.ContentLength > e.Codec.MaxSize {
			http.Error(w, "", http.StatusRequestEntityTooLarge)
			log.Errorf("Request body too large: max %d bytes, was %d", e.Codec.MaxSize, r.ContentLength)
			return
		}

		// slurp the data from the request
		if r.ContentLength > 0 {
			data, err = ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "", http.StatusInternalServerError)
				log.Errorf("Error reading request body: %s", err)
				return
			}
		}
		if id == "" { // collection
			switch r.Method {
			case "GET":
				rv, err = e.GetCollection(r, data)
			case "POST":
				rv, err = e.PostCollection(r, data)
			default:
				err = ErrNotImplemented
			}
		} else {
			// supported methods: HEAD, GET, POST, PUT, DELETE
			switch r.Method {
			case "HEAD":
				rv, err = e.Head(r, id, data)
			case "GET":
				rv, err = e.Get(r, id, data)
			case "POST":
				rv, err = e.Post(r, id, data)
			case "PUT":
				rv, err = e.Put(r, id, data)
			case "DELETE":
				rv, err = e.Delete(r, id, data)
			default:
				err = ErrNotImplemented
			}
		}

		// marshal the returned object
    data, marshalErr := e.Codec.Marshal(rv)
		if marshalErr != nil {
			http.Error(w, "", http.StatusInternalServerError)
			log.Errorf("Error marshaling return value: %s", marshalErr)
			return
		}

    // write the marshaled object to w
    switch(err) {
    case nil:
      statusCode = http.StatusOK
    case ErrNotImplemented:
      statusCode = http.StatusNotImplemented
    default:
			log.Errorf("Error returned during REST: id %s, method %s, error %s", id, r.Method, err)
			statusCode, ok = e.StatusCodeLookup[err]
			if !ok {
				statusCode = http.StatusInternalServerError
			}
		}
    w.Header().Set("X-Handled-By", "github.com/goldibex/rest")
		w.WriteHeader(statusCode)
    w.Write(data)
	}
}

func notAcceptableHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "", http.StatusNotAcceptable)
}

func notAllowedHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "", http.StatusMethodNotAllowed)
}

// Router builds a Gorilla router that will handle requests RESTfully using the endpoint.
// If the calling function passes nil for r, Router will create a new mux.Router.
func (e *Endpoint) Router(r *mux.Router) *mux.Router {
  if r == nil {
		r = mux.NewRouter()
	}
	eHandler := e.handlerGen()

	// collection path
	r.Path("/"+e.Name).
		Methods("GET", "POST").
		Headers("Accept", e.Codec.Accepts).
		HandlerFunc(eHandler)

	// collection path with wrong accept (triggers 406)
	r.Path("/"+e.Name).
		Methods("GET", "POST").
		HandlerFunc(notAcceptableHandler)

	// collection path with wrong method (triggers 405)
	r.Path("/"+e.Name).
		Methods("HEAD", "PUT", "DELETE", "OPTIONS", "TRACE", "CONNECT").
		HandlerFunc(notAllowedHandler)

	// object path
  r.Path("/"+e.Name+"/{id:[A-Za-z0-9-]+}").
		Methods("HEAD", "GET", "POST", "PUT", "DELETE").
		Headers("Accept", e.Codec.Accepts).
		HandlerFunc(eHandler)

	// object path with wrong accept (triggers 406)
  r.Path("/"+e.Name+"/{id:[A-Za-z0-9-]+}").
		Methods("HEAD", "GET", "POST", "PUT", "DELETE").
		HandlerFunc(notAcceptableHandler)

	// object path with wrong method (triggers 405)
  r.Path("/"+e.Name+"/{id:[A-Za-z0-9-]+}").
		Methods("OPTIONS", "TRACE", "CONNECT").
		HandlerFunc(notAllowedHandler)

	return r
}

// Handler creates an http.Handler ready for use with the http package, in case
// using the Gorilla mux package is undesirable.
func (e *Endpoint) Handler() http.Handler {
	return e.Router(nil)
}
