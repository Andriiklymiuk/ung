-- Drop indexes
DROP INDEX IF EXISTS idx_tracking_sessions_contract;
DROP INDEX IF EXISTS idx_contracts_active;
DROP INDEX IF EXISTS idx_contracts_client;

-- Remove columns from tracking_sessions
ALTER TABLE tracking_sessions DROP COLUMN hours;
ALTER TABLE tracking_sessions DROP COLUMN contract_id;

-- Drop contracts table
DROP TABLE IF EXISTS contracts;
