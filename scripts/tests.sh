#!/bin/bash

set -e
export TEST_DATABASE_URL="postgres://test_user:test_password@localhost:5433/test_db?sslmode=disable"

docker-compose -f ./docker/testbd-docker-compose.yml up -d
echo "Running tests..."
go test ./... -count=1 -v
docker-compose -f ./docker/testbd-docker-compose.yml down
