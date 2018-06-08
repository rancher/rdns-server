# Installation

## How to use

```
# This will build a single node etcd
# It is recommended to use etcd HA
$ docker-compose -f ./etcd-compose.yaml up -d

# This will run coredns and rdns-server
$ docker-compose -f ./rdns-compose.yaml up -d
```

### Before running

If you want to use other root domain, you can update the `corefile`. This demo uses `lb.rancher.cloud`.

According to the root domain, create the corresponding path in etcd:

```
$ etcdctl mkdir /rdns/lb/rancher/cloud
```

### Add test records

```
$ etcdctl set /rdns/lb/rancher/cloud/zhibo/x1 '{"host":"1.1.1.1"}'
{"host":"1.1.1.1"}
$ etcdctl set /rdns/lb/rancher/cloud/zhibo/x2 '{"host":"2.2.2.2"}'
{"host":"2.2.2.2"}
```

Use dig for test:

```
dig @<coredns-address> zhibo.lb.rancher.cloud A

; <<>> DiG 9.8.3-P1 <<>> @172.22.100.5 zhibo.lb.rancher.cloud A
; (1 server found)
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 7723
;; flags: qr aa rd ra; QUERY: 1, ANSWER: 2, AUTHORITY: 0, ADDITIONAL: 0

;; QUESTION SECTION:
;zhibo.lb.rancher.cloud.	IN	A

;; ANSWER SECTION:
zhibo.lb.rancher.cloud. 160	IN	A	2.2.2.2
zhibo.lb.rancher.cloud. 160	IN	A	1.1.1.1
```

## Backup and Restore

### Dependencies

* etcdctl and [etcd-backup](https://github.com/giantswarm/etcd-backup)

* AWS S3 bucket and the access and secret key with write permissions for this bucket.

### Backup data to S3

1. Modify fields in `backup_etcd_s3.sh`, such as `s3-bucket`, `s3-region`, etc.

2. Reference script `backup_etcd_s3.sh`, and use crontab to perform, such as copy to `/etc/cron.daily/`.

```
mv backup_etcd_s3.sh /etc/cron.daily/backup_etcd_s3
```

### Restore data from S3

1. Make public the backup files.

2. In `restore_etcd.sh` script file,change the `BACKUP_V2_SUFFIX` and `BACKUP_V3_SUFFIX` fields according to the file you want to restore.

3. Execute `restore_etcd.sh` script file.
