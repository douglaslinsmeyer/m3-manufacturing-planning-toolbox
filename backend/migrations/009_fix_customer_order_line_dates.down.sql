-- Rollback: Convert customer_order_lines date columns from INTEGER back to DATE
-- This is the reverse of the up migration

-- Convert INTEGER (YYYYMMDD) back to DATE format

-- DWDT - Requested delivery date
ALTER TABLE customer_order_lines
    ALTER COLUMN dwdt TYPE DATE
    USING CASE
        WHEN dwdt IS NULL OR dwdt = 0 THEN NULL
        ELSE TO_DATE(dwdt::TEXT, 'YYYYMMDD')
    END;

-- CODT - Confirmed delivery date
ALTER TABLE customer_order_lines
    ALTER COLUMN codt TYPE DATE
    USING CASE
        WHEN codt IS NULL OR codt = 0 THEN NULL
        ELSE TO_DATE(codt::TEXT, 'YYYYMMDD')
    END;

-- PLDT - Planning date
ALTER TABLE customer_order_lines
    ALTER COLUMN pldt TYPE DATE
    USING CASE
        WHEN pldt IS NULL OR pldt = 0 THEN NULL
        ELSE TO_DATE(pldt::TEXT, 'YYYYMMDD')
    END;

-- FDED - First delivery date
ALTER TABLE customer_order_lines
    ALTER COLUMN fded TYPE DATE
    USING CASE
        WHEN fded IS NULL OR fded = 0 THEN NULL
        ELSE TO_DATE(fded::TEXT, 'YYYYMMDD')
    END;

-- LDED - Last delivery date
ALTER TABLE customer_order_lines
    ALTER COLUMN lded TYPE DATE
    USING CASE
        WHEN lded IS NULL OR lded = 0 THEN NULL
        ELSE TO_DATE(lded::TEXT, 'YYYYMMDD')
    END;

-- Drop the index created in up migration
DROP INDEX IF EXISTS idx_co_lines_dwdt_codt;
