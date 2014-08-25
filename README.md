# In-App Cloud Data [![Build Status](https://travis-ci.org/inappcloud/data.svg?branch=master)](https://travis-ci.org/inappcloud/data)

A data microservice.
You can use it directly as a REST API or use the package as part of your Go web application.
It is built with Goji and requires PostgreSQL.
`cmd/data-server/main.go` will show you how to use the package in your own code.

# Goal

It's a service to request a PostgreSQL database through a REST API.

# API

It uses [JSON API](http://jsonapi.org) for request and response format.

## Get a collection

```
GET /{collection}
Accept: application/json
```

## Get a document

```
GET /{collection}/{id}
Accept: application/json
```

## Create a document

```
POST /{collection}
Content-Type: application/json
Accept: application/json

{
  "data": [{
    "name": "yame"
  }]
}
```

## Update a document

```
PUT /{collection}/{id}
Content-Type: application/json
Accept: application/json

{
  "data": [{
    "name": "yame"
  }]
}
```

## Delete a document

```
DELETE /{collection}/{id}
```

# Run

## Locally

```
make setup
make server
```

If you have a custom environment for PostgreSQL, please update variable `DATABASE_URL` in `.env`.
It runs on port `8080` by default but you can add `PORT` in `.env` to change the port.

# Test

```
make
```

If you have a custom environment for PostgreSQL, please update `TEST_DB` in `Makefile`.
