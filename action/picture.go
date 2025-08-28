package action

import (
	"fmt"
	"net/http"

	"github.com/watsonserve/filed/services"
)

type PictureAction struct {
	listSrv http.Handler
	dav     http.Handler
}

var imgCache = map[string]bool{"thumb": true, "preview": true, "raw": true}

func NewPictureAction(listSrv http.Handler, fileSrv http.Handler) *PictureAction {
	return &PictureAction{
		listSrv: listSrv,
		dav:     fileSrv,
	}
}

func (d *PictureAction) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	subPath := req.URL.Path[9:]

	if "/" == subPath {
		d.listSrv.ServeHTTP(resp, req)
		return
	}

	query := req.URL.Query()
	lev := query.Get("lev")
	if "" == lev {
		lev = "raw"
	}
	if !imgCache[lev] {
		services.StdJSONResp(resp, nil, http.StatusNotFound, "")
		return
	}
	if "raw" != lev && http.MethodGet != req.Method && http.MethodHead != req.Method {
		services.StdJSONResp(resp, nil, http.StatusMethodNotAllowed, "")
		return
	}
	req.URL.Path = fmt.Sprintf("/%s%s", lev, subPath)
	d.dav.ServeHTTP(resp, req)
}
