rdns-server
========

The rdns-server implements the API interface of Dynamic DNS, its goal is to use a variety of DNS servers such as Route53, CoreDNS etc...

| Default | Backend | Description |
| ------- | ------- | ----------- |
|    *    | Route53 | Store the records in the AWS Route53 service and copy them to the database |
|         | CoreDNS | Store the records in the ETCD and query by CoreDNS |

## Building

`make`

## Running

#### Running RDNS Route53
```shell
MYSQL_ROOT_PASSWORD="[password]" docker-compose -f deploy/route53/mysql-compose.yaml up -d
MYSQL_ROOT_PASSWORD="[password]" database/migrate-up.sh
AWS_HOSTED_ZONE_ID="[aws hosted zone ID]" AWS_ACCESS_KEY_ID="[aws access key ID]" AWS_SECRET_ACCESS_KEY="[aws secret access key]" DSN="[username[:password]@][tcp[(address)]]/rdns?parseTime=true" docker-compose -f deploy/route53/rdns-compose.yaml up -d
```

#### Running RDNS CoreDNS
```shell
docker-compose -f deploy/etcdv3/etcd-compose.yaml up -d
ETCD_ENDPOINTS="http://127.0.0.1:2379" docker-compose -f deploy/etcdv3/rdns-compose.yaml up -d
```

#### Migrate Datum From v0.4.x To v0.5.x

[Guide](https://github.com/Jason-ZW/rdns-migrate-tools/blob/master/README.md)

## Usages

```
NAME:
   rdns-server - control and configure RDNS('2019-06-06T06:47:02Z')

USAGE:
   rdns-server [global options] command [command options] [arguments...]

VERSION:
   v0.5.1

AUTHOR:
   Rancher Labs, Inc.

COMMANDS:
     route53, r53  use aws route53 backend
     OPTIONS:
        --aws_hosted_zone_id value     used to set aws hosted zone ID. [$AWS_HOSTED_ZONE_ID]
        --aws_access_key_id value      used to set aws access key ID. [$AWS_ACCESS_KEY_ID]
        --aws_secret_access_key value  used to set aws secret access key. [$AWS_SECRET_ACCESS_KEY]
        --database value               used to set database. (default: "mysql") [$DATABASE]
        --dsn value                    used to set database dsn. [$DSN]
        --ttl value                    used to set record ttl. (default: "240h") [$TTL]
     etcdv3, ev3   use etcd-v3 backend
     OPTIONS:
        --core_dns_port value     used to set coredns port. (default: "53") [$CORE_DNS_PORT]
        --core_dns_cpu value      used to set coredns cpu, a number (e.g. 3) or a percent (e.g. 50%). (default: "50%") [$CORE_DNS_CPU]
        --ttl value               used to set record ttl. (default: "240h") [$TTL]
        --domain value            used to set etcd root domain. (default: "lb.rancher.cloud") [$DOMAIN]
        --etcd_endpoints value    used to set etcd endpoints. (default: "http://127.0.0.1:2379") [$ETCD_ENDPOINTS]
        --etcd_prefix_path value  used to set etcd prefix path. (default: "/rdnsv3") [$ETCD_PREFIX_PATH]
        --core_dns_file value     used to set coredns file. (default: "/etc/rdns/config/Corefile") [$CORE_DNS_FILE]

GLOBAL OPTIONS:
   --debug, -d     used to set debug mode. [$DEBUG]
   --listen value  used to set listen port. (default: ":9333") [$LISTEN]
   --frozen value  used to set the duration when the domain name can be used again. (default: "2160h") [$FROZEN]
   --version, -v   print the version
```

## API References

| <sub>API</sub> | <sub>Method</sub> | <sub>Header</sub> | <sub>Payload</sub> | <sub>Description</sub> |
| --- | ------ | ------ | ------- | ----------- |
| <sub>/v1/domain</sub> | <sub>POST</sub> | <sub>**Content-Type:** application/json <br/><br/> **Accept:** application/json</sub> | <sub>{"hosts": ["4.4.4.4", "2.2.2.2"], "subdomain": {"sub1": ["9.9.9.9","4.4.4.4"], "sub2": ["5.5.5.5","6.6.6.6"]}}</sub> | <sub>Create A Records</sub> |
| <sub>/v1/domain/&lt;FQDN&gt;</sub> | <sub>GET</sub> | <sub>**Content-Type:** application/json <br/><br/> **Accept:** application/json <br/><br/> **Authorization:** Bearer &lt;Token&gt;</sub> | <sub>-</sub> | <sub>Get A Records</sub> |
| <sub>/v1/domain/&lt;FQDN&gt;</sub> | <sub>PUT</sub> | <sub>**Content-Type:** application/json <br/><br/> **Accept:** application/json <br/><br/> **Authorization:** Bearer &lt;Token&gt;</sub> | <sub>{"hosts": ["4.4.4.4", "3.3.3.3"], "subdomain": {"sub1": ["9.9.9.9","4.4.4.4"], "sub3": ["5.5.5.5","6.6.6.6"]}}</sub> | <sub>Update A Records</sub> |
| <sub>/v1/domain/&lt;FQDN&gt;</sub> | <sub>DELETE</sub> | <sub>**Content-Type:** application/json <br/><br/> **Accept:** application/json <br/><br/> **Authorization:** Bearer &lt;Token&gt;</sub> | <sub>-</sub> | <sub>Delete A Records</sub> |
| <sub>/v1/domain/&lt;FQDN&gt;/txt</sub> | <sub>POST</sub> | <sub>**Content-Type:** application/json <br/><br/> **Accept:** application/json <br/><br/> **Authorization:** Bearer &lt;Token&gt;</sub> | <sub>{"text": "xxxxxx"}</sub> | <sub>Create TXT Record</sub> |
| <sub>/v1/domain/&lt;FQDN&gt;/txt</sub> | <sub>GET</sub> | <sub>**Content-Type:** application/json <br/><br/> **Accept:** application/json <br/><br/> **Authorization:** Bearer &lt;Token&gt;</sub> | <sub>-</sub> | <sub>Get TXT Record</sub> |
| <sub>/v1/domain/&lt;FQDN&gt;/txt</sub> | <sub>PUT</sub> | <sub>**Content-Type:** application/json <br/><br/> **Accept:** application/json <br/><br/> **Authorization:** Bearer &lt;Token&gt;</sub> | <sub>{"text": "xxxxxxxxx"}</sub> | <sub>Update TXT Record</sub> |
| <sub>/v1/domain/&lt;FQDN&gt;/txt</sub> | <sub>DELETE</sub> | <sub>**Content-Type:** application/json <br/><br/> **Accept:** application/json <br/><br/> **Authorization:** Bearer &lt;Token&gt;</sub> | <sub>-</sub> | <sub>Delete TXT Record</sub> |
| <sub>/v1/domain/&lt;FQDN&gt;/renew</sub> | <sub>PUT</sub> | <sub>**Content-Type:** application/json <br/><br/> **Accept:** application/json <br/><br/> **Authorization:** Bearer &lt;Token&gt;</sub> | <sub>-</sub> | <sub>Renew Records</sub> |
| <sub>/metrics</sub> | <sub>GET</sub> | <sub>-</sub> | <sub>-</sub> | <sub>Prometheus metrics</sub> |

## License
Copyright (c) 2014-2019 [Rancher Labs, Inc.](http://rancher.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
