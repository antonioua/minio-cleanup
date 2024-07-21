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

var listFilesCmd = &cobra.Command{
	Use:   "list",
	Short: "list  files in specific bucket",
	Long: `A longer description that spans multiple lines and likely contains
For example:

./miniolistBuckets list -b smp-to-oss-sandbox -n 100 -t 1`,

	Run: listFiles,
}

func init() {
	rootCmd.AddCommand(listFilesCmd)
	listFilesCmd.Flags().StringP("older-than", "o", "", "Filter files older than duration (e.g., '5d', '1h', '30m', '45s', '2d3h4m')")
	listFilesCmd.Flags().StringP("prefix", "p", "", "Filter files with specific prefix (e.g., 'inbox/')")
	listFilesCmd.Flags().StringP("suffix", "s", "", "Filter files with specific suffix (e.g., '.json')")
	listFilesCmd.Flags().StringP("bucket", "b", "", "Bucket name")
}

func listFiles(cmd *cobra.Command, args []string) {
	olderThanStr, _ := cmd.Flags().GetString("older-than")
	prefix, _ := cmd.Flags().GetString("prefix")
	suffix, _ := cmd.Flags().GetString("suffix")
	bucketName, _ := cmd.Flags().GetString("bucket")
	//numThreads, _ := cmd.Flags().GetInt("threads")

	olderThanDuration, err := longduration.ParseDuration(olderThanStr)
	if err != nil {
		log.Fatalf("Invalid older-than duration format: %v", err)
	}

	fmt.Println("Running listFiles...")

	minioClient, err := minio.New("localhost:8888", &minio.Options{
		Creds:  credentials.NewStaticV4("minioconsole", "minioconsole123", ""),
		Secure: false,
	})

	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	objectCh := minioClient.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	numOfObjects := 0

	currentTIme := time.Now()
	olderThanTime := currentTIme.Add(-olderThanDuration)

	for object := range objectCh {
		if object.Err != nil {
			fmt.Println(object.Err)
			return
		}

		if object.LastModified.Before(olderThanTime) && strings.HasSuffix(object.Key, suffix) {
			fmt.Println(object.Key)
			fmt.Println(object.LastModified)
			numOfObjects++
		}
	}

	fmt.Println("\nDone.")
	fmt.Println("Total number of found objects: ", numOfObjects)
}
