package scommerce

import (
	"bytes"
	"context"
	"io"
	"os"
)

var _ FileIO = &OSFileIO{}
var _ FileIO = &BytesFileIO{}

type FileIdentifier interface {
	GetToken(ctx context.Context) (string, error)
}

type FileReader interface {
	io.Reader
	FileIdentifier
}

type FileWriter interface {
	io.Writer
	FileIdentifier
}

type FileCloser interface {
	io.Closer
	FileIdentifier
}

type FileReadCloser interface {
	FileReader
	FileCloser
}

type FileWriteCloser interface {
	FileWriter
	FileCloser
}

type FileIO interface {
	FileReader
	FileWriter
	FileCloser
	io.Seeker
}

type FileStorage interface {
	GeneralClosable
	Connect(ctx context.Context) error
	Open(ctx context.Context, token string) (FileIO, error)
	Create(ctx context.Context, token string) (FileIO, error)
	Exists(ctx context.Context, token string) (bool, error)
	Delete(ctx context.Context, token string) error
	DeleteAll(ctx context.Context, path string) error
}

type OSFileIO struct {
	File  *os.File
	Token string
}

type BytesFileIO struct {
	File  *bytes.Reader
	Token string
}

func (file *OSFileIO) Close() error {
	return file.File.Close()
}

func (file *OSFileIO) GetToken(ctx context.Context) (string, error) {
	return file.Token, nil
}

func (file *OSFileIO) Read(p []byte) (n int, err error) {
	n, err = file.File.Read(p)
	return
}

func (file *OSFileIO) Write(p []byte) (n int, err error) {
	return file.File.Write(p)
}

func (file *OSFileIO) Seek(offset int64, whence int) (int64, error) {
	return file.File.Seek(offset, whence)
}

func (file *BytesFileIO) Close() error {
	return nil
}

func (file *BytesFileIO) GetToken(ctx context.Context) (string, error) {
	return file.Token, nil
}

func (file *BytesFileIO) Read(p []byte) (n int, err error) {
	return file.File.Read(p)
}

func (file *BytesFileIO) Write(p []byte) (n int, err error) {
	return 0, nil
}

func (file *BytesFileIO) Seek(offset int64, whence int) (int64, error) {
	return file.File.Seek(offset, whence)
}
