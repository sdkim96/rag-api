package blobstore

import (
	"context"
	"fmt"
	"io"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
)

type AzureBlob struct {
	uri   string
	props *blob.GetPropertiesResponse
}

func (b *AzureBlob) URI() string {
	return b.uri
}

func (b *AzureBlob) MimeType() string {
	if b.props.ContentType == nil {
		return ""
	}
	return *b.props.ContentType
}

func (b *AzureBlob) Size() int64 {
	if b.props.ContentLength == nil {
		return 0
	}
	return *b.props.ContentLength
}

type AzureBlobStore struct {
	c             *azblob.Client
	accountName   string
	containerName string
}

func NewAzureBlobStore(accountName, connStr, containerName string) (*AzureBlobStore, error) {
	c, err := azblob.NewClientFromConnectionString(connStr, nil)
	if err != nil {
		return nil, err
	}
	return &AzureBlobStore{
		c:             c,
		accountName:   accountName,
		containerName: containerName,
	}, nil
}

func (a *AzureBlobStore) uri(key string) string {
	return fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s", a.accountName, a.containerName, key)
}

func (a *AzureBlobStore) Get(ctx context.Context, key string) (Blob, error) {
	props, err := a.c.ServiceClient().
		NewContainerClient(a.containerName).
		NewBlobClient(key).
		GetProperties(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &AzureBlob{
		uri:   a.uri(key),
		props: &props,
	}, nil
}

func (a *AzureBlobStore) Upload(ctx context.Context, key, mimeType string, r io.Reader) error {
	_, err := a.c.UploadStream(ctx, a.containerName, key, r, &azblob.UploadStreamOptions{
		HTTPHeaders: &blob.HTTPHeaders{
			BlobContentType: &mimeType,
		},
	})
	return err
}

func (a *AzureBlobStore) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	resp, err := a.c.DownloadStream(ctx, a.containerName, key, nil)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (a *AzureBlobStore) Delete(ctx context.Context, key string) error {
	_, err := a.c.DeleteBlob(ctx, a.containerName, key, nil)
	return err
}

var _ BlobStore = (*AzureBlobStore)(nil)
