package cmd

import (
	"bytes"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/spf13/cobra"
	"log"
	"sync"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate number of files in specific bucket",
	Long: `A longer description that spans multiple lines and likely contains
For example:

./minioCleanupBuckets generate -b smp-to-oss-sandbox -n 100 -t 1`,

	Run: generateFiles,
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringP("bucket", "b", "", "Bucket name")
	generateCmd.Flags().StringP("prefix", "p", "", "Filter files with specific prefix (e.g., 'inbox')")
	generateCmd.Flags().IntP("files-number", "n", 0, "Number of files to generate")
	generateCmd.Flags().IntP("workers", "w", 1, "Number of workers, a.k.a. number of concurrent requests")
}

type Job struct {
	BucketName string
	ObjectName string
}

func worker(id int, minioClient *minio.Client, ctx context.Context, jobs <-chan Job, results chan<- string, wg *sync.WaitGroup) {
	content := []byte("Hello world!")

	for job := range jobs {
		fmt.Println("worker: ", id, " has started the job:  ", job)
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

	fmt.Println("Running generateFiles...")

	minioClient, err := minio.New("localhost:8888", &minio.Options{
		Creds:  credentials.NewStaticV4("minioconsole", "minioconsole123", ""),
		Secure: false,
	})

	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	objectNames := make([]string, 0, numFiles)
	for i := 0; i < numFiles; i++ {
		objectNames = append(objectNames, fmt.Sprintf("%s/notify_%s.json", prefix, uuid.New().String()))
	}

	jobs := make(chan Job, len(objectNames))
	results := make(chan string, len(objectNames))
	wg := sync.WaitGroup{}

	// Start workers
	for w := 0; w < numWorkers; w++ {
		go worker(w, minioClient, ctx, jobs, results, &wg)
	}

	// Sending jobs to worker pool
	for _, objectName := range objectNames {
		jobs <- Job{ObjectName: objectName, BucketName: bucketName}
		wg.Add(1)
	}
	close(jobs)
	wg.Wait()

	// Collecting results
	for i := 0; i < len(objectNames); i++ {
		fmt.Println("Successfully uploaded: ", <-results)
	}

	fmt.Println("\nDone.")
}
