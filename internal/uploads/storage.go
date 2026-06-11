package uploads

import (
	"context"
	"io"
)

type Storage interface {
	Save(ctx context.Context, filename string, content io.Reader) (UploadResponse, error)
}
