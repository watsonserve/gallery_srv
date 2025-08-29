-- CREATE USER res WITH PASSWORD '$1';
CREATE DATABASE res;
GRANT ALL PRIVILEGES ON DATABASE res TO res;

CREATE TABLE IF NOT EXISTS res_thumb (
    pid uuid PRIMARY KEY,
    hash char(64) UNIQUE,
    ext char[16],
    raw text UNIQUE,
    size int DEFAULT 0
);

CREATE TABLE IF NOT EXISTS res_user_img (
    id SERIAL PRIMARY KEY,
    uid uuid,
    pid uuid,
    rtime int DEFAULT 0,
    ctime int
);

GRANT ALL PRIVILEGES ON TABLE res_thumb TO res;
GRANT ALL PRIVILEGES ON SEQUENCE res_thumb_id_seq TO res;
CREATE INDEX res_ctime_index ON res_thumb(ctime);
CREATE INDEX res_rtime_index ON res_thumb(rtime);

-- select floor(EXTRACT(epoch from ctime)) as ctime from res_thumb;
