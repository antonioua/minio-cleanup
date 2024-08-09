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

./minio_cleanup list --bucket smp-to-oss-sandbox --older-than 10s --prefix inbox --suffix .json --host localhost:8888 --access-key <access_key> --secret-key <secret_key>

Note: for Windows use minio_cleanup.exe`,

	Run: listFiles,
}

func init() {
	rootCmd.AddCommand(listFilesCmd)
	listFilesCmd.Flags().StringP("older-than", "o", "", "Filter files older than duration (e.g., '5d', '1h', '30m', '45s', '2d3h4m')")
	listFilesCmd.Flags().StringP("prefix", "p", "", "Filter files with specific prefix (e.g., 'inbox')")
	listFilesCmd.Flags().StringP("suffix", "s", "", "Filter files with specific suffix (e.g., '.json')")
	listFilesCmd.Flags().StringP("bucket", "b", "", "Bucket name")
	listFilesCmd.Flags().StringP("host", "", "localhost:8888", "Minio host:port")
	listFilesCmd.Flags().StringP("access-key", "", "", "Minio access key")
	listFilesCmd.Flags().StringP("secret-key", "", "", "Minio secret key")

	if err := listFilesCmd.MarkFlagRequired("bucket"); err != nil {
		log.Fatal(err)
	}
	if err := listFilesCmd.MarkFlagRequired("prefix"); err != nil {
		log.Fatal(err)
	}
	if err := listFilesCmd.MarkFlagRequired("access-key"); err != nil {
		log.Fatal(err)
	}
	if err := listFilesCmd.MarkFlagRequired("secret-key"); err != nil {
		log.Fatal(err)
	}
}

func listFiles(cmd *cobra.Command, args []string) {
	olderThanStr, _ := cmd.Flags().GetString("older-than")
	prefix, _ := cmd.Flags().GetString("prefix")
	suffix, _ := cmd.Flags().GetString("suffix")
	bucketName, _ := cmd.Flags().GetString("bucket")
	host, _ := cmd.Flags().GetString("host")
	accessKey, _ := cmd.Flags().GetString("access-key")
	secretKey, _ := cmd.Flags().GetString("secret-key")

	olderThanDuration, err := longduration.ParseDuration(olderThanStr)
	if err != nil {
		log.Fatalf("Invalid older-than duration format: %v", err)
	}

	fmt.Println("Running listFiles...")

	minioClient, err := minio.New(host, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
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
			//fmt.Println(object.Key)
			//fmt.Println(object.LastModified)
			numOfObjects++
		}
	}

	fmt.Println("\nDone.")
	fmt.Println("Total number of found objects: ", numOfObjects)
}
