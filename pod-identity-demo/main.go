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
	"github.com/tidwall/gjson"
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
	awsContainerCredentialsUri := os.Getenv("AWS_CONTAINER_CREDENTIALS_FULL_URI")
	awsTokenPath := os.Getenv("AWS_CONTAINER_AUTHORIZATION_TOKEN_FILE")
	bucket := os.Getenv("S3_BUCKET_NAME")

	if awsRegion == "" {
		log.Fatalf("AWS_REGION environment variable not set")
	}

	if bucket == "" {
		log.Fatalf("S3_BUCKET_NAME environment variable not set")
	}

	awsToken, err := os.ReadFile(awsTokenPath)
	if err != nil {
		log.Printf("Unable to read aws token in %s", awsTokenPath)
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", awsContainerCredentialsUri, nil)

	if err != nil {
		log.Fatalf("Unable to build http request")
	}
	req.Header.Set("Authorization", string(awsToken))
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed getting the temporary credentials from %s", awsContainerCredentialsUri)
	}
	defer res.Body.Close()

	var bodyString string

	if res.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			log.Fatal(err)
		}
		bodyString = string(bodyBytes)
	} else {
		log.Fatalf("Invalid response code: %d for %s", res.StatusCode, awsContainerCredentialsUri)
	}

	awsAccessKey := gjson.Get(bodyString, "AccessKeyId")
	awsSecretKey := gjson.Get(bodyString, "SecretAccessKey")
	awsSessionToken := gjson.Get(bodyString, "Token")

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(awsRegion),
		Credentials: credentials.NewStaticCredentials(
			awsAccessKey.String(),
			awsSecretKey.String(),
			awsSessionToken.String(),
		),
	})

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

		filename := fmt.Sprintf("%s-%s.jpg", os.Getenv("HOSTNAME"), name)
		_, err = uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(filename),
			Body:   imgBody,
		})
		if err != nil {
			log.Printf("Failed to upload file, %v", err)
		} else {
			log.Printf("Successfully uploaded %q to S3 bucket %q\n", filename, bucket)
		}

		imgBody.Close()
		time.Sleep(30 * time.Second)
	}
}
