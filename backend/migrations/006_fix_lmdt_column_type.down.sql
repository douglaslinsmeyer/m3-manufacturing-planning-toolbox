-- Rollback lmdt column type from INTEGER back to DATE

-- Revert customer_order_lines
ALTER TABLE customer_order_lines
    ALTER COLUMN lmdt TYPE DATE
    USING CASE
        WHEN lmdt IS NULL THEN NULL
        ELSE TO_DATE(lmdt::text, 'YYYYMMDD')
    END;

-- Revert manufacturing_orders
ALTER TABLE manufacturing_orders
    ALTER COLUMN lmdt TYPE DATE
    USING CASE
        WHEN lmdt IS NULL THEN NULL
        ELSE TO_DATE(lmdt::text, 'YYYYMMDD')
    END;

-- Revert planned_manufacturing_orders
ALTER TABLE planned_manufacturing_orders
    ALTER COLUMN lmdt TYPE DATE
    USING CASE
        WHEN lmdt IS NULL THEN NULL
        ELSE TO_DATE(lmdt::text, 'YYYYMMDD')
    END;

-- Revert production_orders
ALTER TABLE production_orders
    ALTER COLUMN lmdt TYPE DATE
    USING CASE
        WHEN lmdt IS NULL THEN NULL
        ELSE TO_DATE(lmdt::text, 'YYYYMMDD')
    END;

-- Revert customer_orders
ALTER TABLE customer_orders
    ALTER COLUMN lmdt TYPE DATE
    USING CASE
        WHEN lmdt IS NULL THEN NULL
        ELSE TO_DATE(lmdt::text, 'YYYYMMDD')
    END;

-- Revert deliveries
ALTER TABLE deliveries
    ALTER COLUMN lmdt TYPE DATE
    USING CASE
        WHEN lmdt IS NULL THEN NULL
        ELSE TO_DATE(lmdt::text, 'YYYYMMDD')
    END;
