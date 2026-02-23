package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path"
	"regexp"
	"strings"
	"time"

	"docapp/core/internal/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var invalidPathChars = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

type DocumentStorage interface {
	EnsureBucket(ctx context.Context) error
	PutObject(ctx context.Context, objectKey, contentType string, content []byte) error
	GetObject(ctx context.Context, objectKey string) ([]byte, error)
	PresignGetObject(ctx context.Context, objectKey string) (string, error)
	BuildDocumentKey(tipo, competencia, cnpj, filename string) string
}

type MinioStorage struct {
	client        *minio.Client
	bucket        string
	presignExpiry time.Duration
}

func NewMinioStorage(cfg *config.Config) (*MinioStorage, error) {
	client, err := minio.New(cfg.StorageEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.StorageAccessKey, cfg.StorageSecretKey, ""),
		Secure: cfg.StorageUseSSL,
		Region: cfg.StorageRegion,
	})
	if err != nil {
		return nil, fmt.Errorf("creating minio client: %w", err)
	}

	expiry := time.Duration(cfg.StoragePresignMinutes) * time.Minute
	if expiry <= 0 {
		expiry = 15 * time.Minute
	}

	return &MinioStorage{
		client:        client,
		bucket:        cfg.StorageBucket,
		presignExpiry: expiry,
	}, nil
}

func (s *MinioStorage) EnsureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("checking bucket existence: %w", err)
	}

	if exists {
		return nil
	}

	if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("creating bucket: %w", err)
	}

	return nil
}

func (s *MinioStorage) PutObject(ctx context.Context, objectKey, contentType string, content []byte) error {
	if _, err := s.client.PutObject(ctx, s.bucket, objectKey, bytes.NewReader(content), int64(len(content)), minio.PutObjectOptions{
		ContentType: contentType,
	}); err != nil {
		return fmt.Errorf("put object %q: %w", objectKey, err)
	}

	return nil
}

func (s *MinioStorage) GetObject(ctx context.Context, objectKey string) ([]byte, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get object %q: %w", objectKey, err)
	}
	defer obj.Close()

	content, err := io.ReadAll(obj)
	if err != nil {
		return nil, fmt.Errorf("read object %q: %w", objectKey, err)
	}

	return content, nil
}

func (s *MinioStorage) PresignGetObject(ctx context.Context, objectKey string) (string, error) {
	url, err := s.client.PresignedGetObject(ctx, s.bucket, objectKey, s.presignExpiry, nil)
	if err != nil {
		return "", fmt.Errorf("presign object %q: %w", objectKey, err)
	}

	return url.String(), nil
}

func (s *MinioStorage) BuildDocumentKey(tipo, competencia, cnpj, filename string) string {
	safeTipo := sanitizePathPart(tipo)
	if safeTipo == "" {
		safeTipo = "desconhecido"
	}

	safeCompetencia := sanitizePathPart(competencia)
	if safeCompetencia == "" {
		safeCompetencia = "sem_competencia"
	}

	safeCNPJ := sanitizePathPart(cnpj)
	if safeCNPJ == "" {
		safeCNPJ = "sem_cnpj"
	}

	safeFilename := sanitizeFileName(filename)

	return path.Join(safeTipo, safeCompetencia, safeCNPJ, safeFilename)
}

func sanitizePathPart(value string) string {
	cleaned := strings.TrimSpace(strings.ToLower(value))
	cleaned = invalidPathChars.ReplaceAllString(cleaned, "_")
	cleaned = strings.Trim(cleaned, "._-")
	return cleaned
}

func sanitizeFileName(value string) string {
	cleaned := strings.TrimSpace(value)
	cleaned = invalidPathChars.ReplaceAllString(cleaned, "_")
	cleaned = strings.Trim(cleaned, "._-")
	if cleaned == "" {
		return "arquivo"
	}
	return cleaned
}
