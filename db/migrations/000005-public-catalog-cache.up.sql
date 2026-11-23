-- Rebuildable, private read models and durable cache versions. Business IDs remain unchanged.
CREATE TABLE catalog_public_state (
 singleton boolean PRIMARY KEY DEFAULT true CHECK(singleton),
 namespace text NOT NULL DEFAULT md5(random()::text || clock_timestamp()::text),
 generation bigint NOT NULL DEFAULT 1,
 dependency_revision bigint NOT NULL DEFAULT 1,
 dependency_hash text NOT NULL,
 payload jsonb NOT NULL,
 next_expiry timestamptz
);
CREATE SEQUENCE catalog_public_revision_seq;
CREATE TABLE catalog_public_products (
 id text PRIMARY KEY,
 revision bigint NOT NULL DEFAULT nextval('catalog_public_revision_seq'),
 fingerprint text NOT NULL,
 payload jsonb NOT NULL
);
CREATE TABLE catalog_public_media (id text PRIMARY KEY);
CREATE TABLE catalog_public_banner_media (id text PRIMARY KEY,data bytea NOT NULL,width integer NOT NULL,height integer NOT NULL);
