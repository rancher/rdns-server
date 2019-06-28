## Backup

### Dependencies

* System Environment `MYSQL_ROOT_PASSWORD=xxx`

* AWS S3 bucket and the access and secret key with write permissions for this bucket.

### Backup data and manual upload to s3

1. Reference script `rdns-backup.sh`, and use crontab to perform, such as copy to `/etc/cron.daily/`.

```
mv backup_etcd_s3.sh /etc/cron.daily/rdns-backup.sh
```