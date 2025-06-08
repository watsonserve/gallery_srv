-- CREATE USER res WITH PASSWORD '$1';
CREATE DATABASE res;
GRANT ALL PRIVILEGES ON DATABASE res TO res;
CREATE TABLE IF NOT EXISTS res_thumb (
    id SERIAL PRIMARY KEY,
    hash char(40) UNIQUE,
    ext text,
    raw text UNIQUE,
    size int,
    rtime TIMESTAMP,
    ctime TIMESTAMP DEFAULT now()
);
GRANT ALL PRIVILEGES ON TABLE res_thumb TO res;
GRANT ALL PRIVILEGES ON SEQUENCE res_thumb_id_seq TO res;
CREATE INDEX res_ctime_index ON res_thumb(ctime);
CREATE INDEX res_rtime_index ON res_thumb(rtime);

-- select floor(EXTRACT(epoch from ctime)) as ctime from res_thumb;
