# minio-cleanup

Common CLI to help you cleanup files using filtering in Minio bucket

## Development
```bash
kubectl port-forward svc/minio -n anton-test 8888:80
kubectl get secret -n minio-operator console-sa-secret -o json | jq '.data.token' -r | base64 -d
kubectl port-forward svc/console -n minio-operator 9090:9090
```

```bash
cd /Users/antonk/Documents/projects/my-projects/minio-cleanup
go build -o minio_cleanup
```

## Usage
```bash
./minio_cleanup generate --bucket "smp-to-oss-sandbox" --prefix "inbox" --files-number 1000 --workers 20
./minio_cleanup list -b smp-to-oss-sandbox --older-than 10s -p "inbox" -s ".json" //TODO
./minio_cleanup clean  -b smp-to-oss-sandbox -o 1s -p "inbox" -s ".json" //TODO
```

## TODO
- Check if flags was set.
- Define required flags.
