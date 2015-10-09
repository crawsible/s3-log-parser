package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func main() {
	client := s3.New(&aws.Config{
		Region: aws.String("us-west-2"),
	})
	downloader := s3manager.NewDownloader(&s3manager.DownloadOptions{
		S3: client,
	})

	params := &s3.ListObjectsInput{
		Bucket:  aws.String("lattice-logs"),
		MaxKeys: aws.Int64(1<<31 - 1),
		Prefix:  aws.String("logs/"),
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

	latestLogFiles := logFiles[len(logFiles)-21 : len(logFiles)-1]

	buffer := &aws.WriteAtBuffer{}
	for _, f := range latestLogFiles {
		_, err := downloader.Download(buffer, &s3.GetObjectInput{
			Bucket: aws.String("lattice-logs"),
			Key:    aws.String(f),
		})
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}

	logs := strings.Split(string(buffer.Bytes()), "\n")
	ips := []string{}
	reConcourse := regexp.MustCompile("^.*user/lattice-concourse.*$")
	reIP := regexp.MustCompile(`^.* ((\d+\.){3}\d+) .*$`)
	for _, l := range logs {
		if reConcourse.MatchString(l) {
			continue
		}

		matches := reIP.FindStringSubmatch(l)
		if len(matches) < 2 {
			continue
		}
		ips = append(ips, matches[1])
	}

	fmt.Println(ips)
}
