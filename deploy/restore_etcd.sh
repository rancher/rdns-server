#!/bin/bash

#backup flag
#example
#https://s3.us-east-2.amazonaws.com/rdns-etcd-bp/rdnsdb-etcd-backup-v2-2018-05-17T08-58-28.tar.gz
#https://s3.us-east-2.amazonaws.com/rdns-etcd-bp/rdnsdb-etcd-backup-v3-2018-05-17T08-58-29.db.tar.gz
#BACKUP_V2_SUFFIX=2018-05-17T08-58-28
#BACKUP_V3_SUFFIX=2018-05-17T08-58-29

BACKUP_V2_SUFFIX=XXX
BACKUP_V3_SUFFIX=YYY

BACKUP_PREFIX=rdnsdb-etcd-backup
ETCD_PATH=/mnt/data/etcd/

BACKUP_V2=${BACKUP_PREFIX}-v2-${BACKUP_V2_SUFFIX}
BACKUP_V3=${BACKUP_PREFIX}-v3-${BACKUP_V3_SUFFIX}

curl -sL https://s3.us-east-2.amazonaws.com/rdns-etcd-bp/${BACKUP_V2}.tar.gz -o /tmp/${BACKUP_V2}.tar.gz
curl -sL https://s3.us-east-2.amazonaws.com/rdns-etcd-bp/${BACKUP_V3}.db.tar.gz -o /tmp/${BACKUP_V3}.db.tar.gz

tar xf /tmp/${BACKUP_V2}.tar.gz -C /tmp
tar xf /tmp/${BACKUP_V3}.db.tar.gz -C /tmp

cp -r /tmp/${BACKUP_V2}/* ${ETCD_PATH}
cp /tmp/${BACKUP_V3}.db ${ETCD_PATH}member/snap/db

docker-compose -f ./etcd-compose.yaml up -d
