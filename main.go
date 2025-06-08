package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/watsonserve/filed/action"
	"github.com/watsonserve/goengine"
	"github.com/watsonserve/goutils"
)

func main() {
	addr := os.Args[1]
	conf, err := goutils.GetConf("/etc/meta.conf")
	if nil != err {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}

	dbConn := goengine.ConnPg(&goengine.DbConf{
		User:   conf["db_user"][0],
		Passwd: conf["db_passwd"][0],
		Host:   conf["db_host"][0],
		Name:   conf["db_name"][0],
		Port:   conf["db_port"][0],
	})
	rootDir := conf["root"][0]
	fmt.Printf("root: %s\n", rootDir)
	d := action.NewAction(dbConn, rootDir)
	p := action.NewPictureAction(dbConn, rootDir)

	router := goengine.InitHttpRoute()
	router.Set("/Pictures/", p.ServeHTTP)
	router.SetWith("^/.*", d.ServeHTTP)
	engine := goengine.New(router, nil)
	http.ListenAndServe(addr, engine)
}
