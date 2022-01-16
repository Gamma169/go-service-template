#!/bin/bash
echo 'Stopping postgres'
docker stop foobar_post
echo 'Removing foobar_post if there'
docker rm foobar_post
echo 'Starting postgres docker database'
docker run -d --name=foobar_post -e POSTGRES_HOST_AUTH_METHOD=trust -e POSTGRES_DB=foo -p 5432:5432 postgres:9.6.17-alpine
echo 'Sleeping a bit for database to start'
sleep 6

echo 'Building Server'
./scripts/build.sh foobar
echo 'Starting server'
DATABASE_NAME=foo DATABASE_USER=postgres DATABASE_HOST=127.0.0.1 RUN_MIGRATIONS="true" ./bin/foobar
