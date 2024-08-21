package file

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/mholt/archiver/v3"
	"go.uber.org/zap"
)

// GetSize get the file size
func GetSize(f multipart.File) (int, error) {
	content, err := io.ReadAll(f)
	return len(content), err
}

// GetExt get the file ext
func GetExt(fileName string) string {
	return path.Ext(fileName)
}

// CheckNotExist check if the file exists
func CheckNotExist(src string) bool {
	_, err := os.Stat(src)

	return os.IsNotExist(err)
}

// CheckPermission check if the file has permission
func CheckPermission(src string) bool {
	_, err := os.Stat(src)

	return os.IsPermission(err)
}

// IsNotExistMkDir create a directory if it does not exist
func IsNotExistMkDir(src string) error {
	if notExist := CheckNotExist(src); notExist {
		if err := MkDir(src); err != nil {
			return err
		}
	}

	return nil
}

// MkDir create a directory
func MkDir(src string) error {
	err := os.MkdirAll(src, os.ModePerm)
	if err != nil {
		return err
	}
	return os.Chmod(src, 0o777)
}

// RMDir remove a directory
func RMDir(src string) error {
	err := os.RemoveAll(src)
	if err != nil {
		return err
	}
	os.Remove(src)
	return nil
}

// Open a file according to a specific mode
func Open(name string, flag int, perm os.FileMode) (*os.File, error) {
	f, err := os.OpenFile(name, flag, perm)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// MustOpen maximize trying to open the file
func MustOpen(fileName, filePath string) (*os.File, error) {
	//dir, err := os.Getwd()
	//if err != nil {
	//	return nil, fmt.Errorf("os.Getwd err: %v", err)
	//}

	src := filePath
	perm := CheckPermission(src)
	if perm {
		return nil, fmt.Errorf("file.CheckPermission Permission denied src: %s", src)
	}

	err := IsNotExistMkDir(src)
	if err != nil {
		return nil, fmt.Errorf("file.IsNotExistMkDir src: %s, err: %v", src, err)
	}

	f, err := Open(src+fileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, fmt.Errorf("fail to OpenFile :%v", err)
	}

	return f, nil
}

// 判断所给路径文件/文件夹是否存在
func Exists(path string) bool {
	_, err := os.Stat(path) // os.Stat获取文件信息
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

// 判断所给路径是否为文件夹
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// 判断所给路径是否为文件
func IsFile(path string) bool {
	return !IsDir(path)
}

func CreateFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return nil
}

func CreateFileAndWriteContent(path string, content string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0o666)
	if err != nil {
		return err
	}

	defer file.Close()
	write := bufio.NewWriter(file)

	if _, err := write.WriteString(content); err != nil {
		return err
	}

	write.Flush()
	return nil
}

// IsNotExistMkDir create a directory if it does not exist
func IsNotExistCreateFile(src string) error {
	if CheckNotExist(src) {
		if err := CreateFile(src); err != nil {
			return err
		}
	}

	return nil
}

func ReadFullFile(path string) []byte {
	file, err := os.Open(path)
	if err != nil {
		return []byte("")
	}
	defer file.Close()
	content, err := io.ReadAll(file)
	if err != nil {
		return []byte("")
	}
	return content
}

// File copies a single file from src to dst
func CopyFile(src, dst, style string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	lastPath := src[strings.LastIndex(src, "/")+1:]

	if !strings.HasSuffix(dst, "/") {
		dst += "/"
	}
	dst += lastPath
	if Exists(dst) {
		if style == "skip" {
			return nil
		}
		os.Remove(dst)
	}

	if srcfd, err = os.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}

/**
 * @description:
 * @param {*} src
 * @param {*} dst
 * @param {string} style
 * @return {*}
 * @method:
 * @router:
 */
func CopySingleFile(src, dst, style string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if Exists(dst) {
		if style == "skip" {
			return nil
		}
		os.Remove(dst)
	}

	if srcfd, err = os.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}

// Check for duplicate file names
func GetNoDuplicateFileName(fullPath string) string {
	dir, fileName := filepath.Split(fullPath)
	fileSuffix := path.Ext(fileName)
	filenameOnly := strings.TrimSuffix(fileName, fileSuffix)
	for i := 0; Exists(fullPath); i++ {
		fullPath = path.Join(dir, filenameOnly+"("+strconv.Itoa(i+1)+")"+fileSuffix)
	}
	return fullPath
}

