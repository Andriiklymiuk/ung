package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Config holds S3 storage configuration
type S3Config struct {
	Bucket          string
	Region          string
	Endpoint        string // Optional: for S3-compatible services
	AccessKeyID     string
	SecretAccessKey string
	UsePathStyle    bool // For S3-compatible services like MinIO
}

// S3Storage handles S3 operations for tenant databases
type S3Storage struct {
	client *s3.Client
	bucket string
	config *S3Config
}

// NewS3Storage creates a new S3 storage instance
func NewS3Storage(cfg *S3Config) (*S3Storage, error) {
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("S3 bucket name is required")
	}

	// Load AWS config
	ctx := context.Background()
	var awsCfg aws.Config
	var err error

	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		// Use provided credentials
		awsCfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(cfg.Region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				cfg.AccessKeyID,
				cfg.SecretAccessKey,
				"",
			)),
		)
	} else {
		// Use default credential chain (IAM role, env vars, etc.)
		awsCfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(cfg.Region),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client
	clientOptions := []func(*s3.Options){}

	if cfg.Endpoint != "" {
		clientOptions = append(clientOptions, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.UsePathStyle = cfg.UsePathStyle
		})
	}

	client := s3.NewFromConfig(awsCfg, clientOptions...)

	return &S3Storage{
		client: client,
		bucket: cfg.Bucket,
		config: cfg,
	}, nil
}

// UploadTenantDB uploads an encrypted tenant database to S3
func (s *S3Storage) UploadTenantDB(ctx context.Context, tenantID string, dbPath string) error {
	// Open the database file
	file, err := os.Open(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database file: %w", err)
	}
	defer file.Close()

	// Generate S3 key (path in bucket)
	key := s.getTenantDBKey(tenantID)

	// Upload to S3
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   file,
		Metadata: map[string]string{
			"tenant-id": tenantID,
			"encrypted": "true",
		},
	})

	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	return nil
}

// DownloadTenantDB downloads a tenant database from S3
func (s *S3Storage) DownloadTenantDB(ctx context.Context, tenantID string, destPath string) error {
	// Generate S3 key
	key := s.getTenantDBKey(tenantID)

	// Get object from S3
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return fmt.Errorf("failed to download from S3: %w", err)
	}
	defer result.Body.Close()

	// Create destination directory if needed
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create destination file
	file, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy data
	if _, err := io.Copy(file, result.Body); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// DeleteTenantDB deletes a tenant database from S3
func (s *S3Storage) DeleteTenantDB(ctx context.Context, tenantID string) error {
	key := s.getTenantDBKey(tenantID)

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}

	return nil
}

// TenantDBExists checks if a tenant database exists in S3
func (s *S3Storage) TenantDBExists(ctx context.Context, tenantID string) (bool, error) {
	key := s.getTenantDBKey(tenantID)

	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		// Check if it's a "not found" error
		return false, nil
	}

	return true, nil
}

// getTenantDBKey generates the S3 key for a tenant database
func (s *S3Storage) getTenantDBKey(tenantID string) string {
	return fmt.Sprintf("tenants/%s/ung.db.encrypted", tenantID)
}

// ListTenantBackups lists all available backups for a tenant
func (s *S3Storage) ListTenantBackups(ctx context.Context, tenantID string) ([]string, error) {
	prefix := fmt.Sprintf("tenants/%s/backups/", tenantID)

	result, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list backups: %w", err)
	}

	backups := make([]string, 0, len(result.Contents))
	for _, obj := range result.Contents {
		backups = append(backups, *obj.Key)
	}

	return backups, nil
}

// CreateBackup creates a timestamped backup of the tenant database
func (s *S3Storage) CreateBackup(ctx context.Context, tenantID string) error {
	// Get current database
	sourceKey := s.getTenantDBKey(tenantID)
	backupKey := fmt.Sprintf("tenants/%s/backups/ung_%d.db.encrypted",
		tenantID,
		time.Now().Unix(),
	)

	// Copy object to backup location
	_, err := s.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(s.bucket),
		CopySource: aws.String(fmt.Sprintf("%s/%s", s.bucket, sourceKey)),
		Key:        aws.String(backupKey),
	})

	if err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	return nil
}
