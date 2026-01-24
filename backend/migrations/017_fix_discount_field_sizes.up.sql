-- Migration 017: Fix discount field sizes to accommodate decimal strings
-- DIP1-DIP6 return values like "0.000000000000000" (17 chars) from Data Fabric

ALTER TABLE customer_order_lines
    ALTER COLUMN dip1 TYPE VARCHAR(30),
    ALTER COLUMN dip2 TYPE VARCHAR(30),
    ALTER COLUMN dip3 TYPE VARCHAR(30),
    ALTER COLUMN dip4 TYPE VARCHAR(30),
    ALTER COLUMN dip5 TYPE VARCHAR(30),
    ALTER COLUMN dip6 TYPE VARCHAR(30);
