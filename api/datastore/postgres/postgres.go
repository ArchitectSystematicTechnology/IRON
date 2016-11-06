package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/Sirupsen/logrus"
	"github.com/iron-io/functions/api/models"
	_ "github.com/lib/pq"
)

const routesTableCreate = `
CREATE TABLE IF NOT EXISTS routes (
	app_name character varying(256) NOT NULL,
	path text NOT NULL,
    image character varying(256) NOT NULL,
	memory integer NOT NULL,
	headers text NOT NULL,
	config text NOT NULL,
	PRIMARY KEY (app_name, path)
);`

const appsTableCreate = `CREATE TABLE IF NOT EXISTS apps (
    name character varying(256) NOT NULL PRIMARY KEY,
	config text NOT NULL
);`

const extrasTableCreate = `CREATE TABLE IF NOT EXISTS extras (
    key character varying(256) NOT NULL PRIMARY KEY,
	value character varying(256) NOT NULL
);`

const routeSelector = `SELECT app_name, path, image, memory, headers, config FROM routes`

type rowScanner interface {
	Scan(dest ...interface{}) error
}

type rowQuerier interface {
	QueryRow(query string, args ...interface{}) *sql.Row
}

type PostgresDatastore struct {
	db *sql.DB
}

func New(url *url.URL) (models.Datastore, error) {
	db, err := sql.Open("postgres", url.String())
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	maxIdleConns := 30 // c.MaxIdleConnections
	db.SetMaxIdleConns(maxIdleConns)
	logrus.WithFields(logrus.Fields{"max_idle_connections": maxIdleConns}).Info("Postgres dialed")

	pg := &PostgresDatastore{
		db: db,
	}

	for _, v := range []string{routesTableCreate, appsTableCreate, extrasTableCreate} {
		_, err = db.Exec(v)
		if err != nil {
			return nil, err
		}
	}

	return pg, nil
}

func (ds *PostgresDatastore) StoreApp(app *models.App) (*models.App, error) {
	cbyte, err := json.Marshal(app.Config)
	if err != nil {
		return nil, err
	}

	_, err = ds.db.Exec(`
	  INSERT INTO apps (name, config)
		VALUES ($1, $2)
		ON CONFLICT (name) DO UPDATE SET
			config = $2;
	`,
		app.Name,
		string(cbyte),
	)

	if err != nil {
		return nil, err
	}

	return app, nil
}

func (ds *PostgresDatastore) RemoveApp(appName string) error {
	_, err := ds.db.Exec(`
	  DELETE FROM apps
	  WHERE name = $1
	`, appName)

	if err != nil {
		return err
	}

	return nil
}

