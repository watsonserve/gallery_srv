#! /bin/bash
# password=d80994a4ce6547a690de992d76f5bebd

docker run -d \
    --network=misaka \
    --ip=172.24.0.2 \
    -v /data/pg_data:/var/lib/postgresql/data \
    --restart=on-failure:3 \
    --name=pg \
    postgres:17

# docker run -it --rm --network=misaka postgres:17 psql -h 172.24.0.2 -U res -W res
