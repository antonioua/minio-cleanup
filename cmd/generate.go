package cmd

import (
	"bytes"
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/spf13/cobra"
	"log"
	"sync"
	"time"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate number of files in specific bucket",
	Long: `A longer description that spans multiple lines and likely contains
For example:

./minio_cleanup generate --bucket smp-to-oss-sandbox --prefix inbox --files-number 1000 --workers 20 --host localhost:8888 --access-key <access_key> --secret-key <secret_key>`,

	Run: generateFiles,
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringP("bucket", "b", "", "Bucket name")
	generateCmd.Flags().StringP("prefix", "p", "", "Filter files with specific prefix (e.g., 'inbox')")
	generateCmd.Flags().IntP("files-number", "n", 0, "Number of files to generate")
	generateCmd.Flags().IntP("workers", "w", 1, "Number of workers, a.k.a. number of concurrent requests")
	generateCmd.Flags().StringP("host", "", "localhost:8888", "Minio host:port")
	generateCmd.Flags().StringP("access-key", "", "", "Minio access key")
	generateCmd.Flags().StringP("secret-key", "", "", "Minio secret key")

	generateCmd.MarkFlagRequired("bucket")
	generateCmd.MarkFlagRequired("prefix")
	generateCmd.MarkFlagRequired("files-number")
	generateCmd.MarkFlagRequired("access-key")
	generateCmd.MarkFlagRequired("secret-key")
}

type Job struct {
	BucketName string
	ObjectName string
}

func uploaderWorker(id int, minioClient *minio.Client, ctx context.Context, jobs <-chan Job, results chan<- string, wg *sync.WaitGroup) {
	content := []byte("Hello world!")

	for job := range jobs {
		//fmt.Println("worker: ", id, " has started the job:  ", job)
		_, err := minioClient.PutObject(ctx, job.BucketName, job.ObjectName, bytes.NewReader(content), int64(len(content)), minio.PutObjectOptions{ContentType: "application/json"})
		if err != nil {
			results <- err.Error()
			log.Fatal(err)
		}
		results <- job.ObjectName
		wg.Done()
	}
}

func generateFiles(cmd *cobra.Command, args []string) {
	bucketName, _ := cmd.Flags().GetString("bucket")
	prefix, _ := cmd.Flags().GetString("prefix")
	numFiles, _ := cmd.Flags().GetInt("files-number")
	numWorkers, _ := cmd.Flags().GetInt("workers")
	host, _ := cmd.Flags().GetString("host")
	accessKey, _ := cmd.Flags().GetString("access-key")
	secretKey, _ := cmd.Flags().GetString("secret-key")

	fmt.Println("Running generateFiles...")

	currentTIme := time.Now()

	minioClient, err := minio.New(host, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})

	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Generate object names
	objectNames := make([]string, 0, numFiles)
	for i := 0; i < numFiles; i++ {
		objectNames = append(objectNames, fmt.Sprintf("%s/notify_%s.json", prefix, time.Now().UnixNano()))
	}

	jobs := make(chan Job, len(objectNames))
	results := make(chan string, len(objectNames))
	defer close(results)
	wg := sync.WaitGroup{}

	// Start workers
	for w := 1; w <= numWorkers; w++ {
		go uploaderWorker(w, minioClient, ctx, jobs, results, &wg)
		//fmt.Println("Worker ", w, " started.")
	}

	// Sending jobs to worker pool
	go func() {
		defer close(jobs)
		for _, objectName := range objectNames {
			jobs <- Job{ObjectName: objectName, BucketName: bucketName}
			wg.Add(1)
		}
		wg.Wait()
	}()

	// Collecting results
	for i := 0; i < len(objectNames); i++ {
		//<-results
		fmt.Println("Successfully uploaded: ", <-results)
	}

	fmt.Println("\nDone.")
	fmt.Println("Took time: ", time.Since(currentTIme))
}
