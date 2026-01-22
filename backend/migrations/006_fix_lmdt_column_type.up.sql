-- Fix lmdt column type from DATE to INTEGER to match M3 YYYYMMDD format
-- This aligns with other M3 date fields like stdt, fidt, pldt which are stored as INTEGER
-- lmdt is a metadata field for change tracking and should match the M3 source format

-- Convert existing DATE values to INTEGER format (YYYYMMDD)
-- Then change column type to INTEGER

-- Fix customer_order_lines
ALTER TABLE customer_order_lines
    ALTER COLUMN lmdt TYPE INTEGER
    USING CASE
        WHEN lmdt IS NULL THEN NULL
        ELSE TO_CHAR(lmdt, 'YYYYMMDD')::INTEGER
    END;

-- Fix manufacturing_orders
ALTER TABLE manufacturing_orders
    ALTER COLUMN lmdt TYPE INTEGER
    USING CASE
        WHEN lmdt IS NULL THEN NULL
        ELSE TO_CHAR(lmdt, 'YYYYMMDD')::INTEGER
    END;

-- Fix planned_manufacturing_orders
ALTER TABLE planned_manufacturing_orders
    ALTER COLUMN lmdt TYPE INTEGER
    USING CASE
        WHEN lmdt IS NULL THEN NULL
        ELSE TO_CHAR(lmdt, 'YYYYMMDD')::INTEGER
    END;

-- Fix production_orders (unified view - keep as INTEGER for consistency with source data)
ALTER TABLE production_orders
    ALTER COLUMN lmdt TYPE INTEGER
    USING CASE
        WHEN lmdt IS NULL THEN NULL
        ELSE TO_CHAR(lmdt, 'YYYYMMDD')::INTEGER
    END;

-- Fix customer_orders
ALTER TABLE customer_orders
    ALTER COLUMN lmdt TYPE INTEGER
    USING CASE
        WHEN lmdt IS NULL THEN NULL
        ELSE TO_CHAR(lmdt, 'YYYYMMDD')::INTEGER
    END;

-- Fix deliveries
ALTER TABLE deliveries
    ALTER COLUMN lmdt TYPE INTEGER
    USING CASE
        WHEN lmdt IS NULL THEN NULL
        ELSE TO_CHAR(lmdt, 'YYYYMMDD')::INTEGER
    END;
