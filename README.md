# aws-redis-simple

A very simple container that serves 2 purposes:

* demonstrate golang as an aws elasticache client
* validate elasticache secure config / connectivity

When deployed as a pod for environment validation, the /liveness endpoint attempts to set a key with the current timestamp, then get the key to make sure it was written.

# tls

TLS is available in aws and uses certs created by a public AWS CA (unlike RDS which is a special non-public bundle) - so any distro's recent public bundle should work.

**TLS on mac >= 12.x** - the AWS certs do not comply with SCT verification - this requires TLS verification be disabled (the connection is still TLS, just without cert verification).  A darwin check has been hard-coded into this example [here]().

# requirements

An elasticache redis instance with:

* in clustered mode (required to enable redis auth in aws)
* auth token configured
* tls in transist
* network connectivity from client pod (security groups, subnet placement, etc.)

# usage

Create a secret containing elasticache redis endopint and auth token:

```bash
kubectl apply --context $REMOTE_CONTEXT1 -f- <<EOF
kind: Secret
apiVersion: v1
metadata:
  name: redis-config
  namespace: default
stringData:
  token: ${REDIS_PASSWORD}
  address: ${REDIS_ADDR}
EOF

kubectl apply -k ./deploy/kustomize --context $REMOTE_CONTEXT1
```
