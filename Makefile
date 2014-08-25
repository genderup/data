TEST?=./
TEST_DB=postgres://localhost/inappcloud_data_test?sslmode=disable

default: test

setup: updatedeps
	psql -c 'create database inappcloud_data;'
	echo "DATABASE_URL=postgres://localhost/inappcloud_data?sslmode=disable" > .env

server:
	env $$(cat .env) go run cmd/data-server/main.go

updatedeps:
	go get -u -v ./...
	go get -u -v github.com/lib/pq

test:
	psql -c 'drop database if exists inappcloud_data_test;'
	psql -c 'create database inappcloud_data_test;'
	psql -c 'create table posts (id serial, name text); create table comments (id serial, body text);' -d inappcloud_data_test
	env DATABASE_URL=$(TEST_DB) go test $(TEST)
