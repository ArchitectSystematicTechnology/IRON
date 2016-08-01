package models

import (
	"errors"
	"net/http"
	"path"

	apiErrors "github.com/go-openapi/errors"
)

var (
	ErrRoutesCreate     = errors.New("Could not create route")
	ErrRoutesUpdate     = errors.New("Could not update route")
	ErrRoutesRemoving   = errors.New("Could not remove route from datastore")
	ErrRoutesGet        = errors.New("Could not get route from datastore")
	ErrRoutesList       = errors.New("Could not list routes from datastore")
	ErrRoutesNotFound   = errors.New("Route not found")
	ErrRoutesMissingNew = errors.New("Missing new route")
)

type Routes []*Route

type Route struct {
	Name    string      `json:"name"`
	AppName string      `json:"appname"`
	Path    string      `json:"path"`
	Image   string      `json:"image"`
	Headers http.Header `json:"headers,omitempty"`

	Requirements
}

var (
	ErrRoutesValidationMissingName    = errors.New("Missing route Name")
	ErrRoutesValidationMissingImage   = errors.New("Missing route Image")
	ErrRoutesValidationMissingAppName = errors.New("Missing route AppName")
	ErrRoutesValidationMissingPath    = errors.New("Missing route Path")
	ErrRoutesValidationInvalidPath    = errors.New("Invalid Path format")
)

func (r *Route) Validate() error {
	var res []error

	if r.Name == "" {
		res = append(res, ErrRoutesValidationMissingName)
	}

	if r.Image == "" {
		res = append(res, ErrRoutesValidationMissingImage)
	}

	if r.AppName == "" {
		res = append(res, ErrRoutesValidationMissingAppName)
	}

	if r.Path == "" {
		res = append(res, ErrRoutesValidationMissingPath)
	}

	if !path.IsAbs(r.Path) {
		res = append(res, ErrRoutesValidationInvalidPath)
	}

	if err := r.Requirements.Validate(); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return apiErrors.CompositeValidationError(res...)
	}

	return nil
}

type RouteFilter struct {
	Path    string
	AppName string
}
