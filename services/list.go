package services

import (
	"net/http"
	"path"

	"github.com/watsonserve/galleried/dao"
	"github.com/watsonserve/galleried/helper"
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
	uid := helper.GetUid(req)
	fileName := helper.GetFileName(req.URL.Path)
	if "" == uid {
		StdJSONResp(resp, nil, http.StatusUnauthorized, "")
		return
	}
	err := d.dbi.Del(uid, fileName)
	if nil != err {
		StdJSONResp(resp, nil, http.StatusBadRequest, err.Error())
		return
	}

	StdJSONResp(resp, nil, 0, "")
}

func (d *ListService) drop(resp http.ResponseWriter, req *http.Request) {
	uid := helper.GetUid(req)
	fileName := helper.GetFileName(req.URL.Path)
	if "" == uid {
		StdJSONResp(resp, nil, http.StatusUnauthorized, "")
		return
	}
	err := d.dbi.Drop(uid, fileName)
	if nil != err {
		StdJSONResp(resp, nil, http.StatusBadRequest, err.Error())
		return
	}

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
