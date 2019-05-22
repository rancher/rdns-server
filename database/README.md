# Database

Using [SQL schema migration tool](https://github.com/rubenv/sql-migrate) to migrate database.

| Default | Backend | Description |
| ------- | ------- | ----------- |
|    *    |  MYSQL  | Store data for management |

## Init Database

```shell
MYSQL_ROOT_PASSWORD=xxx ./migrate-up.sh
```

## Backup Database

### Configure cron

```shell
apt-get update && apt-get install -y cron

export MYSQL_ROOT_PASSWORD=xxx

crontab -e
0 4 * * 0 /usr/bin/rdns-backup.sh

service cron restart
service cron reload
```