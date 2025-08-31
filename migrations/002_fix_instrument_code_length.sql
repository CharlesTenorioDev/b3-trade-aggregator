-- migrations/002_fix_instrument_code_length.sql

-- Fix the instrument_code field length to accommodate longer B3 instrument codes
-- B3 instrument codes can be up to 20 characters (e.g., PETR4, VALE3, etc.)

-- Drop the existing index first (since it references the column we're changing)
DROP INDEX IF EXISTS idx_trades_instrument_date;

-- Alter the instrument_code column to allow longer values
ALTER TABLE trades ALTER COLUMN instrument_code TYPE VARCHAR(20);

-- Recreate the index
CREATE INDEX IF NOT EXISTS idx_trades_instrument_date ON trades (instrument_code, trade_date);
