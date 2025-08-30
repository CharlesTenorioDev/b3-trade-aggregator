-- migrations/001_create_trades_table.up.sql

-- Cria a tabela 'trades' para armazenar as negociações
CREATE TABLE IF NOT EXISTS trades (
    id BIGSERIAL PRIMARY KEY,                          -- ID único da negociação (chave primária)
    trade_date DATE NOT NULL,                          -- Data em que a negociação ocorreu
    instrument_code VARCHAR(10) NOT NULL,              -- Código do instrumento (ticker), ex: PETR4
    negotiated_price NUMERIC(18, 4) NOT NULL,          -- Valor unitário do ativo com precisão financeira
    negotiated_quantity INTEGER NOT NULL,              -- Quantidade de ativos negociados
    closing_time VARCHAR(9) NOT NULL,                  -- Horário da negociação no formato HHMMSSmmm
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()  -- Timestamp da criação do registro
);

-- Cria um índice composto para otimizar as consultas que filtram por ticker e data.
-- Este índice é crucial para a performance das suas consultas agregadas.
CREATE INDEX IF NOT EXISTS idx_trades_instrument_date ON trades (instrument_code, trade_date);