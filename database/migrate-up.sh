#!/bin/bash

set -ex

cd $(dirname $0)

export MYSQL_ROOT_PASSWORD=${MYSQL_ROOT_PASSWORD}

go get -v github.com/rubenv/sql-migrate/...

sed -i "s/<MYSQL_ROOT_PASSWORD>/$MYSQL_ROOT_PASSWORD/g" migrate.yml

sql-migrate up -config=migrate.yml -env mysql