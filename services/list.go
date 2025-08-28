package services

import (
	"database/sql"
	"errors"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/watsonserve/filed/helper"
	"github.com/watsonserve/goengine"
)

type ListService struct {
	raw string
	dao *goengine.DAO
}

func NewListService(dbConn *sql.DB, root string) *ListService {
	dao := goengine.InitDAO(dbConn)
	dao.Prepare("real_name", "SELECT raw FROM res_thumb WHERE hash=$1")
	dao.Prepare("list", "SELECT id FROM res_thumb WHERE rtime=0 ORDER BY ctime DESC OFFSET=$1")
	dao.Prepare("list_limit", "SELECT id FROM res_thumb WHERE rtime=0 ORDER BY ctime DESC OFFSET=$1 LIMIT=$2")
	dao.Prepare("delt", "UPDATE res_thumb SET rtime=$2 WHERE id=$1 AND rtime=0")
	dao.Prepare("drop", "DELETE FROM res_thumb WHERE id=$1 AND rtime<>0")
	dao.Prepare("inst", "INSERT INTO res_thumb (id, ext, size, raw) VALUES ($1, $2, $3, $4)")

	return &ListService{
		raw: path.Clean(path.Join(root, "raw")),
		dao: dao,
	}
}

func (d *ListService) selectList(rangeList []helper.Segment) (rows *sql.Rows, err error) {
	offset := int32(0)
	length := int32(0)

	if nil != rangeList {
		if 1 < len(rangeList) {
			return nil, errors.New("multipart is not be allowed")
		}
		sep := rangeList[0]
		offset = sep.Start
		if -1 != sep.End {
			length = sep.End - sep.Start
		}
	}
	if length < 1 {
		return d.dao.StmtMap["list"].Query(offset)
	}
	return d.dao.StmtMap["list_limit"].Query(offset, length)
}

func (d *ListService) List(resp http.ResponseWriter, req *http.Request) {
	rangeList := helper.GetRange(&req.Header)
	rows, err := d.selectList(rangeList)

	if nil != err {
		StdJSONResp(resp, nil, http.StatusServiceUnavailable, err.Error())
		return
	}
	list := make([]string, 0)
	for rows.Next() {
		pid := ""
		err = rows.Scan(&pid)
		list = append(list, pid)
	}
	StdJSONResp(resp, list, 0, "")
}

func (d *ListService) Put(resp http.ResponseWriter, req *http.Request) {
	reqHeader := &req.Header
	origin := helper.GetOrigin(reqHeader)
	if nil == origin {
		StdJSONResp(resp, nil, http.StatusBadRequest, "Header Origin Not Found")
		return
	}
	cLen := helper.GetContentLength(reqHeader)
	if -1 == cLen {
		StdJSONResp(resp, nil, http.StatusLengthRequired, "")
		return
	}
	if 0 != cLen {
		StdJSONResp(resp, nil, http.StatusRequestEntityTooLarge, "Any Content Not Accepted")
		return
	}
	extNames, err := helper.GetExtNameByReq(reqHeader)
	if nil != err {
		StdJSONResp(resp, nil, http.StatusUnsupportedMediaType, "")
		return
	}
	extName := extNames[0]

	baseName, err := helper.CreateNewFile(d.raw, extName, 0660)
	if nil != err {
		StdJSONResp(resp, nil, http.StatusServiceUnavailable, err.Error())
		return
	}
	_, err = d.dao.StmtMap["inst"].Exec(baseName, extName, baseName+extName)
	if nil != err {
		StdJSONResp(resp, nil, http.StatusBadRequest, err.Error())
		return
	}
	origin.Path = baseName + extName
	resp.Header().Set("Location", origin.String())
	StdJSONResp(resp, nil, http.StatusCreated, "")
}

func (d *ListService) delt(resp http.ResponseWriter, req *http.Request) {
	hash := req.URL.Path[1:]
	_, err := d.dao.StmtMap["delt"].Exec(hash, time.Now().Unix())
	if nil != err {
		StdJSONResp(resp, nil, http.StatusBadRequest, err.Error())
		return
	}

	StdJSONResp(resp, nil, 0, "")
}

func (d *ListService) drop(resp http.ResponseWriter, req *http.Request) {
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

func (d *ListService) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		d.List(resp, req)
		return
	case http.MethodPut:
		d.Put(resp, req)
		return
	default:
	}
	StdJSONResp(resp, nil, http.StatusMethodNotAllowed, "")
}
