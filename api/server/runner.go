package server

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/iron-io/functions/api/models"
	"github.com/iron-io/functions/api/runner"
	titancommon "github.com/iron-io/titan/common"
	"github.com/satori/go.uuid"
)

func handleSpecial(c *gin.Context) {
	ctx := c.MustGet("ctx").(context.Context)
	log := titancommon.Logger(ctx)

	err := Api.UseSpecialHandlers(c)
	if err != nil {
		log.WithError(err).Errorln("Error using special handler!")
		// todo: what do we do here? Should probably return a 500 or something
	}
}

func handleRunner(c *gin.Context) {
	if strings.HasPrefix(c.Request.URL.Path, "/v1") {
		c.Status(http.StatusNotFound)
		return
	}

	ctx := c.MustGet("ctx").(context.Context)
	log := titancommon.Logger(ctx)

	reqID := uuid.NewV5(uuid.Nil, fmt.Sprintf("%s%s%d", c.Request.RemoteAddr, c.Request.URL.Path, time.Now().Unix())).String()
	c.Set("reqID", reqID) // todo: put this in the ctx instead of gin's

	log = log.WithFields(logrus.Fields{"request_id": reqID})

	var err error
	var payload io.Reader

	if c.Request.Method == "POST" || c.Request.Method == "PUT" {
		payload = c.Request.Body
	} else if c.Request.Method == "GET" {
		reqPayload := c.Request.URL.Query().Get("payload")
		if len(reqPayload) > 0 {
			payload = strings.NewReader(reqPayload)
		}
	}

	// Load complete body and close
	defer func() {
		io.Copy(ioutil.Discard, c.Request.Body)
		c.Request.Body.Close()
	}()

	appName := c.Param("app")
	if appName == "" {
		// check context, app can be added via special handlers
		a, ok := c.Get("app")
		if ok {
			appName = a.(string)
		}
	}
	// if still no appName, we gotta exit
	if appName == "" {
		log.WithError(err).Error(models.ErrAppsNotFound)
		c.JSON(http.StatusBadRequest, simpleError(models.ErrAppsNotFound))
		return
	}
	route := c.Param("route")
	if route == "" {
		route = c.Request.URL.Path
	}

	log.WithFields(logrus.Fields{"app": appName, "path": route}).Debug("Finding route on datastore")

	app, err := Api.Datastore.GetApp(appName)
	if err != nil || app == nil {
		log.WithError(err).Error(models.ErrAppsNotFound)
		c.JSON(http.StatusNotFound, simpleError(models.ErrAppsNotFound))
		return
	}

	routes, err := Api.Datastore.GetRoutesByApp(appName, &models.RouteFilter{})
	if err != nil {
		log.WithError(err).Error(models.ErrRoutesList)
		c.JSON(http.StatusInternalServerError, simpleError(models.ErrRoutesList))
		return
	}

	log.WithField("routes", routes).Debug("Got routes from datastore")
	for _, el := range routes {
		log = log.WithFields(logrus.Fields{
			"app": appName, "route": el.Path, "image": el.Image, "request_id": reqID})

		// Request count metric
		metricBaseName := "server.handleRunner." + appName + "."
		runner.LogMetricCount(ctx, (metricBaseName + "requests"), 1)

		if params, match := matchRoute(el.Path, route); match {

			var stdout bytes.Buffer // TODO: should limit the size of this, error if gets too big. akin to: https://golang.org/pkg/io/#LimitReader
			stderr := runner.NewFuncLogger(appName, route, el.Image, reqID)

			envVars := map[string]string{
				"METHOD":      c.Request.Method,
				"ROUTE":       el.Path,
				"REQUEST_URL": c.Request.URL.String(),
			}

			// app config
			for k, v := range app.Config {
				envVars["CONFIG_"+strings.ToUpper(k)] = v
			}

			// route config
			for k, v := range el.Config {
				envVars["CONFIG_"+strings.ToUpper(k)] = v
			}

			// params
			for _, param := range params {
				envVars["PARAM_"+strings.ToUpper(param.Key)] = param.Value
			}

			// headers
			for header, value := range c.Request.Header {
				envVars["HEADER_"+strings.ToUpper(header)] = strings.Join(value, " ")
			}

			cfg := &runner.Config{
				Image:   el.Image,
				Timeout: 30 * time.Second,
				ID:      reqID,
				AppName: appName,
				Stdout:  &stdout,
				Stderr:  stderr,
				Input:   payload,
				Env: map[string]string{
					"REQUEST_URL": c.Request.URL.String(),
				},
			}

			metricStart := time.Now()
			if result, err := Api.Runner.Run(c, cfg); err != nil {
				log.WithError(err).Error(models.ErrRunnerRunRoute)
				c.JSON(http.StatusInternalServerError, simpleError(models.ErrRunnerRunRoute))
			} else {
				for k, v := range el.Headers {
					c.Header(k, v[0])
				}

				if result.Status() == "success" {
					c.Data(http.StatusOK, "", stdout.Bytes())
					runner.LogMetricCount(ctx, (metricBaseName + "succeeded"), 1)

				} else {
					// log.WithFields(logrus.Fields{"app": appName, "route": el, "req_id": reqID}).Debug(stderr.String())
					// Error count metric
					runner.LogMetricCount(ctx, (metricBaseName + "error"), 1)

					c.AbortWithStatus(http.StatusInternalServerError)
				}
			}
			// Execution time metric
			metricElapsed := time.Since(metricStart)
			runner.LogMetricTime(ctx, (metricBaseName + "time"), metricElapsed)
			runner.LogMetricTime(ctx, "server.handleRunner.exec_time", metricElapsed)
			return
		}
	}

	log.WithError(err).Error(models.ErrRunnerRouteNotFound)
	c.JSON(http.StatusNotFound, simpleError(models.ErrRunnerRouteNotFound))
}

var fakeHandler = func(http.ResponseWriter, *http.Request, Params) {}

func matchRoute(baseRoute, route string) (Params, bool) {
	tree := &node{}
	tree.addRoute(baseRoute, fakeHandler)
	handler, p, _ := tree.getValue(route)
	if handler == nil {
		return nil, false
	}

	return p, true
}
