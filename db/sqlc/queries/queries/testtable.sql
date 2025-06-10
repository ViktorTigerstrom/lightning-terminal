-- name: InsertTesttable :one
INSERT INTO testtable (label, alias)
VALUES ($1, $2)
    RETURNING id;
