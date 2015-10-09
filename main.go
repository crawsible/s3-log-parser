package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func LatestLogFilesList(client *s3.S3, downloader *s3manager.Downloader, numLogs int) ([]string, error) {
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
		return []string{}, err
	}

	return logFiles[len(logFiles)-numLogs-1 : len(logFiles)-1], nil
}

func GetLogLines(client *s3.S3, downloader *s3manager.Downloader, logFiles []string) ([]string, error) {
	buffer := &aws.WriteAtBuffer{}
	for i, f := range logFiles {
		if i > 0 {
			fmt.Printf("\033[1A")
		}
		fmt.Printf("Downloading log %d of %d...\n", i+1, len(logFiles))
		_, err := downloader.Download(buffer, &s3.GetObjectInput{
			Bucket: aws.String("lattice-logs"),
			Key:    aws.String(f),
		})
		if err != nil {
			return []string{}, err
		}
	}

	return strings.Split(string(buffer.Bytes()), "\n"), nil
}

func main() {
	client := s3.New(&aws.Config{
		Region: aws.String("us-west-2"),
	})
	downloader := s3manager.NewDownloader(&s3manager.DownloadOptions{
		S3: client,
	})

	logFiles, err := LatestLogFilesList(client, downloader, 10)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	var logs []string
	logs, err = GetLogLines(client, downloader, logFiles)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	IPs := map[string]int{}
	//reConcourse := regexp.MustCompile("^.*user/lattice-concourse.*$")
	reIP := regexp.MustCompile(`^.* ((\d+\.){3}\d+) .*$`)
	TotalDownloads := 0
	for _, l := range logs {
		//if reConcourse.MatchString(l) {
		//continue
		//}

		matches := reIP.FindStringSubmatch(l)
		if len(matches) < 2 {
			continue
		}

		TotalDownloads++
		IPs[matches[1]]++
	}

	fmt.Println(IPs)

	AWSIPRanges, err := GetAWSPublicIPRanges()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	AWSDownloads := 0
	for IP := range IPs {
		IPValue, err := getIPValue(IP)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		for _, IPRange := range AWSIPRanges {
			if IPValue >= IPRange[0] && IPValue < IPRange[1] {
				AWSDownloads += IPs[IP]
				break
			}
		}
	}

	fmt.Println("Total Downloads: ", TotalDownloads)
	fmt.Println("AWS Downloads:   ", AWSDownloads)
}
