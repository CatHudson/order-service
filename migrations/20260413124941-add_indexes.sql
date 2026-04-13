-- +migrate Up
CREATE INDEX idx_orders_status ON orders (status);
CREATE INDEX idx_orders_account_id_created_at ON orders (account_id, created_at);

-- +migrate Down
DROP INDEX IF EXISTS idx_orders_account_id_created_at;
DROP INDEX IF EXISTS idx_orders_status;
