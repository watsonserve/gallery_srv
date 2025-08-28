package services

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/watsonserve/filed/helper"
)

type FileService struct {
	rootPath string
}

func NewFileService(dbConn *sql.DB, root string) *FileService {
	return &FileService{
		rootPath: path.Clean(root),
	}
}

func (d *FileService) SendFile(resp http.ResponseWriter, req *http.Request) {
	absPath := path.Clean(path.Join(d.rootPath, req.URL.Path))
	fp, err := os.Open(absPath)
	if nil != err {
		StdJSONResp(resp, nil, http.StatusNotFound, err.Error())
		return
	}
	defer fp.Close()

	meta, err := helper.GetMeta(fp)
	if nil != err {
		StdJSONResp(resp, nil, http.StatusBadRequest, err.Error())
		return
	}

	respHeader := resp.Header()
	respHeader.Set("Content-Type", meta.ContentType)
	respHeader.Set("Content-Length", fmt.Sprintf("%d", meta.Size))
	respHeader.Set("Content-Digest", fmt.Sprintf("sha-256=:%s:", meta.Sha256Hash))
	respHeader.Set("Last-Modified", meta.ModTime.String())
	if http.MethodHead == req.Method {
		resp.Write(nil)
		return
	}
	io.Copy(resp, fp)
}

func (d *FileService) Upload(resp http.ResponseWriter, req *http.Request) {
	reqHeader := &req.Header
	cType := strings.Split(reqHeader.Get("Content-Type"), ";")[0]
	if !strings.HasPrefix(cType, "image/") {
		StdJSONResp(resp, nil, http.StatusUnsupportedMediaType, "Accept Image Only")
		return
	}
	origin := helper.GetOrigin(reqHeader)
	if nil == origin {
		StdJSONResp(resp, nil, http.StatusBadRequest, "Header Origin Not Found")
		return
	}
	// hash := helper.GetDigest(reqHeader, "sha-256")
	// if "" == hash {
	// 	StdJSONResp(resp, nil, http.StatusBadRequest, "Content-Digest sha-256 Required")
	// 	return
	// }
	absPath := path.Clean(path.Join(d.rootPath, req.URL.Path))
	fp, err := os.OpenFile(absPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0660)
	if nil != err {
		StdJSONResp(resp, nil, http.StatusNotFound, err.Error())
		return
	}
	defer fp.Close()
	io.Copy(fp, req.Body)
	origin.Path = req.URL.Path[4:]
	resp.Header().Set("Location", origin.String())
	StdJSONResp(resp, nil, http.StatusSeeOther, "")
}

func (d *FileService) GenPreview(resp http.ResponseWriter, req *http.Request) {

	StdJSONResp(resp, nil, http.StatusCreated, "")
}

func (d *FileService) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodHead:
		fallthrough
	case http.MethodGet:
		d.SendFile(resp, req)
		return
	case http.MethodPut:
		d.Upload(resp, req)
		return
	case http.MethodPost:
		d.GenPreview(resp, req)
		return
	default:
	}
	StdJSONResp(resp, nil, http.StatusMethodNotAllowed, "")
}
