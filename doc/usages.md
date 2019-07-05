# Usages

```
NAME:
   rdns-server - control and configure RDNS('2019-06-06T06:47:02Z')

USAGE:
   rdns-server [global options] command [command options] [arguments...]

VERSION:
   v0.5.2

AUTHOR:
   Rancher Labs, Inc.

COMMANDS:
     route53, r53  use aws route53 backend
     OPTIONS:
        --aws_hosted_zone_id value     used to set aws hosted zone ID. [$AWS_HOSTED_ZONE_ID]
        --aws_access_key_id value      used to set aws access key ID. [$AWS_ACCESS_KEY_ID]
        --aws_secret_access_key value  used to set aws secret access key. [$AWS_SECRET_ACCESS_KEY]
        --database value               used to set database. (default: "mysql") [$DATABASE]
        --database_lease_time value    used to set database lease time. (default: "240h") [$DATABASE_LEASE_TIME]
        --dsn value                    used to set database dsn. [$DSN]
        --ttl value                    used to set rout53 ttl. (default: "10") [$TTL]
     etcdv3, ev3   use etcd-v3 backend
     OPTIONS:
        --core_dns_port value           used to set coredns port. (default: "53") [$CORE_DNS_PORT]
        --core_dns_cpu value            used to set coredns cpu, a number (e.g. 3) or a percent (e.g. 50%). (default: "50%") [$CORE_DNS_CPU]
        --core_dns_db_file value        used to set coredns file plugin db's file (e.g. /etc/rdns/config/dbfile). [$CORE_DNS_DB_FILE_NAME]
        --core_dns_db_zone value        used to set coredns file plugin db's zone (e.g. api.lb.rancher.cloud). [$CORE_DNS_DB_ZONE]
        --ttl value                     used to set cordns ttl. (default: "60") [$TTL]
        --domain value                  used to set etcd root domain. (default: "lb.rancher.cloud") [$DOMAIN]
        --etcd_endpoints value          used to set etcd endpoints. (default: "http://127.0.0.1:2379") [$ETCD_ENDPOINTS]
        --etcd_prefix_path value        used to set etcd prefix path. (default: "/rdnsv3") [$ETCD_PREFIX_PATH]
        --etcd_lease_time value         used to set etcd lease time. (default: "240h") [$ETCD_LEASE_TIME]
        --core_dns_file value           used to set coredns file. (default: "/etc/rdns/config/Corefile") [$CORE_DNS_FILE]

GLOBAL OPTIONS:
   --debug, -d     used to set debug mode. [$DEBUG]
   --listen value  used to set listen port. (default: ":9333") [$LISTEN]
   --frozen value  used to set the duration when the domain name can be used again. (default: "2160h") [$FROZEN]
   --version, -v   print the version
```