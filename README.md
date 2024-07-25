# minio-cleanup

Common CLI to help you cleanup files using filtering in Minio bucket

## Development
```bash
kubectl port-forward svc/minio -n anton-test 8888:80
kubectl get secret -n minio-operator console-sa-secret -o json | jq '.data.token' -r | base64 -d
kubectl port-forward svc/console -n minio-operator 9090:9090
```

```bash
go build -o minio_cleanup
```

## Usage
```bash
./minio_cleanup --help
```

## TODO
- Check if flags was set.
- Define required flags.
