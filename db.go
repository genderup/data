package data

import (
	"database/sql"
	"net/url"
	"strings"

	"github.com/inappcloud/query"
	"github.com/inappcloud/query/where"
)

type DB struct {
	*sql.DB
}

func mapScan(r *sql.Rows, body *body) error {
	defer r.Close()

	columns, err := r.Columns()
	if err != nil {
		return err
	}

	for r.Next() {
		dest := make(map[string]interface{})

		values := make([]interface{}, len(columns))
		for i := range values {
			values[i] = new(interface{})
		}

		err = r.Scan(values...)
		if err != nil {
			return err
		}

		for i, column := range columns {
			dest[column] = *(values[i].(*interface{}))
		}

		body.Data = append(body.Data, dest)
	}

	for i, doc := range body.Data {
		for k, v := range doc {
			switch t := v.(type) {
			case []byte:
				body.Data[i][k] = string(t)
			}
		}
	}

	return r.Err()
}

func (db *DB) readQuery(coll string, qs url.Values) (*sql.Rows, error) {
	q := query.Select(coll).Limit(qs.Get("limit")).Offset(qs.Get("offset")).Fields(qs.Get("fields")).Where(where.Parse(qs.Get("where")))
	return db.Query(q.String(), q.Params()...)
}

func (db *DB) findQuery(coll string, id string, qs url.Values) (*sql.Rows, error) {
	q := query.Select(coll).Limit(qs.Get("limit")).Offset(qs.Get("offset")).Fields(qs.Get("fields")).Where(where.And(where.Parse(qs.Get("where")), where.Eq("id", id)))
	return db.Query(q.String(), q.Params()...)
}

func (db *DB) deleteQuery(coll string, id string, qs url.Values) (sql.Result, error) {
	q := query.Delete(coll).Limit(qs.Get("limit")).Offset(qs.Get("offset")).Where(where.And(where.Parse(qs.Get("where")), where.Eq("id", id)))
	return db.Exec(q.String(), q.Params()...)
}

func (db *DB) columnsQuery(coll string, data map[string]interface{}) ([]string, []string, []interface{}, error) {
	rows, err := db.Query("SELECT column_name FROM information_schema.columns WHERE table_schema = 'public' AND table_name = $1 AND column_name != 'id'", coll)

	if err != nil {
		return nil, nil, nil, err
	}

	defer rows.Close()
	cols := []string{}

	for rows.Next() {
		var col string

		err = rows.Scan(&col)
		if err != nil {
			return nil, nil, nil, err
		}

		cols = append(cols, col)
	}

	if err = rows.Err(); err != nil {
		return nil, nil, nil, err
	}

	keys := []string{}
	values := []interface{}{}
	for k, v := range data {
		if contains(cols, k) {
			keys = append(keys, k)
			values = append(values, v)
		}
	}

	return cols, keys, values, nil
}

func (db *DB) insertQuery(coll string, data map[string]interface{}) (*sql.Rows, error) {
	_, keys, values, err := db.columnsQuery(coll, data)

	if err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		return db.Query(query.Insert(coll).Returning("*").String())
	}

	q := query.Insert(coll).Fields(strings.Join(keys, ",")).Values(values...).Returning("*")
	return db.Query(q.String(), q.Params()...)
}

func (db *DB) updateQuery(coll string, id string, data map[string]interface{}, qs url.Values) (*sql.Rows, error) {
	_, keys, values, err := db.columnsQuery(coll, data)

	if err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		q := query.Select(coll).Where(where.Eq("id", id))
		return db.Query(q.String(), q.Params()...)
	}

	q := query.Update(coll).Fields(strings.Join(keys, ",")).Values(values...).Where(where.And(where.Parse(qs.Get("where")), where.Eq("id", id))).Returning("*")
	return db.Query(q.String(), q.Params()...)
}

func (db *DB) collectionExists(coll string) bool {
	var count int
	q := query.Select("pg_tables").Fields("COUNT(*)").Where(where.And(where.Eq("schemaname", "public"), where.Eq("tablename", coll)))
	db.QueryRow(q.String(), q.Params()...).Scan(&count)
	return count > 0
}

func contains(arr []string, s string) bool {
	for _, v := range arr {
		if s == v {
			return true
		}
	}

	return false
}
