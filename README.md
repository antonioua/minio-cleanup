# minio-cleanup

Common CLI to help you cleanup files in bucket and using filtering for file names.

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
./minio_cleanup remove --bucket smp-to-oss-sandbox --older-than 10s --prefix inbox --suffix .json --workers 20 --host localhost:8888 --access-key <access_key> --secret-key <secret_key>`
```

[Container images.](https://hub.docker.com/r/xdesigns/minio-cleanup/tags)

## TODO
- [x] Check if flags was set.
- [x] Define required flags.
- [ ] Fix binary name in release to be minio_cleanup and not minio-cleanup. Also check help to match it.
- [ ] Move all commands from readme to Makefile.
- [ ] Fix GoReleaser Github Action.
- [ ] Remove printing removal of each file and generating to speed up the application.
- [ ] Remove hardcoded size of results channel for removing.
- [ ] Print example if cmd was chosen but flags bot set.
- [ ] Fix number of "Removed objects", now it's shows as doubled.
