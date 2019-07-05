package s3

import (
	"context"
	"io"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// ObjectStorage is a S3-compatible storage interface.
type ObjectStorage interface {
	Download(ctx context.Context, w io.WriterAt, URI string) (int64, error)
}

// ObjectStorageImpl is our implementatino of the ObjectStorage interface.
type ObjectStorageImpl struct {
	client     s3iface.S3API
	downloader *s3manager.Downloader
}

// New returns a pointer to a new ObjectStorageImpl.
func New(sess *session.Session) *ObjectStorageImpl {
	client := s3.New(sess)
	return &ObjectStorageImpl{
		client:     client,
		downloader: s3manager.NewDownloaderWithClient(client),
	}
}

// Download writes the contents of a remote file into the given writer.
func (s *ObjectStorageImpl) Download(ctx context.Context, w io.WriterAt, URI string) (n int64, err error) {
	bucket, key, err := getBucketAndKey(URI)
	if err != nil {
		return -1, err
	}
	req := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	return s.downloader.DownloadWithContext(ctx, w, req)
}

func getBucketAndKey(URI string) (bucket string, key string, err error) {
	u, err := url.Parse(URI)
	if err != nil {
		return "", "", err
	}
	return u.Hostname(), strings.TrimPrefix(u.Path, "/"), nil
}
