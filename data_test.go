package data_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/inappcloud/data"
	"github.com/inappcloud/jsonapi"
)

var readTestData = []struct {
	path string
	code int
	body string
}{
	{"/posts", 200, `{"data":[{"id":1,"name":"foobar"},{"id":2,"name":"foofoo"}]}` + "\n"},
	{"/comments", 200, `{"data":[{"body":"barfoo","id":1}]}` + "\n"},
	{`/posts?where={"name":"title"}`, 200, `{"data":[]}` + "\n"},
	{`/posts?where={"name":"foobar"}`, 200, `{"data":[{"id":1,"name":"foobar"}]}` + "\n"},
	{"/posts?limit=1", 200, `{"data":[{"id":1,"name":"foobar"}]}` + "\n"},
	{"/posts?offset=1", 200, `{"data":[{"id":2,"name":"foofoo"}]}` + "\n"},
	{"/posts?fields=id", 200, `{"data":[{"id":1},{"id":2}]}` + "\n"},
	{"/nocollection", 404, err(jsonapi.ErrCollectionNotFound("nocollection"))},
	{"/POSTS", 404, err(jsonapi.ErrCollectionNotFound("POSTS"))},
}

func TestRead(t *testing.T) {
	db := NewTestDB()
	defer db.Close()

	db.Exec(`INSERT INTO posts (name) VALUES ('foobar')`)
	db.Exec(`INSERT INTO posts (name) VALUES ('foofoo')`)
	db.Exec(`INSERT INTO comments (body) VALUES ('barfoo')`)

	for _, test := range readTestData {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("GET", test.path, nil)

		data.Mux(db).ServeHTTP(w, r)

		eq(t, test.code, w.Code)
		eq(t, test.body, w.Body.String())
	}
}

var findTestData = []struct {
	path string
	code int
	body string
}{
	{"/posts/1", 200, `{"data":[{"id":1,"name":"foobar"}]}` + "\n"},
	{"/comments/1", 200, `{"data":[{"body":"barfoo","id":1}]}` + "\n"},
	{`/posts/1?where={"name":"foobar"}`, 200, `{"data":[{"id":1,"name":"foobar"}]}` + "\n"},
	{`/posts/1?where={"name":"foo"}`, 404, err(jsonapi.ErrResourceNotFound("posts", "1"))},
	{"/nocollection/1", 404, err(jsonapi.ErrCollectionNotFound("nocollection"))},
	{"/posts/2", 404, err(jsonapi.ErrResourceNotFound("posts", "2"))},
}

func TestFind(t *testing.T) {
	db := NewTestDB()
	defer db.Close()

	db.Exec(`INSERT INTO posts (name) VALUES ('foobar')`)
	db.Exec(`INSERT INTO comments (body) VALUES ('barfoo')`)

	for _, test := range findTestData {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("GET", test.path, nil)

		data.Mux(db).ServeHTTP(w, r)

		eq(t, test.code, w.Code)
		eq(t, test.body, w.Body.String())
	}
}

var createTestData = []struct {
	path    string
	body    string
	expCode int
	expBody string
}{
	{"/posts", `{"data":[{"name":"foobar"}]}`, 201, `{"data":[{"id":1,"name":"foobar"}]}` + "\n"},
	{"/comments", `{"data":[{"body":"foofoo"}]}`, 201, `{"data":[{"body":"foofoo","id":1}]}` + "\n"},
	{"/posts", `{"data":[{"unkown":"foofoo"}]}`, 201, `{"data":[{"id":2,"name":null}]}` + "\n"},
	{"/nocollection", `{"data":[{"name":"foobar"}]}`, 404, err(jsonapi.ErrCollectionNotFound("nocollection"))},
	{"/posts", `{"data":[]}`, 422, err(jsonapi.ErrNoData)},
	{"/posts", `{}`, 422, err(jsonapi.ErrNoData)},
	{"/posts", ``, 400, err(jsonapi.ErrBadRequest)},
}

func TestCreate(t *testing.T) {
	db := NewTestDB()
	defer db.Close()

	for _, test := range createTestData {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", test.path, bytes.NewBufferString(test.body))

		data.Mux(db).ServeHTTP(w, r)

		eq(t, test.expCode, w.Code)
		eq(t, test.expBody, w.Body.String())
	}
}

var updateTestData = []struct {
	path    string
	body    string
	expCode int
	expBody string
}{
	{"/posts/1", `{"data":[{"name":"newfoobar"}]}`, 200, `{"data":[{"id":1,"name":"newfoobar"}]}` + "\n"},
	{"/comments/1", `{"data":[{"body":"newfoofoo"}]}`, 200, `{"data":[{"body":"newfoofoo","id":1}]}` + "\n"},
	{"/posts/1", `{"data":[{"unknown":"newfoobar2"}]}`, 200, `{"data":[{"id":1,"name":"newfoobar"}]}` + "\n"},
	{`/posts/1?where={"name":"newfoobar"}`, `{"data":[{"name":"newfoobar2"}]}`, 200, `{"data":[{"id":1,"name":"newfoobar2"}]}` + "\n"},
	{`/posts/1?where={"name":"foo"}`, `{"data":[{"name":"newfoobar3"}]}`, 404, err(jsonapi.ErrResourceNotFound("posts", "1"))},
	{"/posts/1", `{"data":[]}`, 422, err(jsonapi.ErrNoData)},
	{"/posts/1", `{}`, 422, err(jsonapi.ErrNoData)},
	{"/posts/1", ``, 400, err(jsonapi.ErrBadRequest)},
	{"/nocollection/1", `{"data":[{"name":"newfoobar"}]}`, 404, err(jsonapi.ErrCollectionNotFound("nocollection"))},
	{"/posts/2", `{"data":[{"name":"newfoobar"}]}`, 404, err(jsonapi.ErrResourceNotFound("posts", "2"))},
}

func TestUpdate(t *testing.T) {
	db := NewTestDB()
	defer db.Close()

	db.Exec(`INSERT INTO posts (name) VALUES ('foobar')`)
	db.Exec(`INSERT INTO comments (body) VALUES ('barfoo')`)

	for _, test := range updateTestData {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("PUT", test.path, bytes.NewBufferString(test.body))

		data.Mux(db).ServeHTTP(w, r)

		eq(t, test.expCode, w.Code)
		eq(t, test.expBody, w.Body.String())
	}
}

var deleteTestData = []struct {
	path    string
	expCode int
	expBody string
}{
	{"/posts/1", 204, ""},
	{"/nocollection/1", 404, err(jsonapi.ErrCollectionNotFound("nocollection"))},
	{`/posts/2?where={"name":"foo"}`, 404, err(jsonapi.ErrResourceNotFound("posts", "2"))},
	{`/posts/2?where={"name":"foofoo"}`, 204, ""},
	{"/posts/3", 404, err(jsonapi.ErrResourceNotFound("posts", "2"))},
}

func TestDelete(t *testing.T) {
	db := NewTestDB()
	defer db.Close()

	db.Exec(`INSERT INTO posts (name) VALUES ('foobar')`)
	db.Exec(`INSERT INTO posts (name) VALUES ('foofoo')`)

	for _, test := range deleteTestData {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("DELETE", test.path, nil)

		data.Mux(db).ServeHTTP(w, r)

		eq(t, test.expCode, w.Code)
	}
}
