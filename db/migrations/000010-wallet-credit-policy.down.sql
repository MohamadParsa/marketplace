-- A rollback must never silently discard credit restrictions or financial history.
DO $$ BEGIN
 IF EXISTS(SELECT 1 FROM commerce_credit_lots) OR EXISTS(SELECT 1 FROM commerce_wallet_entries WHERE kind='admin-credit') THEN
  RAISE EXCEPTION 'Credit policy has financial data; use a forward migration';
 END IF;
END $$;
DROP TABLE commerce_credit_lots;
ALTER TABLE commerce_wallet_entries DROP CONSTRAINT commerce_wallet_entries_kind_check;
ALTER TABLE commerce_wallet_entries DROP CONSTRAINT commerce_wallet_entries_check;
ALTER TABLE commerce_wallet_entries ADD CONSTRAINT commerce_wallet_entries_kind_check
 CHECK (kind IN ('deposit-credit','order-payment','manual-refund','order-cancel-credit'));
ALTER TABLE commerce_wallet_entries ADD CONSTRAINT commerce_wallet_entries_check
 CHECK ((kind IN ('deposit-credit','order-cancel-credit') AND amount>0)
 OR (kind IN ('order-payment','manual-refund') AND amount<0));
