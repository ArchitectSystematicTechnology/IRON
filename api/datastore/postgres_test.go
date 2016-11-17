package datastore

import (
	"database/sql"
	"fmt"
	"os/exec"
	"testing"
	"time"

	"github.com/iron-io/functions/api/models"
)

const tmpPostgres = "postgres://postgres@127.0.0.1:15432/funcs?sslmode=disable"

func preparePostgresTest() func() {
	fmt.Println("initializing postgres for test")
	exec.Command("docker", "rm", "-f", "iron-postgres-test").Run()
	exec.Command("docker", "run", "--name", "iron-postgres-test", "-p", "15432:5432", "-d", "postgres").Run()
	for {
		db, err := sql.Open("postgres", "postgres://postgres@127.0.0.1:15432?sslmode=disable")
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		db.Exec(`CREATE DATABASE funcs;`)
		_, err = db.Exec(`GRANT ALL PRIVILEGES ON DATABASE funcs TO postgres;`)
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	fmt.Println("postgres for test ready")
	return func() {
		exec.Command("docker", "rm", "-f", "iron-postgres-test").Run()
	}
}

func TestPostgres(t *testing.T) {
	close := preparePostgresTest()
	defer close()
	buf := setLogBuffer()

	ds, err := New(tmpPostgres)
	if err != nil {
		t.Fatalf("Error when creating datastore: %v", err)
	}

	testApp := &models.App{
		Name: "Test",
	}

	testRoute := &models.Route{
		AppName: testApp.Name,
		Path:    "/test",
		Image:   "iron/hello",
	}

	// Testing insert app
	_, err = ds.InsertApp(nil)
	if err != models.ErrDatastoreEmptyApp {
		t.Log(buf.String())
		t.Fatalf("Test InsertApp(nil): expected error `%v`, but it was `%v`", models.ErrDatastoreEmptyApp, err)
	}

	_, err = ds.InsertApp(&models.App{})
	if err != models.ErrDatastoreEmptyAppName {
		t.Log(buf.String())
		t.Fatalf("Test InsertApp(nil): expected error `%v`, but it was `%v`", models.ErrDatastoreEmptyAppName, err)
	}

	_, err = ds.InsertApp(testApp)
	if err != nil {
		t.Log(buf.String())
		t.Fatalf("Test InsertApp: unexpected error: %v", err)
	}

	_, err = ds.InsertApp(testApp)
	if err != models.ErrAppsAlreadyExists {
		t.Log(buf.String())
		t.Fatalf("Test InsertApp duplicated: expected error `%v`, but it was `%v`", models.ErrAppsAlreadyExists, err)
	}

	_, err = ds.UpdateApp(&models.App{
		Name: testApp.Name,
		Config: map[string]string{
			"TEST": "1",
		},
	})
	if err != nil {
		t.Log(buf.String())
		t.Fatalf("Test UpdateApp: error when updating app: %v", err)
	}

	// Testing get app
	_, err = ds.GetApp("")
	if err != models.ErrDatastoreEmptyAppName {
		t.Log(buf.String())
		t.Fatalf("Test GetApp: expected error to be `%v`, but it was `%v`", models.ErrDatastoreEmptyAppName, err)
	}

	app, err := ds.GetApp(testApp.Name)
	if err != nil {
		t.Log(buf.String())
		t.Fatalf("Test GetApp: unexpected error: %v", err)
	}
	if app.Name != testApp.Name {
		t.Log(buf.String())
		t.Fatalf("Test GetApp: expected app name to be `%s`, but it was `%s`", app.Name, testApp.Name)
	}

	// Testing list apps
	apps, err := ds.GetApps(&models.AppFilter{})
	if err != nil {
		t.Log(buf.String())
		t.Fatalf("Test GetApps: unexpected error: %v", err)
	}
	if len(apps) == 0 {
		t.Fatal("Test GetApps: expected result count to be greater than 0")
	}
	if apps[0].Name != testApp.Name {
		t.Log(buf.String())
		t.Fatalf("Test GetApps: expected `app.Name` to be `%s` but it was `%s`", app.Name, testApp.Name)
	}

	// Testing app delete
	err = ds.RemoveApp("")
	if err != models.ErrDatastoreEmptyAppName {
		t.Log(buf.String())
		t.Fatalf("Test RemoveApp: expected error `%v`, but it was `%v`", models.ErrDatastoreEmptyAppName, err)
	}

	err = ds.RemoveApp(testApp.Name)
	if err != nil {
		t.Log(buf.String())
		t.Fatalf("Test RemoveApp: unexpected error: %v", err)
	}
	app, err = ds.GetApp(testApp.Name)
	if err != models.ErrAppsNotFound {
		t.Log(buf.String())
		t.Fatalf("Test GetApp(removed): expected error `%v`, but it was `%v`", models.ErrAppsNotFound, err)
	}
	if app != nil {
		t.Log(buf.String())
		t.Fatalf("Test RemoveApp: failed to remove the app")
	}

	err = ds.RemoveApp(testApp.Name)
	if err != models.ErrAppsNotFound {
		t.Log(buf.String())
		t.Fatalf("Test RemoveApp: expected error `%v`, but it was `%v`", models.ErrAppsNotFound, err)
	}

	// Test update inexistent app
	_, err = ds.UpdateApp(&models.App{
		Name: testApp.Name,
		Config: map[string]string{
			"TEST": "1",
		},
	})
	if err != models.ErrAppsNotFound {
		t.Log(buf.String())
		t.Fatalf("Test UpdateApp(inexistent): expected error `%v`, but it was `%v`", models.ErrAppsNotFound, err)
	}

	// Insert app again to test routes
	ds.InsertApp(testApp)

	// Testing insert route
	_, err = ds.InsertRoute(nil)
	if err != models.ErrDatastoreEmptyRoute {
		t.Log(buf.String())
		t.Fatalf("Test InsertRoute(nil): expected error `%v`, but it was `%v`", models.ErrDatastoreEmptyRoute, err)
	}

	_, err = ds.InsertRoute(testRoute)
	if err != nil {
		t.Log(buf.String())
		t.Fatalf("Test InsertRoute: unexpected error: %v", err)
	}

	_, err = ds.InsertRoute(testRoute)
	if err != models.ErrRoutesAlreadyExists {
		t.Log(buf.String())
		t.Fatalf("Test InsertRoute duplicated: expected error to be `%v`, but it was `%v`", models.ErrRoutesAlreadyExists, err)
	}

	_, err = ds.UpdateRoute(testRoute)
	if err != nil {
		t.Log(buf.String())
		t.Fatalf("Test UpdateRoute: unexpected error: %v", err)
	}

	// Testing get
	_, err = ds.GetRoute("a", "")
	if err != models.ErrDatastoreEmptyRoutePath {
		t.Log(buf.String())
		t.Fatalf("Test GetRoute(empty route path): expected error `%v`, but it was `%v`", models.ErrDatastoreEmptyRoutePath, err)
	}

	_, err = ds.GetRoute("", "a")
	if err != models.ErrDatastoreEmptyAppName {
		t.Log(buf.String())
		t.Fatalf("Test GetRoute(empty app name): expected error `%v`, but it was `%v`", models.ErrDatastoreEmptyAppName, err)
	}

	route, err := ds.GetRoute(testApp.Name, testRoute.Path)
	if err != nil {
		t.Log(buf.String())
		t.Fatalf("Test GetRoute: unexpected error %v", err)
	}
	if route.Path != testRoute.Path {
		t.Log(buf.String())
		t.Fatalf("Test GetRoute: expected route.Path to be `%v`, but it was `%v`", route.Path, testRoute.Path)
	}

	// Testing list routes
	routes, err := ds.GetRoutesByApp(testApp.Name, &models.RouteFilter{})
	if err != nil {
		t.Log(buf.String())
		t.Fatalf("Test GetRoutes: unexpected error %v", err)
	}
	if len(routes) == 0 {
		t.Fatal("Test GetRoutes: expected result count to be greater than 0")
	}
	if routes[0].Path != testRoute.Path {
		t.Log(buf.String())
		t.Fatalf("Test GetRoutes: expected app.Name to be `%v` but it was `%v`", testRoute.Path, routes[0].Path)
	}

	// Testing list routes
	routes, err = ds.GetRoutes(&models.RouteFilter{Image: testRoute.Image})
	if err != nil {
		t.Log(buf.String())
		t.Fatalf("Test GetRoutes: unexpected error: %v", err)
	}
	if len(routes) == 0 {
		t.Fatal("Test GetRoutes: expected result count to be greater than 0")
	}
	if routes[0].Path != testRoute.Path {
		t.Log(buf.String())
		t.Fatalf("Test GetRoutes: expected app.Name to be `%v` but it was `%v`", testRoute.Path, routes[0].Path)
	}

	// Testing app delete
	err = ds.RemoveRoute("", "")
	if err != models.ErrDatastoreEmptyAppName {
		t.Log(buf.String())
		t.Fatalf("Test RemoveRoute(empty app name): expected error `%v`, but it was `%v`", models.ErrDatastoreEmptyAppName, err)
	}

	err = ds.RemoveRoute("a", "")
	if err != models.ErrDatastoreEmptyRoutePath {
		t.Log(buf.String())
		t.Fatalf("Test RemoveRoute(empty route path): expected error `%v`, but it was `%v`", models.ErrDatastoreEmptyRoutePath, err)
	}

	err = ds.RemoveRoute(testRoute.AppName, testRoute.Path)
	if err != nil {
		t.Log(buf.String())
		t.Fatalf("Test RemoveRoute: unexpected error: %v", err)
	}

	err = ds.RemoveRoute(testRoute.AppName, testRoute.Path)
	if err != models.ErrRoutesNotFound {
		t.Log(buf.String())
		t.Fatalf("Test RemoveRoute: expected error `%v`, but it was `%v`", models.ErrRoutesNotFound, err)
	}

	_, err = ds.UpdateRoute(&models.Route{
		AppName: testRoute.AppName,
		Path:    testRoute.Path,
		Image:   "test",
	})
	if err != models.ErrRoutesNotFound {
		t.Log(buf.String())
		t.Fatalf("Test UpdateRoute inexistent: expected error to be `%v`, but it was `%v`", models.ErrRoutesNotFound, err)
	}

	route, err = ds.GetRoute(testRoute.AppName, testRoute.Path)
	if err != models.ErrRoutesNotFound {
		t.Log(buf.String())
		t.Fatalf("Test GetRoute: expected error `%v`, but it was `%v`", models.ErrRoutesNotFound, err)
	}
	if route != nil {
		t.Log(buf.String())
		t.Fatalf("Test RemoveRoute: failed to remove the route")
	}
}
