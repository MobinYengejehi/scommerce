package fs

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/MobinYengejehi/scommerce/scommerce"
)

var _ scommerce.FileStorage = &LocalDiskFileStorage{}

type LocalDiskFileStorage struct {
	Directory string
}

func (fs *LocalDiskFileStorage) Close(ctx context.Context) error {
	return nil
}

func (fs *LocalDiskFileStorage) Connect(ctx context.Context) error {
	return nil
}

func NewLocalDiskFileStorage(directory string) *LocalDiskFileStorage {
	return &LocalDiskFileStorage{
		Directory: directory,
	}
}

func (fs *LocalDiskFileStorage) Create(ctx context.Context, token string) (scommerce.FileIO, error) {
	file, err := os.Create(filepath.Join(fs.Directory, token))
	if err != nil {
		return nil, err
	}
	return &scommerce.OSFileIO{
		File:  file,
		Token: token,
	}, nil
}

func (fs *LocalDiskFileStorage) Exists(ctx context.Context, token string) (bool, error) {
	file, err := os.Open(token)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	defer file.Close()
	return true, nil
}

func (fs *LocalDiskFileStorage) Open(ctx context.Context, token string) (scommerce.FileIO, error) {
	file, err := os.OpenFile(filepath.Join(fs.Directory, token), os.O_RDWR, 0o777)
	if err != nil {
		return nil, err
	}
	return &scommerce.OSFileIO{
		File:  file,
		Token: token,
	}, nil
}

func (fs *LocalDiskFileStorage) Delete(ctx context.Context, token string) error {
	return os.Remove(filepath.Join(fs.Directory, token))
}

func (fs *LocalDiskFileStorage) DeleteAll(ctx context.Context, path string) error {
	return os.RemoveAll(path)
}
