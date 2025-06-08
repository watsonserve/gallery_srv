package action

import (
	"database/sql"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/watsonserve/goengine"
)

type Root struct {
	prefix string
}

func NewRoot(root string) *Root {
	return &Root{
		prefix: root,
	}
}

func (r *Root) Open(name string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(r.prefix+name, flag, perm)
}

type Action struct {
	tmp         *Root
	root        http.Handler
	preview     http.Handler
	thumb       http.Handler
	imgCacheRgx *regexp.Regexp
	dao         *goengine.DAO
}

func NewAction(dbConn *sql.DB, root string) *Action {
	return &Action{
		tmp:         NewRoot(root + "/tmp"),
		root:        http.FileServer(http.Dir(root + "/raw")),
		preview:     http.FileServer(http.Dir(root + "/preview")),
		thumb:       http.FileServer(http.Dir(root + "/thumb")),
		imgCacheRgx: regexp.MustCompile(`^/\.(thumb|preview)/(.+)`),
		dao:         nil,
	}
}

func (d *Action) file(resp http.ResponseWriter, req *http.Request) {
	absPath := req.URL.Path
	matcher := d.imgCacheRgx.FindAllString(absPath, 2)
	if 2 < len(matcher) {
		switch matcher[1] {
		case "thumb":
			d.thumb.ServeHTTP(resp, req)
		case "preview":
			d.preview.ServeHTTP(resp, req)
		default:
		}
	}
	d.root.ServeHTTP(resp, req)
}

func (d *Action) signPut(resp http.ResponseWriter, req *http.Request) {
	reqHeader := req.Header
	hash := getDigestSHA1(reqHeader.Get("Repr-Digest"))
	if "" == hash {
		StdJSONResp(resp, nil, http.StatusBadRequest, "Repr-Digest Not Found")
		return
	}
	offset, length, err := getPart(reqHeader.Get("Range"))
	if nil != err {
		StdJSONResp(resp, nil, http.StatusBadRequest, err.Error())
		return
	}
	fp, err := d.tmp.Open(hash, os.O_RDWR|os.O_CREATE, 0660)
	if nil != err {
		StdJSONResp(resp, nil, http.StatusServiceUnavailable, err.Error())
		return
	}
	defer fp.Close()
	if length < 1 {
		checkHash(resp, hash, fp)
		return
	}
	code, msg := write(fp, offset, req)
	StdJSONResp(resp, nil, code, msg)
}

func (d *Action) put(resp http.ResponseWriter, req *http.Request) {
	reqHeader := req.Header
	contentType := strings.Split(reqHeader.Get("Content-Type"), ";")
	switch contentType[0] {
	case "":
		StdJSONResp(resp, nil, http.StatusBadRequest, "")
		return
	case "multipart/form-data":
		return
	default:
	}
	d.signPut(resp, req)
	// hash := goutils.SHA1(dstFileName)
	// imghelper.IMRead(dstFileName)
}

func (d *Action) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		d.file(resp, req)
		return
	case http.MethodPut:
		d.put(resp, req)
		return
	default:
	}
	StdJSONResp(resp, nil, http.StatusMethodNotAllowed, "")
}
