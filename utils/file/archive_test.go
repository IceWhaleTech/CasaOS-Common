package file

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/golang/snappy"
	"github.com/pierrec/lz4/v4"
	"github.com/ulikunitz/xz"
)

func TestArchiveDownloadFormats(t *testing.T) {
	root := filepath.Join(t.TempDir(), "selected")
	if err := os.MkdirAll(filepath.Join(root, "empty-dir"), 0755); err != nil {
		t.Fatal(err)
	}
	for name, data := range map[string]string{"内容.txt": "download bytes", "empty.txt": ""} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(data), 0640); err != nil {
			t.Fatal(err)
		}
	}
	// Symlinks, including a link back to the parent, must not be followed.
	if err := os.Symlink(filepath.Dir(root), filepath.Join(root, "loop")); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"selected/内容.txt": "download bytes", "selected/empty.txt": "", "selected/empty-dir/": ""}
	for _, format := range []string{"zip", "tar", "targz", "tarbz2", "tarxz", "tarlz4", "tarsz"} {
		t.Run(format, func(t *testing.T) {
			_, a, err := GetCompressionAlgorithm(format)
			if err != nil {
				t.Fatal(err)
			}
			var out bytes.Buffer
			if err := a.Create(&out); err != nil {
				t.Fatal(err)
			}
			if err := AddFile(a, root, root); err != nil {
				t.Fatal(err)
			}
			if err := a.Close(); err != nil {
				t.Fatal(err)
			}
			got := map[string]string{}
			if format == "zip" {
				z, err := zip.NewReader(bytes.NewReader(out.Bytes()), int64(out.Len()))
				if err != nil {
					t.Fatal(err)
				}
				for _, f := range z.File {
					r, err := f.Open()
					if err != nil {
						t.Fatal(err)
					}
					data, err := io.ReadAll(r)
					r.Close()
					if err != nil {
						t.Fatal(err)
					}
					got[f.Name] = string(data)
				}
			} else {
				var r io.Reader = bytes.NewReader(out.Bytes())
				switch format {
				case "targz":
					z, err := gzip.NewReader(r)
					if err != nil {
						t.Fatal(err)
					}
					defer z.Close()
					r = z
				case "tarbz2":
					r = bzip2.NewReader(r)
				case "tarxz":
					z, err := xz.NewReader(r)
					if err != nil {
						t.Fatal(err)
					}
					r = z
				case "tarlz4":
					r = lz4.NewReader(r)
				case "tarsz":
					r = snappy.NewReader(r)
				}
				tr := tar.NewReader(r)
				for {
					h, err := tr.Next()
					if err == io.EOF {
						break
					}
					if err != nil {
						t.Fatal(err)
					}
					data, err := io.ReadAll(tr)
					if err != nil {
						t.Fatal(err)
					}
					name := h.Name
					if h.Typeflag == tar.TypeDir {
						name += "/"
					}
					got[name] = string(data)
				}
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("archive members differ: %#v", got)
			}
		})
	}
}

func TestArchiveRejectsUnsafeNamesAndPropagatesClose(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "source")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		t.Fatal(err)
	}
	for _, format := range []string{"zip", "tar"} {
		_, a, _ := GetCompressionAlgorithm(format)
		if err := a.Create(io.Discard); err != nil {
			t.Fatal(err)
		}
		for _, name := range []string{"../escape", "/absolute", "a/../../escape", "C:/escape", `a\escape`, ".", "", "a\x00b"} {
			if err := a.Write(ArchiveFile{FileInfo: info, Name: name, Reader: f}); err == nil {
				t.Fatalf("unsafe name accepted: %q", name)
			}
		}
		if err := AddFile(a, f.Name(), filepath.Join(filepath.Dir(f.Name()), "unrelated")); err == nil {
			t.Fatal("source outside common directory accepted")
		}
		a.Close()
	}
	for _, format := range []string{"zip", "tar", "targz", "tarbz2", "tarxz", "tarlz4", "tarsz"} {
		_, a, _ := GetCompressionAlgorithm(format)
		if err := a.Create(archiveFailureWriter{}); err != nil {
			continue // A codec may write its header immediately.
		}
		if err := a.Close(); !errors.Is(err, errArchiveWrite) {
			t.Fatalf("%s close failure hidden: %v", format, err)
		}
	}
	if _, _, err := GetCompressionAlgorithm("unknown"); err == nil {
		t.Fatal("unknown format accepted")
	}
}

var errArchiveWrite = errors.New("archive destination failed")

type archiveFailureWriter struct{}

func (archiveFailureWriter) Write([]byte) (int, error) { return 0, errArchiveWrite }
