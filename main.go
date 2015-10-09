package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {
	client := s3.New(&aws.Config{
		Region: aws.String("us-west-2"),
	})

	params := &s3.ListObjectsInput{
		Bucket: aws.String("lattice-logs"),
		Prefix: aws.String("logs/"),
	}

	logFiles := []string{}
	err := client.ListObjectsPages(params, func(page *s3.ListObjectsOutput, lastPage bool) bool {
		for _, o := range page.Contents {
			logFiles = append(logFiles, *o.Key)
		}
		return true
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(logFiles)
}