// Dir copies a whole directory recursively
func CopyDir(src string, dst string, style string) error {
	var err error
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	if !srcinfo.IsDir() {
		if err = CopyFile(src, dst, style); err != nil {
			fmt.Println(err)
		}
		return nil
	}
	// dstPath := dst
	lastPath := src[strings.LastIndex(src, "/")+1:]
	dst += "/" + lastPath
	// for i := 0; Exists(dst); i++ {
	// 	dst = dstPath + "/" + lastPath + strconv.Itoa(i+1)
	// }
	if Exists(dst) {
		if style == "skip" {
			return nil
		}
		os.Remove(dst)
	}
	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	var fds []fs.DirEntry
	if fds, err = os.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := filepath.Join(src, fd.Name())
		dstfp := dst // filepath.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = CopyDir(srcfp, dstfp, style); err != nil {
				fmt.Println(err)
			}
		} else {
			if err = CopyFile(srcfp, dstfp, style); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

func WriteToPath(data []byte, path, name string) error {
	fullPath := path
	if strings.HasSuffix(path, "/") {
		fullPath += name
	} else {
		fullPath += "/" + name
	}
	return WriteToFullPath(data, fullPath, 0o666)
}

func WriteToFullPath(data []byte, fullPath string, perm fs.FileMode) error {
	if err := IsNotExistCreateFile(fullPath); err != nil {
		return err
	}

	file, err := os.OpenFile(fullPath,
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
		perm,
	)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(data)

	return err
}

// 最终拼接
func SpliceFiles(dir, path string, length int, startPoint int) error {
	fullPath := path

	if err := IsNotExistCreateFile(fullPath); err != nil {
		return err
	}

	file, _ := os.OpenFile(fullPath,
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
		0o666,
	)
	defer file.Close()
	bufferedWriter := bufio.NewWriter(file)
	for i := 0; i < length+startPoint; i++ {
		data, err := os.ReadFile(dir + "/" + strconv.Itoa(i+startPoint))
		if err != nil {
			return err
		}
		_, err = bufferedWriter.Write(data)
		if err != nil {
			return err
		}
	}

	bufferedWriter.Flush()

	return nil
}

func GetCompressionAlgorithm(t string) (string, archiver.Writer, error) {
	switch t {
	case "zip", "":
		return ".zip", archiver.NewZip(), nil
	case "tar":
		return ".tar", archiver.NewTar(), nil
	case "targz":
		return ".tar.gz", archiver.NewTarGz(), nil
	case "tarbz2":
		return ".tar.bz2", archiver.NewTarBz2(), nil
	case "tarxz":
		return ".tar.xz", archiver.NewTarXz(), nil
	case "tarlz4":
		return ".tar.lz4", archiver.NewTarLz4(), nil
	case "tarsz":
		return ".tar.sz", archiver.NewTarSz(), nil
	default:
		return "", nil, errors.New("format not implemented")
	}
}
func IsBrokenSymlink(path string) (bool, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return false, fmt.Errorf("error getting file info: %w", err)
	}

	// file is not a symlink
	if info.Mode()&os.ModeSymlink == 0 {
		return false, nil
	}

	target, err := os.Readlink(path)
	if err != nil {
		return false, fmt.Errorf("error reading symlink: %w", err)
	}

	_, err = os.Stat(target)
	if os.IsNotExist(err) {
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("error checking target: %w", err)
	}

	return false, nil
}

