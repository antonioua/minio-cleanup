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
	generateCmd.Flags().StringP("prefix", "p", "", "Filter files with specific prefix (e.g., 'inbox')")
	generateCmd.Flags().StringP("bucket", "b", "", "Bucket name")
	generateCmd.Flags().IntP("files", "n", 0, "Number of files to generate")
	generateCmd.Flags().IntP("workers", "t", 1, "Number of workers, a.k.a. number of concurrent requests")
}

func generateFiles(cmd *cobra.Command, args []string) {
	prefix, _ := cmd.Flags().GetString("prefix")
	bucketName, _ := cmd.Flags().GetString("bucket")
	numFiles, _ := cmd.Flags().GetInt("files")
	workers, _ := cmd.Flags().GetInt("threads")

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

	wg := sync.WaitGroup{}
	wg.Add(numFiles)

	fmt.Println(objectNames)

	for _, objectName := range objectNames {

		go func(objectName string) {
			defer wg.Done()
			content := []byte("Hello world!")
			_, err = minioClient.PutObject(ctx, bucketName, objectName, bytes.NewReader(content), int64(len(content)), minio.PutObjectOptions{ContentType: "application/json"})
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Successfully uploaded: ", objectName)
		}(objectName)

	}

	wg.Wait()
	fmt.Println("\nDone.")
}
