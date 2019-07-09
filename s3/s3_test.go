package s3

import (
	"context"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/spf13/afero"
)

var fs = afero.Afero{Fs: afero.NewMemMapFs()}

func tempFile(t *testing.T) afero.File {
	file, err := fs.TempFile("", "")
	t.Logf("Created temporary file: %s", file.Name())
	if err != nil {
		t.Error(err)
	}
	return file
}

type mockS3Client struct {
	s3iface.S3API
	t *testing.T
	f afero.File
}

func (c *mockS3Client) GetObjectWithContext(ctx aws.Context, input *s3.GetObjectInput, opts ...request.Option) (*s3.GetObjectOutput, error) {
	return &s3.GetObjectOutput{
		Body:         c.f,
		ContentRange: aws.String("1"),
	}, nil
}

func TestObjectStorageImpl_Download(t *testing.T) {
	const want = "Hello world!"

	// Input file S3 mock is to read from
	fi := tempFile(t)
	defer fi.Close()
	fmt.Fprint(fi, want) // Write the contents
	fi.Seek(0, 0)

	// Output file we want to validate
	fo := tempFile(t)
	defer fo.Close()

	s3c := &mockS3Client{t: t, f: fi}
	s3d := s3manager.NewDownloaderWithClient(s3c)
	client := &ObjectStorageImpl{client: s3c, downloader: s3d}

	var err error

	_, err = client.Download(context.TODO(), fo, "[invalid-url]:12345")
	if err == nil {
		t.Error("Download() should have returned an error but didn't")
	}

	_, err = client.Download(context.TODO(), fo, "s3://foo/bar")
	if err != nil {
		t.Error(err)
	}

	fo.Seek(0, 0)
	data, err := ioutil.ReadAll(fo)
	if err != nil {
		t.Error(err)
	}
	have := string(data)
	if want != have {
		t.Errorf("want %s, got %s", want, have)
	}
}

func Test_getBucketAndKey(t *testing.T) {
	testCases := []struct {
		url     string
		bucket  string
		key     string
		wantErr bool
	}{
		{"s3://rdss-bucker-2344/filename.jpg", "rdss-bucker-2344", "filename.jpg", false},
		{"s3://a-different-bucket/wqefqwef/cert.pem", "a-different-bucket", "wqefqwef/cert.pem", false},
		{"[invalid-url]:12345", "", "", true},
	}
	for _, tc := range testCases {
		bucket, key, err := getBucketAndKey(tc.url)
		if tc.wantErr {
			if bucket != "" || key != "" || err == nil {
				t.Errorf("getBucketAndKey() was expected to fail but didn't")
			}
			return
		}
		if err != nil {
			t.Errorf("Unexpected error in getBucketAndKey: %s", err)
		}
		if bucket != tc.bucket {
			t.Errorf("Unexpected bucket - got: %s, want: %s", bucket, tc.bucket)
		}
		if key != tc.key {
			t.Errorf("Unexpected key - got: %s, want: %s", key, tc.key)
		}
	}
}
