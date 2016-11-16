package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/iron-io/functions/api/datastore"
	"github.com/iron-io/functions/api/models"
	"github.com/iron-io/functions/api/mqs"
	"github.com/iron-io/functions/api/runner"
	"github.com/iron-io/runner/common"
)

var tmpBolt = "/tmp/func_test_bolt.db"

func testRouter(s *Server) *gin.Engine {
	r := s.Router
	r.Use(gin.Logger())
	ctx := context.Background()
	r.Use(func(c *gin.Context) {
		ctx, _ := common.LoggerWithFields(ctx, extractFields(c))
		c.Set("ctx", ctx)
		c.Next()
	})
	s.bindHandlers()
	return r
}

func routerRequest(t *testing.T, router *gin.Engine, method, path string, body io.Reader) (*http.Request, *httptest.ResponseRecorder) {
	req, err := http.NewRequest(method, "http://127.0.0.1:8080"+path, body)
	if err != nil {
		t.Fatalf("Test: Could not create %s request to %s: %v", method, path, err)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	return req, rec
}

func newRouterRequest(t *testing.T, method, path string, body io.Reader) (*http.Request, *httptest.ResponseRecorder) {
	req, err := http.NewRequest(method, "http://127.0.0.1:8080"+path, body)
	if err != nil {
		t.Fatalf("Test: Could not create %s request to %s: %v", method, path, err)
	}

	rec := httptest.NewRecorder()

	return req, rec
}

func getErrorResponse(t *testing.T, rec *httptest.ResponseRecorder) models.Error {
	respBody, err := ioutil.ReadAll(rec.Body)
	if err != nil {
		t.Error("Test: Expected not empty response body")
	}

	var errResp models.Error
	err = json.Unmarshal(respBody, &errResp)
	if err != nil {
		t.Error("Test: Expected response body to be a valid models.Error object")
	}

	return errResp
}

func prepareBolt(t *testing.T) (models.Datastore, func()) {
	os.Remove(tmpBolt)
	ds, err := datastore.New("bolt://" + tmpBolt)
	if err != nil {
		t.Fatal("Error when creating datastore: %s", err)
	}
	return ds, func() {
		os.Remove(tmpBolt)
	}
}

func TestFullStack(t *testing.T) {
	buf := setLogBuffer()
	ds, closeBolt := prepareBolt(t)
	defer closeBolt()

	tasks := make(chan runner.TaskRequest)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go runner.StartWorkers(ctx, 1, testRunner(t), tasks)

	s := New(ds, &mqs.Mock{}, testRunner(t), tasks)
	router := testRouter(s)

	for i, test := range []struct {
		method       string
		path         string
		body         string
		expectedCode int
	}{
		{"POST", "/v1/apps", `{ "app": { "name": "myapp" } }`, http.StatusCreated},
		{"GET", "/v1/apps", ``, http.StatusOK},
		{"GET", "/v1/apps/myapp", ``, http.StatusOK},
		{"POST", "/v1/apps/myapp/routes", `{ "route": { "name": "myroute", "path": "/myroute", "image": "iron/hello" } }`, http.StatusCreated},
		{"POST", "/v1/apps/myapp/routes", `{ "route": { "name": "myroute2", "path": "/myroute2", "image": "iron/error" } }`, http.StatusCreated},
		{"GET", "/v1/apps/myapp/routes/myroute", ``, http.StatusOK},
		{"GET", "/v1/apps/myapp/routes/myroute2", ``, http.StatusOK},
		{"GET", "/v1/apps/myapp/routes", ``, http.StatusOK},
		{"POST", "/r/myapp/myroute", `{ "name": "Teste" }`, http.StatusOK},
		{"POST", "/r/myapp/myroute2", `{ "name": "Teste" }`, http.StatusInternalServerError},
		{"DELETE", "/v1/apps/myapp/routes/myroute", ``, http.StatusOK},
		{"DELETE", "/v1/apps/myapp", ``, http.StatusOK},
		{"GET", "/v1/apps/myapp", ``, http.StatusNotFound},
		{"GET", "/v1/apps/myapp/routes/myroute", ``, http.StatusInternalServerError},
	} {
		_, rec := routerRequest(t, router, test.method, test.path, bytes.NewBuffer([]byte(test.body)))

		if rec.Code != test.expectedCode {
			t.Log(buf.String())
			t.Errorf("Test %d: Expected status code to be %d but was %d",
				i, test.expectedCode, rec.Code)
		}
	}
}
