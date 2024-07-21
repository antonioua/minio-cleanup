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
	generateCmd.Flags().IntP("files", "n", 0, "Number of files to generate")
	generateCmd.Flags().IntP("threads", "t", 1, "Number of threads")
}

func generateFiles(cmd *cobra.Command, args []string) {
	bucketName, _ := cmd.Flags().GetString("bucket")
	numFiles, _ := cmd.Flags().GetInt("files")
	numThreads, _ := cmd.Flags().GetInt("threads")

	fmt.Println("Running generateFiles for: ", bucketName, numFiles, numThreads)

	//ctx, cancel := context.WithCancel(context.Background())
	//defer cancel()

	minioClient, err := minio.New("localhost:8888", &minio.Options{
		Creds:  credentials.NewStaticV4("minioconsole", "minioconsole123", ""),
		Secure: false,
	})

	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	objectNames := make([]string, numFiles)
	for i := 0; i < numFiles; i++ {
		objectNames = append(objectNames, fmt.Sprintf("inbox/notify_%s.json", uuid.New().String()))
	}

	wg := sync.WaitGroup{}
	wg.Add(numFiles)
	for _, objectName := range objectNames {
		go func(objectName string) {
			content := []byte(objectName)
			_, err = minioClient.PutObject(ctx, bucketName, objectName, bytes.NewReader(content), int64(len(content)), minio.PutObjectOptions{ContentType: "application/json"})
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Successfully uploaded: ", objectName)
		}(objectName)
	}

	wg.Wait()
}
