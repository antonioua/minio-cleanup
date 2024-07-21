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

var cleanFilesCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean up files in specific bucket",
	Long: `A longer description that spans multiple lines and likely contains
For example:

./minioCleanupBuckets clean -b smp-to-oss-sandbox -n 100 -t 1`,

	//Run: func(cmd *cobra.Command, args []string) {
	//	if len(args) < 2 {
	//		cmd.Help()
	//		os.Exit(0)
	//	}
	//	cleanupFiles(cmd, args)
	//},
	Run: cleanupFiles,
}

func init() {
	rootCmd.AddCommand(cleanFilesCmd)
	cleanFilesCmd.Flags().StringP("older-than", "o", "", "Filter files older than duration (e.g., '5d', '1h', '30m', '45s', '2d3h4m')")
	cleanFilesCmd.Flags().StringP("prefix", "p", "", "Filter files with specific prefix (e.g., 'inbox')")
	cleanFilesCmd.Flags().StringP("suffix", "s", "", "Filter files with specific suffix (e.g., '.json')")
	cleanFilesCmd.Flags().StringP("bucket", "b", "", "Bucket name")
	//cleanFilesCmd.Flags().IntP("threads", "t", 1, "Number of threads")
}

func cleanupFiles(cmd *cobra.Command, args []string) {
	olderThanStr, _ := cmd.Flags().GetString("older-than")
	prefix, _ := cmd.Flags().GetString("prefix")
	suffix, _ := cmd.Flags().GetString("suffix")
	bucketName, _ := cmd.Flags().GetString("bucket")
	//numThreads, _ := cmd.Flags().GetInt("threads")

	olderThanDuration, err := longduration.ParseDuration(olderThanStr)
	if err != nil {
		log.Fatalf("Invalid older-than duration format: %v", err)
	}

	fmt.Println("Running cleanupFiles...")

	minioClient, err := minio.New("localhost:8888", &minio.Options{
		Creds:  credentials.NewStaticV4("minioconsole", "minioconsole123", ""),
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

	//minioClient.RemoveObjects
}
