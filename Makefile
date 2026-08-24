# Makefile for Online Job Portal Backend API

.PHONY: run test seed

run:
	go run ./cmd/api

test:
	go test ./...

seed:
	go run ./scripts/seed.go
