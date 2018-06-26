# Rdns-server tests

## Integration test

### Requirement

>Rdns-server,coreDNS and etcd have already been successfully deployed.


### Build image
```
./build-image.sh
```

### Run the test
```
docker run --name test --rm -e ENV_HOST="172.0.0.1" wchao241/rdns-server-test:latest
```
>`ENV_HOST` is rdns-server host IP