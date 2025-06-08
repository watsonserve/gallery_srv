package action

import (
	"crypto/sha1"
	"database/sql"
	"errors"
	"io"
	"mime"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/watsonserve/goengine"
	"github.com/watsonserve/goutils"
)

type PictureAction struct {
	dir string
	dao *goengine.DAO
}

func NewPictureAction(dbConn *sql.DB, root string) *PictureAction {
	dao := goengine.InitDAO(dbConn)
	dao.Prepare("real_name", "SELECT raw FROM res_thumb WHERE hash=$1")
	dao.Prepare("list", "SELECT hash FROM res_thumb WHERE rtime IS NULL ORDER BY ctime DESC OFFSET $1 LIMIT $2")
	dao.Prepare("delt", "UPDATE res_thumb SET rtime=$2 WHERE hash=$1 AND rtime IS NOT NULL")
	dao.Prepare("drop", "DELETE FROM res_thumb WHERE hash=$1 AND rtime IS NOT NULL")
	dao.Prepare("inst", "INSERT INTO res_thumb (hash, ext, size, raw) VALUES ($1, $2, $3, $4)")

	return &PictureAction{
		dir: root,
		dao: dao,
	}
}

func (d *PictureAction) list(resp http.ResponseWriter, req *http.Request) {
	// bytes=200-1000, 2000-6576, 19000-
	rangeVal := strings.Split(req.Header.Get("Range"), "=")
	strRange := "0-"
	if 2 == len(rangeVal) {
		strRange = rangeVal[1]
	}
	rangeList := strings.Split(strRange, ",")
	if 1 < len(rangeList) {
		StdJSONResp(resp, "[]", http.StatusRequestedRangeNotSatisfiable, "multipart is not be allowed")
	}
	sep := strings.Split(rangeList[0], "-")
	offset, err := strconv.ParseInt(sep[0], 10, 32)
	if nil != err || offset < 0 {
		offset = 0
	}
	end, err := strconv.ParseInt(sep[1], 10, 32)
	if nil != err || end < offset {
		end = -1 // infinity
	}
	length := end - offset
	if length < 1 || 1000 < length {
		StdJSONResp(resp, "[]", http.StatusRequestedRangeNotSatisfiable, "max 1000 rows by once query")
		return
	}
	rows, err := d.dao.StmtMap["list"].Query(offset, length)
	if nil != err {
		StdJSONResp(resp, nil, http.StatusServiceUnavailable, err.Error())
		return
	}
	defer rows.Close()
	list := make([]string, 0)
	for rows.Next() {
		hash := ""
		err = rows.Scan(&hash)
		list = append(list, hash)
	}
	StdJSONResp(resp, list, 0, "")
}

func (d *PictureAction) multiPut(resp http.ResponseWriter, req *http.Request) {
	err := req.ParseMultipartForm(4 << 20)
	if nil != err {
		StdJSONResp(resp, nil, http.StatusBadRequest, err.Error())
		return
	}
	uFile, _, err := req.FormFile("file")
	if nil != err {
		StdJSONResp(resp, nil, http.StatusBadRequest, err.Error())
		return
	}
	defer uFile.Close()
	dstFileName := goutils.RandomString(32)
	dst, err := os.OpenFile(dstFileName, os.O_CREATE, 0660)
	if nil != err {
		StdJSONResp(resp, nil, http.StatusServiceUnavailable, err.Error())
		return
	}
	defer dst.Close()
	_, err = io.Copy(dst, uFile)
	if nil != err {
		StdJSONResp(resp, nil, http.StatusServiceUnavailable, err.Error())
		return
	}
}

func getDigestSHA1(s string) string {
	reprDigest := strings.Split(s, ",")
	hashSHA1 := ""
	for _, item := range reprDigest {
		kv := strings.Split(strings.TrimSpace(item), "=")
		if len(kv) < 2 {
			continue
		}
		key := strings.ToLower(kv[0])
		if "sha" == key || "sha-1" == key || "sha1" == key {
			hashSHA1 = kv[1]
			break
		}
	}
	return hashSHA1
}

