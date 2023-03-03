# aws-redis-simple

A very simple container that serves 2 purposes:

* demonstrate golang as an aws elasticache client
* validate elasticache secure config / connectivity

On startup the container attempts to set/get a key before configuring the /readiness endpoint in the gin router, so the pod will never become healthy if elasticache isn't working.  The /liveness endpoint performs the set/get on each request.

# tls

TLS is available in aws and uses certs created by a public AWS CA (unlike RDS which is a special non-public bundle) - so public bundles from any distro should work.  The [Dockerfile](./Dockerfile) in this repo uses certs from `alpine:latest`.

**TLS on mac >= 12.x** - the AWS certs do not comply with SCT verification - this requires TLS verification be disabled (the connection is still TLS, just without cert verification).  A darwin check has been hard-coded into this example [here](https://github.com/bensolo-io/aws-redis-simple/blob/7a3e33dbf4df8436342961c21544c1a12e155967/main.go#L51-L56).

# requirements

An elasticache redis instance with:

* in clustered mode (required to enable redis auth in aws)
* auth token configured
* tls in transist
* network connectivity from client pod (security groups, subnet placement, etc.)

This [terraform example](https://github.com/bensolo-io/cloud-gitops-examples/tree/main/terraform/redis-1region) creates a vpc, redis cluster, and eks cluster for testing.

# run as a k8s pod

**deploy**

Create a secret containing elasticache redis endopint and auth token:

```bash
kubectl apply --context $MGMT_CONTEXT -f- <<EOF
kind: Secret
apiVersion: v1
metadata:
  name: redis-config
  namespace: default
stringData:
  token: ${REDIS_PASSWORD}
  host: ${REDIS_HOST}
  port: ${REDIS_PORT}
EOF
```
Deploy as a pod (attempts to set and get a key each time readiness or liveness is requested; pod will not be healthy if anything fails):

```bash
kubectl apply -k ./deploy/kustomize --context $MGMT_CONTEXT
```

**use redis-cli**

The tester pod is based on >>> which contains several network testing tools.  It also has the `redis-cli` installed.

```bash
# auth token is set in REDISCLI_AUTH which is picked up by redis-cli
kubectl exec -it $(kubectl get pods -n default -l app=redis-tester -oname) -n default -- bash -c '/usr/bin/redis-cli -h ${REDIS_HOST} -p ${REDIS_POST} --tls -n 0'
```

# run as docker image

```bash
echo "REDIS_HOST=${REDIS_HOST}" > .redis-config.txt
echo "REDIS_PASSWORD=${REDIS_PASSWORD}" >> .redis-config.txt
echo "LOG_LEVEL=debug" >> .redis-config.txt

docker run -p 8080:8080 --env-file .redis-config.txt kodacd/aws-elasticache-redis-tester:latest
```

check for the log output "initial redis check OK" (you can also curl the /liveliness endpoint):

```
{"level":"info","time":"2023-02-27T14:27:05Z","message":"initial redis check OK"}
```

# run as binary

```bash
REDIS_ADDR="" REDIS_PASSWORD="" LOG_LEVEL="debug" go run main.go
```

# config options

```go
type Config struct {
	Port                    int    `env:"PORT" envDefault:"8080"`
	LogLevel                string `env:"LOG_LEVEL" envDefault:"info"`
	LogNoColor              bool   `env:"LOG_NO_COLOR" envDefault:"false"`
	RedisAddr               string `env:"REDIS_ADDR,required"`
	RedisDbIndex            int    `env:"REDIS_DB_INDEX" envDefault:"0"`
	RedisPassword           string `env:"REDIS_PASSWORD,required"`
	RedisTestKeyName        string `env:"REDIS_TEST_KEY_NAME" envDefault:"local"`
	RedisInsecureSkipVerify bool   `env:"REDIS_INSECURE_SKIP_VERIFY" envDefault:"false"`
}
```