package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const unsplashURL = "https://source.unsplash.com/random"

func downloadRandomImage() (io.ReadCloser, error) {
	resp, err := http.Get(unsplashURL)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func main() {
	awsRegion := os.Getenv("AWS_REGION")
	awsAccessKey := os.Getenv("AWS_ACCESS_KEY")
	awsSecretKey := os.Getenv("AWS_SECRET_KEY")
	bucket := os.Getenv("S3_BUCKET_NAME")

	// if awsRegion == "" {
	// 	log.Fatalf("AWS_REGION environment variable not set")
	// }

	// if bucket == "" {
	// 	log.Fatalf("S3_BUCKET_NAME environment variable not set")
	// }

	podName := os.Getenv("HOSTNAME")

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(awsRegion),
	})

	if awsAccessKey != "" && awsSecretKey != "" {
		sess, err = session.NewSession(&aws.Config{
			Region: aws.String(awsRegion),
			Credentials: credentials.NewStaticCredentials(
				awsAccessKey,
				awsSecretKey,
				"",
			),
		})
	}

	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}

	uploader := s3manager.NewUploader(sess)

	// Download a random image from unsplash every
	// 30 seconds & upload it to the S3 bucket
	for {
		imgBody, err := downloadRandomImage()
		if err != nil {
			log.Printf("Failed to download image: %v", err)
			continue
		}

		now := time.Now()
		name := now.Format("02 Jan 2006 15:04:05")

		filename := fmt.Sprintf("%s-%s.jpg", podName, name)
		_, err = uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(filename),
			Body:   imgBody,
		})
		if err != nil {
			log.Printf("Failed to upload file, %v", err)
		} else {
			fmt.Printf("Successfully uploaded %q to S3 bucket %q\n", filename, bucket)
		}

		imgBody.Close()
		time.Sleep(30 * time.Second)
	}
}
