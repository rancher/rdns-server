# Integration Test

## Build

```
docker build -t "rancher/rdns-integration-test" .
```

## Run

```
docker run --name rdns-integration-test --rm --net=host -e ENV_RDNS_ENDPOINT="http://127.0.0.1:9333/v1" rancher/rdns-integration-test
```