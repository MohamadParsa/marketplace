-- Derived sales input, without customer/order/payment identifiers. Financial state stays authoritative.
CREATE TABLE catalog_sales_state (
 singleton boolean PRIMARY KEY DEFAULT true CHECK(singleton),
 revision bigint NOT NULL DEFAULT 1,
 fingerprint text NOT NULL,
 payload jsonb NOT NULL
);
