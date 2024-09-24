package s3Handler

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var s3Client *s3.Client

const (
	S3_REGION = "eu-west-2"
	S3_BUCKET = "my-programming-agent-img-store"
)

// uploadToS3 uploads the file to S3 and returns the file URL
func UploadToS3(file multipart.File, handler *multipart.FileHeader) (string, error) {

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(S3_REGION))
	if err != nil {
		log.Fatalf("unable to load AWS SDK config, %v", err)
	}

	// Create an S3 client
	s3Client = s3.NewFromConfig(cfg)

	// Read file content
	size := handler.Size
	buffer := make([]byte, size)
	file.Read(buffer)

	// Prepare the file for upload
	fileBytes := bytes.NewReader(buffer)
	fileType := http.DetectContentType(buffer)
	fileName := filepath.Base(handler.Filename)
	s3Path := "uploads/" + fileName

	// Upload the file to S3
	_, err = s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:        aws.String(S3_BUCKET),
		Key:           aws.String(s3Path),
		Body:          fileBytes,
		ContentLength: &size,
		ContentType:   aws.String(fileType),
	})

	if err != nil {
		return "", err
	}

	// Construct the file URL
	fileURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", S3_BUCKET, S3_REGION, s3Path)
	return fileURL, nil
}
