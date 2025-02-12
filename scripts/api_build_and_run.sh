#!/bin/bash
# Load environment variables from .env file
export $(grep -v '^#' docker/.env | xargs)

docker build -t graphql-api:latest -f docker/dockerfile-api . 
docker run -d --restart=unless-stopped -p 8080:8080 --name graphql-api \
  -e DB_HOST=db \
  -e DB_PORT=5432 \
  -e DB_USER=${DB_USER} \
  -e DB_PASSWORD=${DB_PASSWORD} \
  -e DB_NAME=${DB_NAME} \
  -e STORAGE_TYPE=${STORAGE_TYPE} \
  -e DATABASE_URL=postgres://${DB_USER}:${DB_PASSWORD}@db:5432/${DB_NAME} \
  graphql-api:latest