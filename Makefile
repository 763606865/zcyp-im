APP_NAME := zcyp-im
GOCACHE_DIR := $(CURDIR)/.cache/go-build
MIGRATIONS_DIR := migrations

ifneq (,$(wildcard .env))
include .env
export
endif

MYSQL_HOST ?= $(ZCYP_IM_MYSQL_HOST)
MYSQL_PORT ?= $(ZCYP_IM_MYSQL_PORT)
MYSQL_DATABASE ?= $(ZCYP_IM_MYSQL_DATABASE)
MYSQL_USERNAME ?= $(ZCYP_IM_MYSQL_USERNAME)
MYSQL_PASSWORD ?= $(ZCYP_IM_MYSQL_PASSWORD)

MIGRATE ?= migrate
MYSQL_DSN := mysql://$(MYSQL_USERNAME):$(MYSQL_PASSWORD)@tcp($(MYSQL_HOST):$(MYSQL_PORT))/$(MYSQL_DATABASE)

.PHONY: help run test build migrate-up migrate-down migrate-force

help:
	@echo "Available targets:"
	@echo "  make run           Start the application"
	@echo "  make test          Run go test"
	@echo "  make build         Build the binary"
	@echo "  make migrate-up    Apply all up migrations"
	@echo "  make migrate-down  Roll back one migration"
	@echo "  make migrate-force VERSION=<n>  Force migration version"

run:
	@mkdir -p $(GOCACHE_DIR)
	GOCACHE=$(GOCACHE_DIR) go run .

test:
	@mkdir -p $(GOCACHE_DIR)
	GOCACHE=$(GOCACHE_DIR) go test ./...

build:
	@mkdir -p $(GOCACHE_DIR)
	GOCACHE=$(GOCACHE_DIR) go build -o $(APP_NAME) .

migrate-up:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(MYSQL_DSN)" up

migrate-down:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(MYSQL_DSN)" down 1

migrate-force:
ifndef VERSION
	$(error VERSION is required, usage: make migrate-force VERSION=1)
endif
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(MYSQL_DSN)" force $(VERSION)
