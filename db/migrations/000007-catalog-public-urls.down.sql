-- Export mappings before intentional rollback; never reassign previously issued codes.
DROP TABLE catalog_category_urls;
DROP TABLE catalog_product_codes;
DROP SEQUENCE catalog_product_code_seq;
