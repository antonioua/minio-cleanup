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
	"time"
)

var removeFilesCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove up files in specific bucket",
	Long: `A longer description that spans multiple lines and likely contains
For example:

./minio_cleanup remove -b smp-to-oss-sandbox --older-than 10s --prefix inbox --suffix .json --host localhost:8888 --access-key <access_key> --secret-key <secret_key>`,

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
}

func removeFiles(cmd *cobra.Command, args []string) {
	olderThanStr, _ := cmd.Flags().GetString("older-than")
	prefix, _ := cmd.Flags().GetString("prefix")
	suffix, _ := cmd.Flags().GetString("suffix")
	bucketName, _ := cmd.Flags().GetString("bucket")
	//numWorkers, _ := cmd.Flags().GetInt("workers")
	host, _ := cmd.Flags().GetString("host")
	accessKey, _ := cmd.Flags().GetString("access-key")
	secretKey, _ := cmd.Flags().GetString("secret-key")

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

	ctx := context.Background()

	numOfObjects := 0
	currentTIme := time.Now()
	olderThanTime := currentTIme.Add(-olderThanDuration)

	objectsCh := make(chan minio.ObjectInfo)

	go func() {
		defer close(objectsCh)

		for object := range minioClient.ListObjects(ctx, bucketName, minio.ListObjectsOptions{Prefix: prefix, Recursive: true}) {
			if object.Err != nil {
				log.Fatalln(object.Err)
			}
			if object.LastModified.Before(olderThanTime) && strings.HasSuffix(object.Key, suffix) {
				fmt.Println("Removing: ", object.Key)
				objectsCh <- object
				numOfObjects++
			}
		}
	}()

	for rErr := range minioClient.RemoveObjects(ctx, bucketName, objectsCh, minio.RemoveObjectsOptions{GovernanceBypass: true}) {
		fmt.Println("Error detected during deletion: ", rErr)
	}

	fmt.Println("\nDone.")
	fmt.Println("Total number of removed objects: ", numOfObjects)
}
