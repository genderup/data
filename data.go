package data

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/inappcloud/jsonapi"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"
)

type body struct {
	Data []map[string]interface{} `json:"data"`
}

func Mux(db *sql.DB) http.Handler {
	wrappedDb := &DB{db}

	m := web.New()
	m.NotFound(jsonapi.NotFoundHandler)
	m.Use(jsonapi.ContentTypeHandler)
	m.Use(middleware.EnvInit)
	m.Use(func(c *web.C, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Env["db"] = wrappedDb
			next.ServeHTTP(w, r)
		})
	})
	// m.Use(collectionHandler)
	m.Use(bodyParserHandler)
	m.Use(dataCheckerHandler)

	m.Get("/:collection", readHandler)
	m.Post("/:collection", createHandler)
	m.Get("/:collection/:id", findHandler)
	m.Put("/:collection/:id", updateHandler)
	m.Delete("/:collection/:id", deleteHandler)

	return m
}

func bodyParserHandler(c *web.C, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" || r.Method == "PUT" {
			c.Env["body"] = new(body)
			jsonapi.BodyParserHandler(c.Env["body"], next).ServeHTTP(w, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func dataCheckerHandler(c *web.C, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" || r.Method == "PUT" {
			if len(c.Env["body"].(*body).Data) == 0 {
				jsonapi.Error(w, jsonapi.ErrNoData)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// Doesn't work because it runs before route handling...
// func collectionHandler(c *web.C, next http.Handler) http.Handler {
// 	coll := c.URLParams["collection"]
//
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		if !collectionExists(coll) {
// 			jsonapi.Error(w, jsonapi.ErrCollectionNotFound(coll))
// 			return
// 		}
//
// 		next.ServeHTTP(w, r)
// 	})
// }

func readHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	coll := c.URLParams["collection"]
	db := c.Env["db"].(*DB)

	if !db.collectionExists(coll) {
		jsonapi.Error(w, jsonapi.ErrCollectionNotFound(coll))
		return
	}

	rows, err := db.readQuery(coll, r.URL.Query())
	if err != nil {
		jsonapi.Error(w, err)
		return
	}

	body := new(body)
	err = mapScan(rows, body)
	if err != nil {
		jsonapi.Error(w, err)
		return
	}

	if len(body.Data) == 0 {
		w.Write([]byte(`{"data":[]}` + "\n"))
		return
	}

	json.NewEncoder(w).Encode(body)
}

func findHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	coll := c.URLParams["collection"]
	id := c.URLParams["id"]
	db := c.Env["db"].(*DB)

	if !db.collectionExists(coll) {
		jsonapi.Error(w, jsonapi.ErrCollectionNotFound(coll))
		return
	}

	rows, err := db.findQuery(coll, id, r.URL.Query())
	if err != nil {
		jsonapi.Error(w, err)
		return
	}

	body := new(body)
	err = mapScan(rows, body)
	if err != nil {
		jsonapi.Error(w, err)
		return
	}

	if len(body.Data) == 0 {
		jsonapi.Error(w, jsonapi.ErrResourceNotFound(coll, id))
		return
	}

	json.NewEncoder(w).Encode(body)
}

func createHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	coll := c.URLParams["collection"]
	data := c.Env["body"].(*body).Data[0]
	db := c.Env["db"].(*DB)

	if !db.collectionExists(coll) {
		jsonapi.Error(w, jsonapi.ErrCollectionNotFound(coll))
		return
	}

	rows, err := db.insertQuery(coll, data)
	if err != nil {
		jsonapi.Error(w, err)
		return
	}

	body := new(body)
	err = mapScan(rows, body)
	if err != nil {
		jsonapi.Error(w, err)
		return
	}

	w.WriteHeader(201)
	json.NewEncoder(w).Encode(body)
}

func updateHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	coll := c.URLParams["collection"]
	id := c.URLParams["id"]
	data := c.Env["body"].(*body).Data[0]
	db := c.Env["db"].(*DB)

	if !db.collectionExists(coll) {
		jsonapi.Error(w, jsonapi.ErrCollectionNotFound(coll))
		return
	}

	rows, err := db.updateQuery(coll, id, data, r.URL.Query())
	if err != nil {
		jsonapi.Error(w, err)
		return
	}

	body := new(body)
	err = mapScan(rows, body)
	if err != nil {
		jsonapi.Error(w, err)
		return
	}

	if len(body.Data) == 0 {
		jsonapi.Error(w, jsonapi.ErrResourceNotFound(coll, id))
		return
	}

	json.NewEncoder(w).Encode(body)
}

func deleteHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	coll := c.URLParams["collection"]
	id := c.URLParams["id"]
	db := c.Env["db"].(*DB)

	if !db.collectionExists(coll) {
		jsonapi.Error(w, jsonapi.ErrCollectionNotFound(coll))
		return
	}

	res, err := db.deleteQuery(coll, id, r.URL.Query())
	if err != nil {
		jsonapi.Error(w, err)
		return
	}

	count, _ := res.RowsAffected()
	if count == 0 {
		jsonapi.Error(w, jsonapi.ErrResourceNotFound(coll, id))
		return
	}

	w.WriteHeader(204)
}
