-- Additive source metadata; historical wallet entries and fingerprints stay intact.
ALTER TABLE commerce_wallet_entries DROP CONSTRAINT commerce_wallet_entries_kind_check;
ALTER TABLE commerce_wallet_entries DROP CONSTRAINT commerce_wallet_entries_check;
ALTER TABLE commerce_wallet_entries ADD CONSTRAINT commerce_wallet_entries_kind_check
 CHECK (kind IN ('deposit-credit','order-payment','manual-refund','order-cancel-credit','admin-credit'));
ALTER TABLE commerce_wallet_entries ADD CONSTRAINT commerce_wallet_entries_check
 CHECK ((kind IN ('deposit-credit','order-cancel-credit','admin-credit') AND amount>0)
 OR (kind IN ('order-payment','manual-refund') AND amount<0));
CREATE TABLE commerce_credit_lots (
 id text PRIMARY KEY,
 user_id text NOT NULL REFERENCES review_users(id),
 entry_id text REFERENCES commerce_wallet_entries(id),
 origin_lot_id text REFERENCES commerce_credit_lots(id) DEFERRABLE INITIALLY DEFERRED,
 amount bigint NOT NULL CHECK (amount>0 AND amount<=9000000000000000),
 spent bigint NOT NULL CHECK (spent>=0 AND spent<=amount),
 withdrawable boolean NOT NULL,
 fingerprint text NOT NULL
);
CREATE INDEX commerce_credit_lots_owner ON commerce_credit_lots(user_id);
