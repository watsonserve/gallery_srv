## 上传文件

```
PUT /Pictures/foo.cr2 HTTP/1.1
Content-Type: image/cr2
Origin: https://store.watsonserve.com
Content-Digest: sha-256=:abcdefg...:
If-Match: "uuid1234..."
Expect: 100-continue
Content-Length: 1000000
Cookie: abc=def
```

```
func NewFile(req) {
    eTagVal = GenUUID()
    fp = createFile(eTagVal)
    io.Copy(fp, req.Body)
    hash = sha256(fp)
    fp.Close()
    check(hash, ContentDigest)
    genPreview(fp)
    return { eTagVal, hash }
}


func ServeHTTP(resp, req) {
    found := CheckExist(req.URL.Path)
    eTag := req.Header.Get("If-Match")

    if (!found && !eTag) {
        { eTagVal, hash } = NewFile(req)
        res_thumb.Insert(eTagVal, hash, path.Ext(req.Path), ContentLength)
        res_user_img.Insert(uid, req.Path, eTagVal)
        resp.WriteHeader(201)
        return
    }

    if (found && !eTag) {
        resp.WriteHeader(400)
        return
    }
    if (!found && eTag) {
        resp.WriteHeader(404)
        return
    }

    if (found && eTag) {
        if isMatch(eTag) {
            { eTagVal, hash } = NewFile(req)
            res_thumb.Update(eTagVal, hash, path.Ext(req.Path), ContentLength)
            res_user_img.Update(uid, req.Path, eTagVal)
            resp.WriteHeader(201)
        } else {
            resp.WriteHeader(400)
        }
        return
    }

    exist = res_user_img.Query(uid, req.Path, eTagVal)
    if exist {
        { eTagVal, hash } = NewFile(req)

    }
}
```