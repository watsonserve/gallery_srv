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

## configure
```
# pg_db
db_user=foo
db_passwd=bar
db_host=127.0.0.2
db_name=galleried_db
db_port=5432

#redis
redis_address=127.0.0.3
redis_password=

#session & cookie
sess_name=sess
cookie_prefix=galleried
session_prefix=galleried
domain=localhost

# files store
root=/home/you/pictures

# server
path_prefix=/Pictures
#listen=127.0.0.1:80
listen=:80
```