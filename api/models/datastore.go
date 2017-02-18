package models

import (
	"context"
	"errors"
)

//TODO what assumptions can be made about appName and routePath?
// - are they sanitized/validated?
// - leading slash?
// - ./../ cleaned up?
// - no illegal chars?
// OR is it a public api as well...
//
type Datastore interface {
	// GetApp returns the app called appName or nil if it doesn't exist
	GetApp(ctx context.Context, appName string) (*App, error)
	GetApps(ctx context.Context, filter *AppFilter) ([]*App, error)
	InsertApp(ctx context.Context, app *App) (*App, error)
	//TODO are app.Config's fields merged with or do they replace existing fields
	UpdateApp(ctx context.Context, app *App) (*App, error)
	RemoveApp(ctx context.Context, appName string) error

	GetRoute(ctx context.Context, appName, routePath string) (*Route, error)
	GetRoutes(ctx context.Context, filter *RouteFilter) (routes []*Route, err error)
	GetRoutesByApp(ctx context.Context, appName string, filter *RouteFilter) (routes []*Route, err error)
	InsertRoute(ctx context.Context, route *Route) (*Route, error)
	//TODO are route's .Config and .Header fields merged with or do they replace existing fields
	UpdateRoute(ctx context.Context, route *Route) (*Route, error)
	RemoveRoute(ctx context.Context, appName, routePath string) error

	// The following provide a generic key value store for arbitrary data, can be used by extensions to store extra data
	// todo: should we namespace these by app? Then when an app is deleted, it can delete any of this extra data too.
	Put(context.Context, []byte, []byte) error
	Get(context.Context, []byte) ([]byte, error)
}

var (
	ErrDatastoreEmptyAppName   = errors.New("Missing app name")
	ErrDatastoreEmptyRoutePath = errors.New("Missing route name")
	ErrDatastoreEmptyApp       = errors.New("Missing app")
	ErrDatastoreEmptyRoute     = errors.New("Missing route")
	ErrDatastoreEmptyKey       = errors.New("Missing key")
)