func (ds *PostgresDatastore) GetApp(name string) (*models.App, error) {
	row := ds.db.QueryRow("SELECT name, config FROM apps WHERE name=$1", name)

	var resName string
	var config string
	err := row.Scan(&resName, &config)

	res := &models.App{
		Name: resName,
	}

	json.Unmarshal([]byte(config), &res.Config)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func scanApp(scanner rowScanner, app *models.App) error {
	err := scanner.Scan(
		&app.Name,
	)

	return err
}

func (ds *PostgresDatastore) GetApps(filter *models.AppFilter) ([]*models.App, error) {
	res := []*models.App{}

	rows, err := ds.db.Query(`
		SELECT DISTINCT *
		FROM apps`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var app models.App
		err := scanApp(rows, &app)

		if err != nil {
			return nil, err
		}
		res = append(res, &app)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (ds *PostgresDatastore) StoreRoute(route *models.Route) (*models.Route, error) {
	hbyte, err := json.Marshal(route.Headers)
	if err != nil {
		return nil, err
	}

	cbyte, err := json.Marshal(route.Config)
	if err != nil {
		return nil, err
	}

	_, err = ds.db.Exec(`
		INSERT INTO routes (
			app_name, 
			path, 
			image,
			memory,
			headers,
			config
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (app_name, path) DO UPDATE SET
			path = $2,
			image = $3,
			memory = $4,
			headers = $5,
			config = $6;
		`,
		route.AppName,
		route.Path,
		route.Image,
		route.Memory,
		string(hbyte),
		string(cbyte),
	)

	if err != nil {
		return nil, err
	}
	return route, nil
}

func (ds *PostgresDatastore) RemoveRoute(appName, routePath string) error {
	_, err := ds.db.Exec(`
		DELETE FROM routes
		WHERE path = $1 AND app_name = $2
	`, routePath, appName)

	if err != nil {
		return err
	}
	return nil
}

func scanRoute(scanner rowScanner, route *models.Route) error {
	var headerStr string
	var configStr string

	err := scanner.Scan(
		&route.AppName,
		&route.Path,
		&route.Image,
		&route.Memory,
		&headerStr,
		&configStr,
	)

	if headerStr == "" {
		return models.ErrRoutesNotFound
	}

	json.Unmarshal([]byte(headerStr), &route.Headers)
	json.Unmarshal([]byte(configStr), &route.Config)

	return err
}

func getRoute(qr rowQuerier, routePath string) (*models.Route, error) {
	var route models.Route

	row := qr.QueryRow(fmt.Sprintf("%s WHERE path=$1", routeSelector), routePath)
	err := scanRoute(row, &route)

	if err == sql.ErrNoRows {
		return nil, models.ErrRoutesNotFound
	} else if err != nil {
		return nil, err
	}
	return &route, nil
}

func (ds *PostgresDatastore) GetRoute(appName, routePath string) (*models.Route, error) {
	return getRoute(ds.db, routePath)
}

func (ds *PostgresDatastore) GetRoutes(filter *models.RouteFilter) ([]*models.Route, error) {
	res := []*models.Route{}
	filterQuery := buildFilterQuery(filter)
	rows, err := ds.db.Query(fmt.Sprintf("%s %s", routeSelector, filterQuery))
	// todo: check for no rows so we don't respond with a sql 500 err
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var route models.Route
		err := scanRoute(rows, &route)
		if err != nil {
			continue
		}
		res = append(res, &route)

	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (ds *PostgresDatastore) GetRoutesByApp(appName string, filter *models.RouteFilter) ([]*models.Route, error) {
	res := []*models.Route{}
	filter.AppName = appName
	filterQuery := buildFilterQuery(filter)
	rows, err := ds.db.Query(fmt.Sprintf("%s %s", routeSelector, filterQuery))
	// todo: check for no rows so we don't respond with a sql 500 err
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var route models.Route
		err := scanRoute(rows, &route)
		if err != nil {
			continue
		}
		res = append(res, &route)

	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func buildFilterQuery(filter *models.RouteFilter) string {
	filterQuery := ""

	filterQueries := []string{}
	if filter.Path != "" {
		filterQueries = append(filterQueries, fmt.Sprintf("path = '%s'", filter.Path))
	}

	if filter.AppName != "" {
		filterQueries = append(filterQueries, fmt.Sprintf("app_name = '%s'", filter.AppName))
	}

	if filter.Image != "" {
		filterQueries = append(filterQueries, fmt.Sprintf("image = '%s'", filter.Image))
	}

	for i, field := range filterQueries {
		if i == 0 {
			filterQuery = fmt.Sprintf("WHERE %s ", field)
		} else {
			filterQuery = fmt.Sprintf("%s AND %s", filterQuery, field)
		}
	}

	return filterQuery
}

func (ds *PostgresDatastore) Put(key, value []byte) error {
	_, err := ds.db.Exec(`
	    INSERT INTO extras (
			key,
			value
		)
		VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE SET
			value = $1;
		`, value)

	if err != nil {
		return err
	}

	return nil
}

func (ds *PostgresDatastore) Get(key []byte) ([]byte, error) {
	row := ds.db.QueryRow("SELECT value FROM extras WHERE key=$1", key)

	var value []byte
	err := row.Scan(&value)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return value, nil
}
