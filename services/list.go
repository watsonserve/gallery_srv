package services

import (
	"net/http"
	"os"
	"path"
	"time"

	"github.com/watsonserve/filed/dao"
	"github.com/watsonserve/filed/helper"
)

type ListService struct {
	raw string
	dbi *dao.DBI
}

func NewListService(dbi *dao.DBI, root string) *ListService {
	return &ListService{
		raw: path.Clean(path.Join(root, "raw")),
		dbi: dbi,
	}
}

func (d *ListService) List(resp http.ResponseWriter, req *http.Request) {
	uid := helper.GetUid(req)
	if "" == uid {
		StdJSONResp(resp, nil, http.StatusUnauthorized, "")
		return
	}
	rangeList := helper.GetRange(&req.Header)
	list, err := d.dbi.List(uid, rangeList)

	if nil != err {
		StdJSONResp(resp, nil, http.StatusServiceUnavailable, err.Error())
		return
	}
	StdJSONResp(resp, list, 0, "")
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
	default:
	}
	StdJSONResp(resp, nil, http.StatusMethodNotAllowed, "")
}
