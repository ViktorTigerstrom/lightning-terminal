CREATE TABLE IF NOT EXISTS testtable (
    -- The auto incrementing primary key.
    id BIGINT PRIMARY KEY,

    -- The ID that was used to identify the account in the legacy KVDB store.
    -- In order to avoid breaking the API, we keep this field here so that
    -- we can still look up accounts by this ID for the time being.
    alias BIGINT NOT NULL UNIQUE,

    -- An optional label to use for the account. If it is set, it must be
    -- unique.
    label TEXT UNIQUE
);