func AddFile(ar archiver.Writer, path, commonPath string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if !info.IsDir() && !info.Mode().IsRegular() {
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	if path != commonPath {
		//filename := info.Name()
		fpath := strings.Replace(path, commonPath, "", 1)
		fpath = filepath.Join(filepath.Base(commonPath), fpath)
		//filename := info.Name()
		err = ar.Write(archiver.File{
			FileInfo: archiver.FileInfo{
				FileInfo:   info,
				CustomName: fpath,
			},
			ReadCloser: file,
		})
		if err != nil {
			return err
		}
	}

	if info.IsDir() {
		names, err := file.Readdirnames(0)
		if err != nil {
			return err
		}

		for _, name := range names {
			filePath := filepath.Join(path, name)
			isBroken, err := IsBrokenSymlink(filePath)
			if err != nil {
				logger.Error("Failed to check symlink", zap.Any("name", filePath), zap.Error(err))
				continue
			}
			if isBroken {
				continue
			}

			err = AddFile(ar, filePath, commonPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func CommonPrefix(sep byte, paths ...string) string {
	// Handle special cases.
	switch len(paths) {
	case 0:
		return ""
	case 1:
		return path.Clean(paths[0])
	}

	// Note, we treat string as []byte, not []rune as is often
	// done in Go. (And sep as byte, not rune). This is because
	// most/all supported OS' treat paths as string of non-zero
	// bytes. A filename may be displayed as a sequence of Unicode
	// runes (typically encoded as UTF-8) but paths are
	// not required to be valid UTF-8 or in any normalized form
	// (e.g. "é" (U+00C9) and "é" (U+0065,U+0301) are different
	// file names.
	c := []byte(path.Clean(paths[0]))

	// We add a trailing sep to handle the case where the
	// common prefix directory is included in the path list
	// (e.g. /home/user1, /home/user1/foo, /home/user1/bar).
	// path.Clean will have cleaned off trailing / separators with
	// the exception of the root directory, "/" (in which case we
	// make it "//", but this will get fixed up to "/" bellow).
	c = append(c, sep)

	// Ignore the first path since it's already in c
	for _, v := range paths[1:] {
		// Clean up each path before testing it
		v = path.Clean(v) + string(sep)

		// Find the first non-common byte and truncate c
		if len(v) < len(c) {
			c = c[:len(v)]
		}
		for i := 0; i < len(c); i++ {
			if v[i] != c[i] {
				c = c[:i]
				break
			}
		}
	}

	// Remove trailing non-separator characters and the final separator
	for i := len(c) - 1; i >= 0; i-- {
		if c[i] == sep {
			c = c[:i]
			break
		}
	}

	return string(c)
}

func GetFileOrDirSize(path string) (int64, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	if fileInfo.IsDir() {
		return DirSizeB(path + "/")
	}
	return fileInfo.Size(), nil
}

// getFileSize get file size by path(B)
func DirSizeB(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}

func MoveFile(sourcePath, destPath string) error {
	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("couldn't open source file: %s", err)
	}
	outputFile, err := os.Create(destPath)
	if err != nil {
		inputFile.Close()
		return fmt.Errorf("couldn't open dest file: %s", err)
	}
	defer outputFile.Close()
	_, err = io.Copy(outputFile, inputFile)
	inputFile.Close()
	if err != nil {
		return fmt.Errorf("writing to output file failed: %s", err)
	}
	err = os.Remove(sourcePath)
	if err != nil {
		return fmt.Errorf("failed removing original file: %s", err)
	}
	return nil
}

func FindFirstFile(root string, filename string) string {
	var result string

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return filepath.SkipDir
		}

		if info.Name() == filename {
			result = path
			return errors.New("stop walking")
		}
		return nil
	})
	return result
}

func IsDirEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

// about os release
const EtcOsRelease string = "/etc/os-release"
const UsrLibOsRelease string = "/usr/lib/os-release"

// Read and return os-release, trying EtcOsRelease, followed by UsrLibOsRelease.
// err will contain an error message if neither file exists or failed to parse
func ReadOSRelease() (osrelease map[string]string, err error) {
	osrelease, err = ReadFile(EtcOsRelease)
	if err != nil {
		osrelease, err = ReadFile(UsrLibOsRelease)
	}
	return
}
func ReadFile(filename string) (osrelease map[string]string, err error) {
	osrelease = make(map[string]string)
	err = nil

	lines, err := ParseFile(filename)
	if err != nil {
		return
	}

	for _, v := range lines {
		key, value, err := ParseLine(v, "=")
		if err == nil {
			osrelease[key] = value
		}
	}
	return
}
func ParseFile(filename string) (lines []string, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
func ParseLine(line string, sep string) (key string, value string, err error) {
	err = nil

	// skip empty lines
	if len(line) == 0 {
		err = errors.New("skipping: zero-length")
		return
	}

	// skip comments
	if line[0] == '#' {
		err = errors.New("skipping: comment")
		return
	}

	// try to split string at the first '='
	splitString := strings.SplitN(line, sep, 2)
	if len(splitString) != 2 {
		err = errors.New("can not extract key=value")
		return
	}

	// trim white space from key and value
	key = splitString[0]
	key = strings.Trim(key, " ")
	value = splitString[1]
	value = strings.Trim(value, " ")

	// Handle double quotes
	if strings.ContainsAny(value, `"`) {
		first := string(value[0:1])
		last := string(value[len(value)-1:])

		if first == last && strings.ContainsAny(first, `"'`) {
			value = strings.TrimPrefix(value, `'`)
			value = strings.TrimPrefix(value, `"`)
			value = strings.TrimSuffix(value, `'`)
			value = strings.TrimSuffix(value, `"`)
		}
	}
	key = strings.Replace(key, "\t", "", -1)

	// expand anything else that could be escaped
	value = strings.Replace(value, `\"`, `"`, -1)
	value = strings.Replace(value, `\$`, `$`, -1)
	value = strings.Replace(value, `\\`, `\`, -1)
	value = strings.Replace(value, "\\`", "`", -1)
	return
}
func NameAccumulation(path string) string {
	dir, file := filepath.Split(path)
	ext := filepath.Ext(file)
	base := file[0 : len(file)-len(ext)]
	for i := 1; ; i++ {
		newPath := filepath.Join(dir, fmt.Sprintf("%s(%d)%s", base, i, ext))
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
	}
}
