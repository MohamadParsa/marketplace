CREATE TABLE product_qa_questions (
    product_id text NOT NULL,
    id text NOT NULL,
    cursor bigint GENERATED ALWAYS AS IDENTITY UNIQUE,
    payload jsonb NOT NULL,
    status text GENERATED ALWAYS AS (payload->>'status') STORED NOT NULL,
    PRIMARY KEY (product_id, id),
    CHECK (status IN ('pending', 'approved', 'rejected'))
);
CREATE INDEX product_qa_questions_page ON product_qa_questions (product_id, cursor);
CREATE INDEX product_qa_questions_public_page ON product_qa_questions (product_id, cursor) WHERE status = 'approved';

CREATE TABLE product_qa_receipts (
    actor_id text NOT NULL,
    request_key text NOT NULL,
    fingerprint text NOT NULL,
    PRIMARY KEY (actor_id, request_key)
);
