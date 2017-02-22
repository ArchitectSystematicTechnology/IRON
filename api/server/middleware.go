// TODO: it would be nice to move these into the top level folder so people can use these with the "functions" package, eg: functions.AddMiddleware(...)
package server

import (
	"context"
	"net/http"
	"reflect"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	fcommon "github.com/iron-io/functions/api/common"
	"github.com/iron-io/functions/api/models"
)

// Middleware is the interface required for implementing functions middlewar
type Middleware interface {
	// Serve is what the Middleware must implement. Can modify the request, write output, etc.
	// todo: should we abstract the HTTP out of this?  In case we want to support other protocols.
	Serve(ctx MiddlewareContext, w http.ResponseWriter, r *http.Request, app *models.App) error
}

// MiddlewareFunc func form of Middleware
type MiddlewareFunc func(ctx MiddlewareContext, w http.ResponseWriter, r *http.Request, app *models.App) error

// Serve wrapper
func (f MiddlewareFunc) Serve(ctx MiddlewareContext, w http.ResponseWriter, r *http.Request, app *models.App) error {
	return f(ctx, w, r, app)
}

// MiddlewareContext extends context.Context for Middleware
type MiddlewareContext interface {
	context.Context

	// Middleware can call Next() explicitly to call the next middleware in the chain. If Next() is not called and an error is not returned, Next() will automatically be called.
	Next(ctx MiddlewareContext, w http.ResponseWriter, r *http.Request, app *models.App)
	// Index returns the index of where we're at in the chain
	Index() int
	// WithValue same behavior as context.WithValue, but returns MiddlewareContext
	WithValue(key, val interface{}) MiddlewareContext
	// Enables user to replace the context.Context instance, required if user calls context.WithValue so it can be replaced.
	SetContext(ctx context.Context)
}

type middlewareContextImpl struct {
	context.Context

	ginContext  *gin.Context
	nextCalled  bool
	index       int
	middlewares []Middleware
	app         *models.App
}

// WithValue is essentially the same as context.Context, but returns the MiddlewareContext
func (c *middlewareContextImpl) WithValue(key, val interface{}) MiddlewareContext {
	if key == nil {
		panic("nil key")
	}
	if !reflect.TypeOf(key).Comparable() {
		panic("key is not comparable")
	}
	ct2 := context.WithValue(c, key, val)
	mc2 := &middlewareContextImpl{Context: ct2, ginContext: c.ginContext, nextCalled: c.nextCalled, index: c.index, middlewares: c.middlewares}
	return mc2
}

func (c *middlewareContextImpl) Next(ctx MiddlewareContext, w http.ResponseWriter, r *http.Request, app *models.App) {
	c3, log := fcommon.LoggerWithStack(c, "Next")
	ctx.SetContext(c3)
	log.Infoln("Next called", ctx.Index())
	c2 := ctx.(*middlewareContextImpl)
	c2.app = app
	c2.nextCalled = true
	c2.index++
	c2.serveNext()
}
func (c *middlewareContextImpl) SetContext(ctx context.Context) {
	c.Context = ctx
}

func (c *middlewareContextImpl) serveNext() {
	c2, log := fcommon.LoggerWithStack(c, "serveNext")
	log.Infoln("serving middleware", c.Index())
	if c.Index() >= len(c.middlewares) {
		// pass onto gin when we're through functions middleware
		c.ginContext.Set("ctx", c)
		c.ginContext.Next()
		return
	}
	// make shallow copy:
	fctx2 := *c
	fctx2.Context = c2
	fctx2.nextCalled = false
	r := c.ginContext.Request.WithContext(fctx2)
	nextM := c.middlewares[c.Index()]
	err := nextM.Serve(&fctx2, c.ginContext.Writer, r, fctx2.app)
	if err != nil {
		logrus.WithError(err).Warnln("Middleware error")
		// todo: might be a good idea to check if anything has been written yet, and if not, output the error: simpleError(err)
		// see: http://stackoverflow.com/questions/39415827/golang-http-check-if-responsewriter-has-been-written
		c.ginContext.Abort()
		return
	}
	// this will be true if the user called Next() explicitly. If not, let's call it here.
	// if !fctx2.nextCalled {
	// 	// then we automatically call next
	// 	fctx2.Next(c, c.ginContext.Writer, r, fctx2.app)
	// }
}

func (c *middlewareContextImpl) Index() int {
	return c.index
}

// This is for Gin's middleware. Gin will call this and in turn, we'll call all the functions middleware.
func (s *Server) middlewareWrapperFunc(ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(s.middlewares) == 0 {
			return
		}
		ctx = c.MustGet("ctx").(context.Context)
		fctx := &middlewareContextImpl{Context: ctx}
		fctx.app = &models.App{}
		// fctx.index = -1
		fctx.ginContext = c
		fctx.middlewares = s.middlewares
		// start the chain:
		fctx.serveNext()
	}
}

// AddMiddleware adds middleware to all /v1/* routes
func (s *Server) AddMiddleware(m Middleware) {
	s.middlewares = append(s.middlewares, m)
}

// AddAppEndpoint adds middleware to all /v1/* routes
func (s *Server) AddMiddlewareFunc(m func(ctx MiddlewareContext, w http.ResponseWriter, r *http.Request, app *models.App) error) {
	s.AddMiddleware(MiddlewareFunc(m))
}

// AddRunMiddleware adds middleware to the user functions routes, not the API
func (s *Server) AddRunMiddleware(m Middleware) {
	s.runMiddlewares = append(s.runMiddlewares, m)
}

// AddRunMiddleware adds middleware to the user functions routes, not the API
func (s *Server) AddRunMiddlewareFunc(m func(ctx MiddlewareContext, w http.ResponseWriter, r *http.Request, app *models.App) error) {
	s.AddRunMiddleware(MiddlewareFunc(m))
}
