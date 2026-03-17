package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

// StorageClient defines the interface for file storage operations
type StorageClient interface {
	UploadFile(ctx context.Context, file io.Reader, filename string, contentType string) (*UploadFileResult, error)
	DeleteFile(ctx context.Context, key string) error
}

// S3Config contains configuration for S3 client
type S3Config struct {
	Region          string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
	Endpoint        string // Optional: for local development with MinIO
}

// S3Client wraps AWS S3 operations
type S3Client struct {
	client *s3.Client
	bucket string
}

// NewS3Client creates a new S3 client
func NewS3Client(cfg S3Config) (*S3Client, error) {
	var awsConfig aws.Config
	var err error

	// Configure credentials
	credProvider := credentials.NewStaticCredentialsProvider(
		cfg.AccessKeyID,
		cfg.SecretAccessKey,
		"",
	)

	// Load AWS config
	if cfg.Endpoint != "" {
		// Use custom endpoint (e.g., MinIO for local dev)
		customResolver := aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:               cfg.Endpoint,
					HostnameImmutable: true,
					SigningRegion:     cfg.Region,
				}, nil
			},
		)

		awsConfig, err = config.LoadDefaultConfig(context.Background(),
			config.WithRegion(cfg.Region),
			config.WithCredentialsProvider(credProvider),
			config.WithEndpointResolverWithOptions(customResolver),
		)
	} else {
		// Use standard AWS S3
		awsConfig, err = config.LoadDefaultConfig(context.Background(),
			config.WithRegion(cfg.Region),
			config.WithCredentialsProvider(credProvider),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		o.UsePathStyle = cfg.Endpoint != "" // Use path-style for MinIO
	})

	return &S3Client{
		client: client,
		bucket: cfg.Bucket,
	}, nil
}

// UploadFileResult contains information about uploaded file
type UploadFileResult struct {
	Key      string
	URL      string
	FileSize int64
}

// UploadFile uploads a file to S3 and returns the key and URL
func (c *S3Client) UploadFile(ctx context.Context, file io.Reader, filename string, contentType string) (*UploadFileResult, error) {
	// Generate unique key with timestamp and UUID
	timestamp := time.Now().Format("2006/01/02")
	uniqueID := uuid.New().String()
	ext := filepath.Ext(filename)
	key := fmt.Sprintf("videos/%s/%s%s", timestamp, uniqueID, ext)

	// Read file content into buffer to get size
	buffer := new(bytes.Buffer)
	size, err := io.Copy(buffer, file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Upload to S3
	_, err = c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(buffer.Bytes()),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload to S3: %w", err)
	}

	// Generate public URL (assumes bucket is public or has proper access policies)
	url := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", c.bucket, key)

	return &UploadFileResult{
		Key:      key,
		URL:      url,
		FileSize: size,
	}, nil
}

// DeleteFile deletes a file from S3
func (c *S3Client) DeleteFile(ctx context.Context, key string) error {
	_, err := c.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}
	return nil
}

// GeneratePresignedURL generates a presigned URL for temporary access
func (c *S3Client) GeneratePresignedURL(ctx context.Context, key string, expiration time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(c.client)

	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiration
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return request.URL, nil
}
