-- Legacy inline media remains readable until the next workspace write moves it here.
-- Public/read-only startup does not rewrite the existing workspace payload.
CREATE TABLE catalog_media (
    id text PRIMARY KEY CHECK (id <> ''),
    data bytea NOT NULL CHECK (octet_length(data) > 0)
);
