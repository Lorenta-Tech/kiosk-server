package s3

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Lorenta-Tech/kiosk-server/internal/env"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	stagingPrefix = "uploads/staging"
	finalPrefix   = "uploads/final"
	presignExpiry = 15 * time.Minute
)

type Client struct {
	s3      *s3.Client
	presign *s3.PresignClient
	bucket  string
}

func Connect() (*Client, error) {
	var (
		region    = env.GetString("REGION", "ap-south-1")
		accessKey = env.GetString("ACCESS_KEY", "")
		secretKey = env.GetString("SECRETE_KEY", "")
		bucket    = env.GetString("BUCKET", "")
	)

	ctx,cancel := context.WithTimeout(context.Background(),time.Second*30)
	defer cancel()

	cfg, err := awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load s3 config: %w", err)
	}

	svc := s3.NewFromConfig(cfg)

	return &Client{
		s3:      svc,
		presign: s3.NewPresignClient(svc),
		bucket:  bucket,
	}, nil
}

//uploads/staging/{userID}/{sessionID}/{fileID}.pdf
func StagingKey(userID, sessionID, fileID string) string {
	return fmt.Sprintf("%s/%s/%s/%s.pdf", stagingPrefix, userID, sessionID, fileID)
}

//uploads/final/{userID}/{sessionID}/{fileID}.pdf
func FinalKey(stagingkey string) string {
	return strings.Replace(stagingkey, stagingPrefix+"/", finalPrefix+"/", 1)
}

//generates staging presigned url
func (c *Client) PresignPut(ctx context.Context, stagingKey string) (string, error) {
	req, err := c.presign.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(stagingKey),
		ContentType: aws.String("application/pdf"),
	}, s3.WithPresignExpires(presignExpiry))
	if err != nil {
		return "", fmt.Errorf("failed to presign put for key %s: %w", stagingKey, err)
	}

	return req.URL, nil
}

//checks the fileExists used during confirm step
func (c *Client) FileExists(ctx context.Context, key string) (bool, error) {
	_, err := c.s3.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if strings.Contains(err.Error(), "NoSuchKey") ||
			strings.Contains(err.Error(), "NotFound") ||
			strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, fmt.Errorf("failed to head object %s: %w", key, err)
	}

	return true, nil
}

/*
PromoteFile copies a file from staging to final, then deletes from staging.
This is "commit" step that runs on payment webhook verified.
returns the final key on successfull payments
*/
func (c *Client) PromoteFile(ctx context.Context, stagingKey string) (string, error) {
	finalKey := FinalKey(stagingKey)
	copySource := fmt.Sprintf("%s/%s", c.bucket, stagingKey)

	_, err := c.s3.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(c.bucket),
		CopySource: aws.String(copySource),
		Key:        aws.String(finalKey),
	})
	if err != nil {
		return "", fmt.Errorf("failed to copy %s → %s: %w", stagingKey, finalKey, err)
	}

	_, err = c.s3.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(stagingKey),
	})
	if err != nil {
		//file is alredy promoted but fails to delete , No worries Lifecylcle rule will applied here staging files gets cleans within 24h.
		return finalKey, fmt.Errorf("promoted but failed to delete staging %s: %w", stagingKey, err)
	}

	return finalKey, nil
}
//extra function just used when need to mannual cleanup.
func (c *Client) DeleteStagingFile(ctx context.Context, stagingKey string) error {
	_, err := c.s3.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(stagingKey),
	})
	if err != nil {
		return fmt.Errorf("failed to delete staging file %s: %w", stagingKey, err)
	}

	return nil
}

