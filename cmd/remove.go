package cmd

import (
	"context"
	"fmt"
	"github.com/mashiike/longduration"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/spf13/cobra"
	"log"
	"strings"
	"sync"
	"time"
)

var removeFilesCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove up files in specific bucket",
	Long: `A longer description that spans multiple lines and likely contains
For example:

./minio_cleanup remove --bucket smp-to-oss-sandbox --older-than 10s --prefix inbox --suffix .json --workers 20 --host localhost:8888 --access-key <access_key> --secret-key <secret_key>`,

	Run: removeFiles,
}

func init() {
	rootCmd.AddCommand(removeFilesCmd)
	removeFilesCmd.Flags().StringP("older-than", "o", "", "Filter files older than duration (e.g., '5d', '1h', '30m', '45s', '2d3h4m')")
	removeFilesCmd.Flags().StringP("prefix", "p", "", "Filter files with specific prefix (e.g., 'inbox')")
	removeFilesCmd.Flags().StringP("suffix", "s", "", "Filter files with specific suffix (e.g., '.json')")
	removeFilesCmd.Flags().StringP("bucket", "b", "", "Bucket name")
	removeFilesCmd.Flags().IntP("workers", "w", 1, "Number of workers, a.k.a. number of concurrent requests")
	removeFilesCmd.Flags().StringP("host", "", "localhost:8888", "Minio host:port")
	removeFilesCmd.Flags().StringP("access-key", "", "", "Minio access key")
	removeFilesCmd.Flags().StringP("secret-key", "", "", "Minio secret key")
	removeFilesCmd.Flags().StringP("timeout", "", "10m", "Timeout duration (e.g., '5m', '10s', '1h')")

	removeFilesCmd.MarkFlagRequired("older-than")
	removeFilesCmd.MarkFlagRequired("prefix")
	removeFilesCmd.MarkFlagRequired("suffix")
	removeFilesCmd.MarkFlagRequired("bucket")
	removeFilesCmd.MarkFlagRequired("access-key")
	removeFilesCmd.MarkFlagRequired("secret-key")
}

func removalWorker(id int, minioClient *minio.Client, ctx context.Context, jobs <-chan Job, results chan<- string, wg *sync.WaitGroup) {
	for job := range jobs {
		//fmt.Println("worker: ", id, " has started the job:  ", job)
		err := minioClient.RemoveObject(ctx, job.BucketName, job.ObjectName, minio.RemoveObjectOptions{GovernanceBypass: true})
		if err != nil {
			results <- err.Error()
			log.Fatal(err)
		}
		results <- job.ObjectName
		wg.Done()
	}
}

func removeFiles(cmd *cobra.Command, args []string) {
	olderThanStr, _ := cmd.Flags().GetString("older-than")
	prefix, _ := cmd.Flags().GetString("prefix")
	suffix, _ := cmd.Flags().GetString("suffix")
	bucketName, _ := cmd.Flags().GetString("bucket")
	numWorkers, _ := cmd.Flags().GetInt("workers")
	host, _ := cmd.Flags().GetString("host")
	accessKey, _ := cmd.Flags().GetString("access-key")
	secretKey, _ := cmd.Flags().GetString("secret-key")
	timeout, _ := cmd.Flags().GetString("timeout")

	olderThanDuration, err := longduration.ParseDuration(olderThanStr)
	if err != nil {
		log.Fatalf("Invalid older-than duration format: %v", err)
	}

	fmt.Println("Running removeFiles...")

	minioClient, err := minio.New(host, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})

	if err != nil {
		log.Fatal(err)
	}

	timeoutDuration, err := longduration.ParseDuration(timeout)
	if err != nil {
		log.Fatalf("Invalid timeout duration format: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
	defer cancel()

	numOfObjects := 0
	currentTIme := time.Now()
	olderThanTime := currentTIme.Add(-olderThanDuration)

	jobs := make(chan Job, 10000)
	results := make(chan string, 10000)

	wg := sync.WaitGroup{}

	// Start workers
	for w := 1; w <= numWorkers; w++ {
		go removalWorker(w, minioClient, ctx, jobs, results, &wg)
		//fmt.Println("Worker ", w, " started.")
	}

	// Sending jobs to worker pool
	go func() {
		defer close(jobs)
		for object := range minioClient.ListObjects(ctx, bucketName, minio.ListObjectsOptions{Prefix: prefix, Recursive: true}) {
			if object.Err != nil {
				log.Fatalln(object.Err)
			}

			//fmt.Println("Removing: ", object.Key)
			results <- object.Key

			if object.LastModified.Before(olderThanTime) && strings.HasSuffix(object.Key, suffix) {
				jobs <- Job{ObjectName: object.Key, BucketName: bucketName}
				wg.Add(1)
				numOfObjects++
			}
		}
		wg.Wait()
		close(results)
	}()

	// Collecting results
	numOfRemovedObjects := 0

	for {
		select {
		case result, ok := <-results:
			if ok {
				numOfRemovedObjects++
				fmt.Println("Successfully removed: ", result)
				//fmt.Println("Removed objects:", numOfRemovedObjects)
			} else {
				fmt.Println("No more results to process.")
				fmt.Println("\nDone.")
				fmt.Println("Took time: ", time.Since(currentTIme))
				return
			}
		case <-ctx.Done():
			fmt.Println("Timeout reached, stopping...")
			return
		}
	}
}
