-- Fix customer_order_lines date columns from DATE to INTEGER
-- M3 stores dates as INTEGER in YYYYMMDD format (e.g., 20260303)
-- This aligns with MO/MOP date fields (stdt, fidt, pldt) which are already INTEGER
-- Customer order line dates: dwdt, codt, pldt, fded, lded should match this pattern

-- Convert existing DATE values to INTEGER format (YYYYMMDD)
-- Then change column type to INTEGER

-- DWDT - Requested delivery date
ALTER TABLE customer_order_lines
    ALTER COLUMN dwdt TYPE INTEGER
    USING CASE
        WHEN dwdt IS NULL THEN NULL
        ELSE TO_CHAR(dwdt, 'YYYYMMDD')::INTEGER
    END;

-- CODT - Confirmed delivery date
ALTER TABLE customer_order_lines
    ALTER COLUMN codt TYPE INTEGER
    USING CASE
        WHEN codt IS NULL THEN NULL
        ELSE TO_CHAR(codt, 'YYYYMMDD')::INTEGER
    END;

-- PLDT - Planning date
ALTER TABLE customer_order_lines
    ALTER COLUMN pldt TYPE INTEGER
    USING CASE
        WHEN pldt IS NULL THEN NULL
        ELSE TO_CHAR(pldt, 'YYYYMMDD')::INTEGER
    END;

-- FDED - First delivery date
ALTER TABLE customer_order_lines
    ALTER COLUMN fded TYPE INTEGER
    USING CASE
        WHEN fded IS NULL THEN NULL
        ELSE TO_CHAR(fded, 'YYYYMMDD')::INTEGER
    END;

-- LDED - Last delivery date
ALTER TABLE customer_order_lines
    ALTER COLUMN lded TYPE INTEGER
    USING CASE
        WHEN lded IS NULL THEN NULL
        ELSE TO_CHAR(lded, 'YYYYMMDD')::INTEGER
    END;

-- Create index on key date fields for performance
CREATE INDEX IF NOT EXISTS idx_co_lines_dwdt_codt ON customer_order_lines(dwdt, codt) WHERE dwdt IS NOT NULL;

-- Comment explaining the format
COMMENT ON COLUMN customer_order_lines.dwdt IS 'Requested delivery date in M3 format (YYYYMMDD integer, e.g., 20260303)';
COMMENT ON COLUMN customer_order_lines.codt IS 'Confirmed delivery date in M3 format (YYYYMMDD integer)';
COMMENT ON COLUMN customer_order_lines.pldt IS 'Planning date in M3 format (YYYYMMDD integer)';
