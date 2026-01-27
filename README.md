## Secure Task Management API

A task management API built in Go. Supports JWT authentication, PostgreSQL, logging, and monitoring.

## Overview

The API separates handlers from database logic. PostgreSQL handles migrations, JWT manages authentication, passwords are hashed with bcrypt, Zap handles structured logging, and Sentry captures errors and panics.

## Tech Stack

Go 1.21+ with go-chi for routing

PostgreSQL 15 (pgx driver)

JWT auth (golang-jwt/jwt/v5)

Password hashing with bcrypt

Config management using Viper

Structured logging with zap

Error tracking with Sentry

Docker and docker-compose

## Project Structure
cmd/server/          # Entry point
configs/             # Config files (.yaml, .env)
deployments/         # Docker setup
docs/                # Postman collection
internal/            # auth, config, handlers, logger, models, repository
migrations/          # SQL migrations
pkg/utils/           # Shared helper functions

## Setup
Clone repository
git clone <repo-url>
cd secure-task-api

## Configure environment
cp configs/.env.example configs/.env


Example .env:

APP_ENV=development
APP_PORT=8080
POSTGRES_DB=goapi
POSTGRES_USER=postgres
POSTGRES_PASSWORD=<your password>
DATABASE_URL=postgres://postgres:<your password>@localhost:5432/taskdb?sslmode=disable
JWT_SECRET=<random secret>
LOG_LEVEL=debug
SENTRY_DSN=

## Running With Docker
cd deployments
docker-compose up -d
docker-compose logs -f app
docker-compose down

## Local
go mod download
cd migrations
migrate -path . -database "<db url>" up
go run cmd/server/main.go

## API Endpoints Authentication

POST /v1/auth/register – create user

POST /v1/auth/login – validate credentials, get JWT

POST /v1/auth/refresh – refresh JWT

# Tasks (JWT required)

GET /v1/tasks – list user tasks

POST /v1/tasks – create task

GET /v1/tasks/{id} – get task

PUT /v1/tasks/{id} – update task

DELETE /v1/tasks/{id} – delete task

# System

GET /health – check DB connection

GET /debug/panic – trigger panic for testing

## Migrations
migrate create -ext sql -dir migrations -seq <name>
migrate -path migrations -database "<db url>" up
migrate -path migrations -database "<db url>" down 1

## Notes
Passwords are always hashed with bcrypt
JWT middleware protects task routes
Repository pattern keeps SQL out of handlers
Zap logs requests with request IDs
Panics are logged and sent to Sentry
All config is loaded via Viper
No secrets are stored in the repo