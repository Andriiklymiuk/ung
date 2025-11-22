-- Drop unique index on contract_num
DROP INDEX IF EXISTS idx_contracts_contract_num;

-- Remove contract_num column from contracts table
ALTER TABLE contracts DROP COLUMN contract_num;
