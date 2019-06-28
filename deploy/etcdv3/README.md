## Backup and Restore

### Dependencies

* etcdctl and [etcd-backup](https://github.com/giantswarm/etcd-backup)

* AWS S3 bucket and the access and secret key with write permissions for this bucket.

### Backup data to S3

1. Modify fields in `rdns-backup.sh`, such as `aws-s3-bucket`, `aws-s3-region`, etc.

2. Reference script `rdns-backup.sh`, and use crontab to perform, such as copy to `/etc/cron.daily/`.

```
mv backup_etcd_s3.sh /etc/cron.daily/rdns-backup.sh
```

### Restore data from S3

1. Make public the backup files.

2. In `rdns-restore.sh` script file,change the `BACKUP_V2_SUFFIX` and `BACKUP_V3_SUFFIX` fields according to the file you want to restore.

3. Execute `rdns-restore.sh` script file.