func getPart(s string) (int64, int64, error) {
	rangeInfo := strings.Split(s, "=")
	if len(rangeInfo) < 2 {
		return 0, 0, nil
	}
	rangeVals := strings.Split(rangeInfo[1], ",")
	if 1 < len(rangeVals) {
		return -1, 0, errors.New("range overflow")
	}
	oe := strings.Split(rangeVals[0], "-")
	off, err := strconv.ParseInt(oe[0], 10, 32)
	if nil != err {
		return -1, 0, err
	}
	end, err := strconv.ParseInt(oe[1], 10, 32)
	if nil != err {
		return -1, 0, err
	}
	return off, end - off, nil
}

func checkHash(resp http.ResponseWriter, hash string, fp *os.File) {
	content, err := io.ReadAll(fp)
	if nil != err {
		StdJSONResp(resp, nil, http.StatusServiceUnavailable, err.Error())
		return
	}
	_hash := sha1.Sum(content)
	StdJSONResp(resp, string(_hash[:]) == hash, 0, "")
}

func write(fp *os.File, offset int64, req *http.Request) (int, string) {
	var err error = nil
	if 0 < offset {
		_, err = fp.Seek(offset, 0)
	}
	if nil == err {
		_, err = io.Copy(fp, req.Body)
	}
	if nil != err {
		return http.StatusBadRequest, err.Error()
	}
	return 0, ""
}

func (d *PictureAction) put(resp http.ResponseWriter, req *http.Request) {
	reqHeader := req.Header
	cType := strings.Split(reqHeader.Get("Accept-Type"), ";")[0]
	if !strings.HasPrefix(cType, "image/") {
		StdJSONResp(resp, nil, http.StatusForbidden, "Accept Image Only")
		return
	}
	exts, err := mime.ExtensionsByType(cType)
	if nil != err || len(exts) < 1 || "" == exts[0] {
		StdJSONResp(resp, nil, http.StatusForbidden, "Format Can Not Be Accpeted")
		return
	}
	hash := getDigestSHA1(reqHeader.Get("Repr-Digest"))
	if "" == hash {
		StdJSONResp(resp, nil, http.StatusBadRequest, "Repr-Digest Not Found")
		return
	}
	err = d.dao.StmtMap["real_name"].QueryRow(hash).Scan()
	if nil == err {
		StdJSONResp(resp, nil, http.StatusNotModified, "")
		return
	}
	absName := d.dir + "/" + hash + exts[0]
	fp, err := os.OpenFile(absName, os.O_WRONLY|os.O_CREATE, 0660)
	if nil != err {
		StdJSONResp(resp, nil, http.StatusServiceUnavailable, err.Error())
		return
	}
	fp.Close()
	resp.Header().Set("Location", absName)
	StdJSONResp(resp, nil, http.StatusCreated, "")
}

func (d *PictureAction) delt(resp http.ResponseWriter, req *http.Request) {
	hash := req.URL.Path[1:]
	_, err := d.dao.StmtMap["delt"].Exec(hash, time.Now().Unix())
	if nil != err {
		StdJSONResp(resp, nil, http.StatusBadRequest, err.Error())
		return
	}

	StdJSONResp(resp, nil, 0, "")
}

func (d *PictureAction) drop(resp http.ResponseWriter, req *http.Request) {
	hash := req.URL.Path[1:]
	raw := ""
	err := d.dao.StmtMap["real_name"].QueryRow(hash).Scan(raw)
	if nil != err {
		StdJSONResp(resp, nil, http.StatusBadRequest, err.Error())
		return
	}
	_, err = d.dao.StmtMap["drop"].Exec(hash)
	if nil != err {
		StdJSONResp(resp, nil, http.StatusBadRequest, err.Error())
		return
	}
	os.Remove("/thumb/" + hash + ".webp")
	os.Remove("/preview/" + hash + ".webp")
	os.Remove(raw)
	StdJSONResp(resp, nil, 0, "")
}

func (d *PictureAction) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		d.list(resp, req)
		return
	case http.MethodPut:
		d.put(resp, req)
		return
	default:
	}
	StdJSONResp(resp, nil, http.StatusMethodNotAllowed, "")
}
