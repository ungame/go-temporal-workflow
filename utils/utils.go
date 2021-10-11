package utils

import (
	"context"
	"io"
	"log"
	"path/filepath"
	"runtime"
)

func GetEnvFilePath() string {
	_, f, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(f), "../.env")
}

type CloserContext interface {
	Close(ctx context.Context) error
}

func HandleCloseContext(ctx context.Context, closer CloserContext) {
	if closer != nil {
		err := closer.Close(ctx)
		if err != nil {
			log.Printf("error on close %T: %s\n", closer, err)
		}
	}
}

func HandleClose(closer io.Closer) {
	if closer != nil {
		err := closer.Close()
		if err != nil {
			log.Printf("error on close %T: %s\n", closer, err)
		}
	}
}