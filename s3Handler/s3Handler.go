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

// InitS3 Creates a fresh S3 client, required for the running of the other S3 functions
func InitS3() {
	// Load AWS configuration.
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(S3_REGION))
	if err != nil {
		log.Fatalf("unable to load AWS SDK config, %v", err)
	}

	// Create an S3 client
	s3Client = s3.NewFromConfig(cfg)
}

// uploadToS3 uploads the file to S3 and returns the file URL
func UploadToS3(file multipart.File, handler *multipart.FileHeader) (string, error) {
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
	_, err := s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
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

// DeleteFromS3 deletes a file from S3 given its path and returns an error if any occurs
func DeleteFromS3(filePath string) error {
	// Delete the file from S3
	_, err := s3Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(S3_BUCKET),
		Key:    aws.String(filePath),
	})

	if err != nil {
		return fmt.Errorf("failed to delete object %s from S3: %v", filePath, err)
	}

	log.Printf("Successfully deleted %s from S3", filePath)
	return nil
}

// DeleteAllFromS3 deletes ALL files from S3 given the folder and returns an error if any occurs
func DeleteAllFromS3(folderPath string) error {
	fileContents, err := ListAllInS3(folderPath)
	if err != nil {
		return fmt.Errorf("failed to list objects in folder: %v", err)
	}

	if len(fileContents) == 0 {
		log.Println("No files found in the specified folder.")
		return nil
	}

	// Step 2: Delete each object
	for _, key := range fileContents {
		deleteInput := &s3.DeleteObjectInput{
			Bucket: aws.String(S3_BUCKET),
			Key:    aws.String(key),
		}

		_, err = s3Client.DeleteObject(context.TODO(), deleteInput)
		if err != nil {
			return fmt.Errorf("failed to delete object %s: %v", key, err)
		}

		log.Printf("Successfully deleted: %s\n", key)
	}

	log.Println("All files in the folder were deleted.")
	return nil
}

// ListAllInS3 returns a list of all the items inside the given folder
func ListAllInS3(folderPath string) ([]string, error) {
	var fileContents []string
	// List all the objects with the given prefix (folderPath)
	listInput := &s3.ListObjectsV2Input{
		Bucket: aws.String(S3_BUCKET),
		Prefix: aws.String(folderPath),
	}

	listOutput, err := s3Client.ListObjectsV2(context.TODO(), listInput)
	if err != nil {
		return fileContents, fmt.Errorf("failed to list objects in folder: %v", err)
	}

	for _, item := range listOutput.Contents {
		fileContents = append(fileContents, *item.Key)
	}
	return fileContents, nil
}
