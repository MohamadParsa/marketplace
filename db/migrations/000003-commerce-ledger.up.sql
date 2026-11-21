ALTER TABLE review_workspace ADD COLUMN data_origin text NOT NULL DEFAULT 'review' CHECK (data_origin IN ('review','live'));

CREATE TABLE commerce_wallet_entries (
 id text PRIMARY KEY CHECK (id <> ''),
 user_id text NOT NULL REFERENCES review_users(id),
 source text NOT NULL UNIQUE CHECK (source <> ''),
 amount bigint NOT NULL CHECK (amount <> 0 AND amount BETWEEN -9000000000000000 AND 9000000000000000),
 kind text NOT NULL CHECK (kind IN ('deposit-credit','order-payment','manual-refund','order-cancel-credit')),
 created_at timestamptz NOT NULL,
 fingerprint text NOT NULL,
 CHECK ((kind IN ('deposit-credit','order-cancel-credit') AND amount > 0) OR (kind IN ('order-payment','manual-refund') AND amount < 0))
);
CREATE INDEX commerce_wallet_entries_owner_date ON commerce_wallet_entries(user_id,created_at);

CREATE TABLE commerce_deposits (
 id text PRIMARY KEY CHECK (id <> ''),
 user_id text NOT NULL REFERENCES review_users(id),
 reference_key text NOT NULL UNIQUE,
 verified_key text,
 status text NOT NULL CHECK (status IN ('pending','approved','rejected')),
 amount bigint NOT NULL CHECK (amount > 0 AND amount <= 9000000000000000),
 fingerprint text NOT NULL,
 UNIQUE(id,user_id),
 CHECK ((status='approved' AND verified_key IS NOT NULL) OR (status<>'approved' AND verified_key IS NULL))
);
CREATE UNIQUE INDEX commerce_verified_deposit_once ON commerce_deposits(verified_key) WHERE verified_key IS NOT NULL;

CREATE TABLE commerce_receipts (
 id text PRIMARY KEY CHECK (id <> ''),
 deposit_id text NOT NULL,
 user_id text NOT NULL,
 data bytea NOT NULL CHECK (octet_length(data) > 0 AND octet_length(data) <= 4194304),
 fingerprint text NOT NULL,
 FOREIGN KEY(deposit_id,user_id) REFERENCES commerce_deposits(id,user_id)
);
