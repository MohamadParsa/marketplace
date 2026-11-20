CREATE TABLE review_workspace (
    singleton boolean PRIMARY KEY CHECK (singleton),
    payload jsonb NOT NULL CHECK (jsonb_typeof(payload) = 'object')
);

CREATE TABLE review_users (
    id text PRIMARY KEY CHECK (id <> ''),
    phone text NOT NULL UNIQUE CHECK (phone ~ '^09[0-9]{9}$'),
    name text NOT NULL,
    base_role text NOT NULL CHECK (base_role IN ('', 'super-admin', 'operations', 'support', 'finance', 'viewer')),
    active boolean NOT NULL,
    access_version integer NOT NULL CHECK (access_version >= 0)
);

CREATE TABLE review_stores (
    id text PRIMARY KEY CHECK (id <> '')
);

CREATE TABLE review_roles (
    id text PRIMARY KEY CHECK (id <> ''),
    name text NOT NULL CHECK (btrim(name) <> ''),
    scope text NOT NULL CHECK (scope IN ('global', 'store')),
    permissions text[] NOT NULL CHECK (cardinality(permissions) > 0),
    version integer NOT NULL CHECK (version > 0),
    UNIQUE (id, scope)
);

CREATE TABLE review_memberships (
    user_id text NOT NULL REFERENCES review_users (id),
    store_id text NOT NULL REFERENCES review_stores (id),
    PRIMARY KEY (user_id, store_id)
);

CREATE UNIQUE INDEX review_roles_scope_name ON review_roles (scope, lower(name));

CREATE TABLE review_role_grants (
    user_id text NOT NULL REFERENCES review_users (id),
    role_id text NOT NULL,
    store_id text,
    scope text GENERATED ALWAYS AS (CASE WHEN store_id IS NULL THEN 'global' ELSE 'store' END) STORED,
    FOREIGN KEY (role_id, scope) REFERENCES review_roles (id, scope),
    FOREIGN KEY (user_id, store_id) REFERENCES review_memberships (user_id, store_id),
    UNIQUE NULLS NOT DISTINCT (user_id, role_id, store_id)
);

CREATE INDEX review_role_grants_role ON review_role_grants (role_id);
CREATE INDEX review_memberships_store ON review_memberships (store_id);

CREATE TABLE review_purchase_owners (
    purchase_id text PRIMARY KEY CHECK (purchase_id <> ''),
    user_id text NOT NULL REFERENCES review_users (id)
);

CREATE INDEX review_purchase_owners_user ON review_purchase_owners (user_id);
