#!/bin/bash

set -ex

cd $(dirname $0)

export MYSQL_ROOT_PASSWORD=${MYSQL_ROOT_PASSWORD}

mkdir -p /mnt/rdns/backup

apt-get update
apt-get install -y mysql-common mysql-client-core-5.7 mysql-client-5.7 libaio1 golang-docker-credential-helpers
rm -f /bin/sh && ln -s /bin/bash /bin/sh

mysqldump -uroot -p${MYSQL_ROOT_PASSWORD} --databases rdns > /mnt/rdns/backup/rdns_$(date +%Y%m%d_%H%M%S).sql