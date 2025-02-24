# minio-cleanup

This Command Line Interface (CLI) tool assists in cleaning up files in a MinIO bucket efficiently. Here are its main features:

• Multi-threaded Processing: Perform file cleanup operations concurrently to save time and increase efficiency.

• File Name Filtering: Utilize filters to specify which files to target for cleanup, ensuring precise control over the operation.

## Usage

Run as docker container.

```bash
docker run --rm xdesigns/minio-cleanup:latest --help
docker run --rm xdesigns/minio-cleanup:latest remove --timeout 10h --bucket <bucket_name> --older-than 2h --prefix / --suffix .json --workers 10 --host localhost:8888 --access-key <access_key> --secret-key <secret_key>
```

Run as k8s job.

```bash
cat <<EOF | kubectl apply -n <namespace> -f -
apiVersion: batch/v1
kind: Job
metadata:
  name: manual-minio-cleanup-job
spec:
  template:
    spec:
      containers:
      - name: minio-cleanup
        image: xdesigns/minio-cleanup:latest
        command: ["/minio_cleanup", "remove"]
        args:
          - "--bucket"
          - "<bucket_name>"
          - "--older-than"
          - "2h"
          - "--prefix"
          - "/"
          - "--suffix"
          - ".json"
          - "--workers"
          - "10"
          - "--host"
          - "minio.<namespace>.svc.cluster.local:80"
          - "--access-key"
          - "<minio_user>"
          - "--secret-key"
          - "<minio_pwd"
          - "--timeout"
          - "10h"
      restartPolicy: Never
  backoffLimit: 4
EOF

kubectl -n <namespace> logs job/manual-minio-cleanup-job -f

kubectl -n <namespace> delete job/manual-minio-cleanup-job
```

Compile and run.

```bash
go build -o minio_cleanup
./minio_cleanup --help
```

Download binary from the [releases](https://github.com/antonioua/minio-cleanup/releases) page and run.

## Expose MinIO and Console

```bash
kubectl port-forward svc/minio -n <namespace> 8888:80
kubectl get secret -n minio-operator console-sa-secret -o json | jq '.data.token' -r | base64 -d
kubectl port-forward svc/console -n minio-operator 9090:9090
```

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
