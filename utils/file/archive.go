package file

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"io"
	"io/fs"
	"path"
	"strings"

	"github.com/dsnet/compress/bzip2"
	"github.com/golang/snappy"
	"github.com/pierrec/lz4/v4"
	"github.com/ulikunitz/xz"
)

// ArchiveWriter only creates archives. Extraction is deliberately not part of
// this interface; callers must apply their own extraction and resource policy.
type ArchiveWriter interface {
	Create(io.Writer) error
	Write(ArchiveFile) error
	Close() error
}

type ArchiveFile struct {
	fs.FileInfo
	Name   string
	Reader io.Reader
}

type archiveWriter struct {
	format string
	zip    *zip.Writer
	tar    *tar.Writer
	codec  io.WriteCloser
	closed bool
}

// GetCompressionAlgorithm retains existing format names and on-wire formats
// without loading archiver/v3 or its unbounded legacy RAR decoder.
func GetCompressionAlgorithm(format string) (string, ArchiveWriter, error) {
	if format == "" {
		format = "zip"
	}
	extensions := map[string]string{"zip": ".zip", "tar": ".tar", "targz": ".tar.gz", "tarbz2": ".tar.bz2", "tarxz": ".tar.xz", "tarlz4": ".tar.lz4", "tarsz": ".tar.sz"}
	ext, ok := extensions[format]
	if !ok {
		return "", nil, errors.New("format not implemented")
	}
	return ext, &archiveWriter{format: format}, nil
}

func (a *archiveWriter) Create(w io.Writer) error {
	if a.closed || a.zip != nil || a.tar != nil || w == nil {
		return errors.New("archive writer already used or missing destination")
	}
	var err error
	switch a.format {
	case "zip":
		a.zip = zip.NewWriter(w)
		return nil
	case "targz":
		a.codec = gzip.NewWriter(w)
	case "tarbz2":
		a.codec, err = bzip2.NewWriter(w, &bzip2.WriterConfig{Level: bzip2.BestCompression})
	case "tarxz":
		a.codec, err = xz.NewWriter(w)
	case "tarlz4":
		a.codec = lz4.NewWriter(w)
	case "tarsz":
		a.codec = snappy.NewBufferedWriter(w)
	}
	if err != nil {
		return err
	}
	if a.codec != nil {
		w = a.codec
	}
	a.tar = tar.NewWriter(w)
	return nil
}

func (a *archiveWriter) Write(f ArchiveFile) error {
	if a.closed || (a.zip == nil && a.tar == nil) {
		return errors.New("archive writer is not open")
	}
	if f.FileInfo == nil || (!f.Mode().IsRegular() && !f.IsDir()) || f.Size() < 0 {
		return errors.New("unsupported archive entry")
	}
	if f.Name == "." || !fs.ValidPath(f.Name) || strings.ContainsAny(f.Name, "\\\x00:") || path.IsAbs(f.Name) {
		return errors.New("unsafe archive entry name")
	}
	var w io.Writer
	if a.zip != nil {
		h, err := zip.FileInfoHeader(f.FileInfo)
		if err != nil {
			return err
		}
		h.Name, h.Method = f.Name, zip.Deflate
		if f.IsDir() {
			h.Name += "/"
			h.Method = zip.Store
		}
		w, err = a.zip.CreateHeader(h)
		if err != nil {
			return err
		}
	} else {
		h, err := tar.FileInfoHeader(f.FileInfo, "")
		if err != nil {
			return err
		}
		h.Name = f.Name
		if err := a.tar.WriteHeader(h); err != nil {
			return err
		}
		w = a.tar
	}
	if f.IsDir() {
		return nil
	}
	if f.Reader == nil {
		return errors.New("missing archive entry reader")
	}
	_, err := io.CopyN(w, f.Reader, f.Size())
	return err
}

func (a *archiveWriter) Close() error {
	if a.closed {
		return nil
	}
	a.closed = true
	var err error
	if a.zip != nil {
		err = a.zip.Close()
	}
	if a.tar != nil {
		err = a.tar.Close()
	}
	if a.codec != nil {
		err = errors.Join(err, a.codec.Close())
	}
	return err
}
