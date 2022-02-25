#!/bin/bash
echo 'Stopping postgres'
docker stop foobar_post foobar_redis
echo 'Removing foobar_post if there'
docker rm foobar_post foobar_redis
echo 'Starting postgres docker database'
docker run -d --name=foobar_post -e POSTGRES_HOST_AUTH_METHOD=trust -e POSTGRES_DB=foo -p 5432:5432 postgres:9-alpine
echo 'Sleeping a bit for database to start'
sleep 2
echo 'Starting redis'
docker run -d --name=foobar_redis -p 6379:6379 redis:6-alpine
echo 'Sleeping a bit for redis to start'
sleep 4

# stop on error
set -e -o pipefail

echo 'Getting Dependencies'
./scripts/get_deps.sh
echo 'Building Server'
./scripts/build.sh foobar
echo 'Starting server'
REDIS_HOST=127.0.0.1 DATABASE_NAME=foo DATABASE_USER=postgres DATABASE_HOST=127.0.0.1 RUN_MIGRATIONS="true" BAZ_ID="some-id" ./bin/foobar
