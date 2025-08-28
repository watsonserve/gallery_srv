package helper

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

func GenUUIDStr() (string, error) {
	var buf [32]byte
	__uuid, err := uuid.NewV7()
	if nil != err {
		return "", err
	}
	hex.Encode(buf[:], __uuid[:])
	return string(buf[:]), nil
}

/**
 * @return baseName
 */
func CreateNewFile(dir, ext string, perm os.FileMode) (string, error) {
	if 0 < len(ext) && '.' != ext[0] {
		ext = "." + ext
	}
	if "" == dir || '/' != dir[len(dir)-1] {
		dir += "/"
	}
	for i := 0; i < 16; i++ {
		baseName, err := GenUUIDStr()
		if nil != err {
			return "", err
		}
		dst, err := os.OpenFile(path.Join(dir, baseName+ext), os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm)
		if nil == err {
			dst.Close()
			return baseName, err
		}
		if !os.IsExist(err) {
			return "", err
		}
	}
	return "", errors.New("retry timeout")
}

func GetOrigin(header *http.Header) *url.URL {
	strURL := strings.TrimSpace(header.Get("Origin"))
	if "" != strURL {
		uri, err := url.Parse(strURL)
		if nil == err {
			return uri
		}
	}

	strURL = strings.TrimSpace(header.Get("Referer"))
	if "" != strURL {
		uri, err := url.Parse(strURL)
		if nil == err {
			return &url.URL{
				Scheme: uri.Scheme,
				Host:   uri.Host,
			}
		}
	}

	strURL = strings.TrimSpace(header.Get("Host"))
	if "" == strURL {
		return nil
	}
	return &url.URL{
		Scheme: "https:",
		Host:   strURL,
	}
}

func GetExtNameByReq(header *http.Header) ([]string, error) {
	contentType := strings.Split(header.Get("Content-Type"), ";")[0]
	return mime.ExtensionsByType(contentType)
}

func GetContentLength(header *http.Header) int64 {
	contentLength, err := strconv.ParseInt(header.Get("Content-Length"), 10, 64)
	if nil != err {
		return -1
	}
	return contentLength
}

type Segment struct {
	Start int32
	End   int32
}

func GetRange(header *http.Header) []Segment {
	// bytes=200-1000, 2000-6576, 19000-
	rangeVal := strings.Split(header.Get("Range"), "=")
	if len(rangeVal) < 2 {
		return nil
	}
	rangeList := strings.Split(rangeVal[1], ",")
	results := make([]Segment, len(rangeList))

	for i, seg := range rangeList {
		sep := strings.Split(strings.TrimSpace(seg), "-")
		offset, err := strconv.ParseInt(sep[0], 10, 32)
		if nil != err || offset < 0 {
			offset = 0
		}
		end, err := strconv.ParseInt(sep[1], 10, 32)
		if nil != err || end < offset {
			end = -1 // infinity
		}
		results[i] = Segment{Start: int32(offset), End: int32(end)}
	}

	return results
}

func Sha256ByFile(fp *os.File) (string, error) {
	hasher := sha256.New()
	_, err := io.Copy(hasher, fp)
	if nil != err {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

type Meta struct {
	Size        int32
	ModTime     time.Time
	ContentType string
	Sha256Hash  string
}

func GetMeta(fp *os.File) (*Meta, error) {
	stat, err := fp.Stat()
	if nil != err {
		return nil, err
	}

	hash, err := Sha256ByFile(fp)
	if nil != err {
		return nil, err
	}
	return &Meta{
		Size:        int32(stat.Size()),
		ModTime:     stat.ModTime(),
		ContentType: mime.TypeByExtension(path.Ext(stat.Name())),
		Sha256Hash:  hash,
	}, nil
}

/**
 * digestKey: sha-1, sha-256, ...
 */
func GetDigest(header *http.Header, digestKey string) string {
	contentDigest := header.Get("Content-Digest")
	if "" == contentDigest {
		return ""
	}

	reprDigest := strings.Split(strings.ToLower(contentDigest), ",")
	for _, item := range reprDigest {
		kv := strings.Split(strings.TrimSpace(item), "=")
		if len(kv) < 2 {
			continue
		}
		key := kv[0]
		if digestKey == key {
			return kv[1]
		}
	}
	return ""
}

func Write(fp *os.File, offset int64, src io.Reader) error {
	var err error = nil
	if 0 < offset {
		_, err = fp.Seek(offset, 0)
	}
	if nil == err {
		_, err = io.Copy(fp, src)
	}
	return err
}
