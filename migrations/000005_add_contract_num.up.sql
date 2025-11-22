-- Add contract_num column to contracts table
ALTER TABLE contracts ADD COLUMN contract_num TEXT NOT NULL DEFAULT '';

-- Create unique index on contract_num
CREATE UNIQUE INDEX idx_contracts_contract_num ON contracts(contract_num);
