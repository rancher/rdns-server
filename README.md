rdns-server
========

[![Build Status](https://drone-publish.rancher.io/api/badges/rancher/rdns-server/status.svg)](https://drone-publish.rancher.io/rancher/rdns-server)
[![Go Report Card](https://goreportcard.com/badge/github.com/rancher/rdns-server)](https://goreportcard.com/report/github.com/rancher/rdns-server)

The `rdns-server` implements the API interface of Dynamic DNS, its goal is to use a variety of DNS servers such as Route53, CoreDNS etc.
Now `rdns-server` only supports `A/TXT` records, other record types will be added as soon as possible.

| Default | Backend | Description |
| ------- | ------- | ----------- |
|    *    | Route53 | Store the records in the AWS Route53 service and copy them to the database |
|         | Etcdv3 | Store the records in the ETCD and query by CoreDNS |

## Latest Release
* Latest - v0.5.1 - `rancher/rdns-server:v0.5.1-rancher-amd64`.

## Building

`make`

## Running
Different environment variables need to be set for different backend before running.

See [here](https://github.com/rancher/rdns-server/blob/master/doc/usages.md) for the environment variables you can set.

#### Running route53 backend
```
./scripts/start route53
```

#### Running etcdv3 backend
This backend will launches the CoreDNS service by default and users no need to run additional CoreDNS.

```
./scripts/start etcdv3
```

> It's need to modify the configuration files in the `deploy/etcdv3/config` directory if not using the `default` domain.

#### Migrate Datum From v0.4.x To v0.5.x
Now supports migration from the `v0.4.x` data to the new `v0.5.x` data store (etcdv3, route53). 

Please see [here](https://github.com/Jason-ZW/rdns-migrate-tools#rdns-migrate-tools) for details.

## Testing
Now we only add integration tests, others will coming soon.

Please see [here](https://github.com/rancher/rdns-server/tree/master/tests/integration) for details.

## Monitoring
Now provides prometheus metrics data at `/metrics` endpoints.

## API References
Please see [here](https://github.com/rancher/rdns-server/blob/master/doc/apis.md) for details.

## Usages
Please see [here](https://github.com/rancher/rdns-server/blob/master/doc/usages.md) for details.

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
