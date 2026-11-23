-- Retained mappings are not foreign-key cascaded: deleted codes/aliases stay reserved.
CREATE SEQUENCE catalog_product_code_seq START WITH 1000;
CREATE TABLE catalog_product_codes (
 product_id text PRIMARY KEY,
 number bigint NOT NULL UNIQUE CHECK(number>0),
 public_code text GENERATED ALWAYS AS ('chr-' || number::text) STORED UNIQUE
);
CREATE TABLE catalog_category_urls (
 token text PRIMARY KEY CHECK(length(token) BETWEEN 1 AND 128),
 category_id text NOT NULL
);
CREATE INDEX catalog_category_urls_owner ON catalog_category_urls(category_id);
