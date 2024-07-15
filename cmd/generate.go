package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate number of files in specific bucket",
	Long: `A longer description that spans multiple lines and likely contains
For example:

minioCleanupBuckets generate -b <bucket> -n 10 -t 1`,

	Run: generateFiles,
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringP("bucket", "b", "a", "Bucket name")
	generateCmd.Flags().IntP("files", "n", 1, "Number of files to generate")
	generateCmd.Flags().IntP("threads", "t", 1, "Number of threads")
}

func generateFiles(cmd *cobra.Command, args []string) {
	bucket, _ := cmd.Flags().GetString("bucket")
	numFiles, _ := cmd.Flags().GetInt("files")
	numThreads, _ := cmd.Flags().GetInt("threads")

	fmt.Println("Running generateFiles for: ", bucket, numFiles, numThreads)
}
