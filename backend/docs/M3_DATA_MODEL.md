# M3 Data Model Integration

This document describes how M3 Data Fabric data maps to our PostgreSQL schema.

## Overview

Our schema uses a **hybrid approach**:
- **Core fields**: Normalized columns for frequently-queried data
- **Attributes**: JSONB for flexible M3 attributes (ATV1-ATV0, UCA1-UCA0, etc.)
- **Reference linking**: RORC/RORN/RORL/RORX fields to link CO lines → MOs → MOPs

## Critical Fields for Linking

### Reference Order Fields (RORC, RORN, RORL, RORX)

These fields create the relationships between entities:

```sql
-- Link CO lines to Manufacturing Orders
SELECT co.*, mo.*
FROM customer_order_lines co
JOIN manufacturing_orders mo
  ON mo.rorc = 3              -- 3 = Customer Order
  AND mo.rorn = co.order_number
  AND mo.rorl = co.line_number
  AND mo.rorx = co.line_suffix;

-- Link CO lines to Planned Orders
SELECT co.*, mop.*
FROM customer_order_lines co
JOIN planned_manufacturing_orders mop
  ON mop.rorc = 3
  AND mop.rorn = co.order_number
  AND mop.rorl = co.line_number;
```

**RORC Values** (Reference Order Category):
- `3` = Customer Order (CO)
- `4` = Work Order (MO)
- `10` = Forecast
- `20` = Stock Order

## M3 Metadata Fields

### Change Tracking
- `rgdt` (YYYYMMDD) - Registration date
- `rgtm` (HHMMSS) - Registration time
- `lmdt` (YYYYMMDD) - Last modified date - **KEY for incremental loads**
- `lmts` (timestamp) - Last modified timestamp
- `chno` - Change number
- `chid` - Change ID

### Data Lake Fields
- `m3_timestamp` - Timestamp from Data Fabric
- `is_deleted` - Boolean version of M3's string 'deleted' field
- `sync_timestamp` - When we loaded this record

### Incremental Load Strategy

```sql
-- Get records modified since last sync
WHERE deleted = 'false'  -- STRING comparison!
  AND LMDT >= 20240101   -- Last sync date in YYYYMMDD format
ORDER BY LMDT, LMTS
```

## JSONB Attributes Structure

### Customer Order Line Attributes

```json
{
  "builtin_numeric": {
    "ATV1": 123.45,
    "ATV2": 67.89,
    "ATV3": null,
    "ATV4": null,
    "ATV5": null
  },
  "builtin_string": {
    "ATV6": "Color:Red",
    "ATV7": "Size:Large",
    "ATV8": null,
    "ATV9": null,
    "ATV0": null
  },
  "user_defined_alpha": {
    "UCA1": "CustomValue1",
    "UCA2": "CustomValue2",
    "UCA3": null,
    "UCA4": null,
    "UCA5": null,
    "UCA6": null,
    "UCA7": null,
    "UCA8": null,
    "UCA9": null,
    "UCA0": null
  },
  "user_defined_numeric": {
    "UDN1": 1000.50,
    "UDN2": 2500.75,
    "UDN3": null,
    "UDN4": null,
    "UDN5": null,
    "UDN6": null
  },
  "user_defined_dates": {
    "UID1": "2024-01-15",
    "UID2": "2024-02-20",
    "UID3": null
  },
  "user_text": {
    "UCT1": "Long text field content"
  },
  "discounts": {
    "DIP1": 5.0,
    "DIP2": 2.5,
    "DIP3": null,
    "DIA1": 100.00,
    "DIA2": 50.00,
    "DIA3": null
  }
}
```

### Querying JSONB Attributes

```sql
-- Find CO lines with specific color attribute
SELECT *
FROM customer_order_lines
WHERE attributes->'builtin_string'->>'ATV6' = 'Color:Red';

-- Find MOs with numeric attribute > threshold
SELECT *
FROM manufacturing_orders
WHERE (attributes->'user_defined_numeric'->>'UDN1')::numeric > 1000;

-- Check if attribute exists
SELECT *
FROM customer_order_lines
WHERE attributes->'builtin_string' ? 'ATV6';

-- Get all attributes containing a keyword
SELECT order_number, attributes
FROM customer_order_lines
WHERE attributes::text ILIKE '%urgent%';
```

### Populating JSONB from M3 Data

```go
// Example Go code to build attributes JSONB
attributes := map[string]interface{}{
    "builtin_numeric": map[string]interface{}{
        "ATV1": row.ATV1,
        "ATV2": row.ATV2,
        "ATV3": row.ATV3,
        "ATV4": row.ATV4,
        "ATV5": row.ATV5,
    },
    "builtin_string": map[string]interface{}{
        "ATV6": row.ATV6,
        "ATV7": row.ATV7,
        "ATV8": row.ATV8,
        "ATV9": row.ATV9,
        "ATV0": row.ATV0,
    },
    "user_defined_alpha": map[string]interface{}{
        "UCA1": row.UCA1,
        "UCA2": row.UCA2,
        // ... etc
    },
}

attributesJSON, _ := json.Marshal(attributes)
```

## Compass Data Fabric Queries

### Customer Order Lines (OOLINE)

```sql
SELECT
  -- Identifiers
  CONO, DIVI, ORNO, PONR, POSX,
  ITNO, ITDS, ORTY, ORST, FACI, WHLO,

  -- Quantities
  ORQT, RNQT, ALQT, DLQT, IVQT,

  -- Dates
  DWDT, CODT, PLDT, FDED, LDED,

  -- Pricing
  SAPR, NEPR, LNAM, CUCD,

  -- Reference orders (CRITICAL for linking!)
  RORC, RORN, RORL, RORX,

  -- Built-in attributes
  ATV1, ATV2, ATV3, ATV4, ATV5,
  ATV6, ATV7, ATV8, ATV9, ATV0,

  -- User-defined attributes
  UCA1, UCA2, UCA3, UCA4, UCA5, UCA6, UCA7, UCA8, UCA9, UCA0,
  UDN1, UDN2, UDN3, UDN4, UDN5, UDN6,
  UID1, UID2, UID3,
  UCT1,

  -- Discounts
  DIP1, DIP2, DIP3, DIA1, DIA2, DIA3,

  -- Attribute model
  ATNR, ATMO, ATPR,

  -- Metadata
  RGDT, RGTM, LMDT, CHNO, CHID, LMTS,
  timestamp, deleted
FROM OOLINE
WHERE deleted = 'false'
  AND LMDT >= 20240101
ORDER BY LMDT, LMTS;
```

### Manufacturing Orders (MWOHED)

```sql
SELECT
  -- Identifiers
  CONO, DIVI, FACI, MFNO, PRNO, ITNO,

  -- Status
  WHST, WHHS, WMST, MOHS,

  -- Quantities
  ORQT, ORQA, RVQT, RVQA, MAQT, MAQA,

  -- Dates
  STDT, FIDT, RSDT, REFD, RPDT, FSTD, FFID,

  -- Planning
  PRIO, RESP, PLGR, WCLN, PRDY,

  -- Reference orders (CRITICAL!)
  RORC, RORN, RORL, RORX,

  -- Hierarchy
  PRHL, MFHL, PRLO, MFLO,

  -- Attributes
  ATNR, CFIN,

  -- Project
  PROJ, ELNO,

  -- Metadata
  ORTY, GETP, BDCD, SCEX,
  RGDT, RGTM, LMDT, CHNO, CHID, LMTS,
  timestamp, deleted
FROM MWOHED
WHERE deleted = 'false'
  AND LMDT >= 20240101
ORDER BY LMDT, LMTS;
```

### Planned Manufacturing Orders (MMOPLP)

```sql
SELECT
  -- Identifiers
  CONO, DIVI, FACI, PLPN, PLPS, PRNO, ITNO,

  -- Status
  PSTS, WHST, ACTP,

  -- Quantities
  PPQT, ORQA,

  -- Dates
  RELD, STDT, FIDT, MSTI, MFTI, PLDT,

  -- Planning
  RESP, PRIP, PLGR, WCLN, PRDY,

  -- Reference orders (CRITICAL!)
  RORC, RORN, RORL, RORX, RORH,

  -- Hierarchy
  PLLO, PLHL,

  -- Attributes
  ATNR, CFIN,

  -- Project
  PROJ, ELNO,

  -- Metadata
  GETY, NUAU, ORDP, ORTY,
  MSG1, MSG2, MSG3, MSG4,  -- Warning messages
  RGDT, RGTM, LMDT, CHNO, CHID, LMTS,
  timestamp, deleted
FROM MMOPLP
WHERE deleted = 'false'
  AND LMDT >= 20240101
ORDER BY LMDT, LMTS;
```

## Status Field Values

### CO Line Status (ORST)
- `20` - Released
- `33` - Partially delivered
- `35` - Fully delivered
- `40` - Partially invoiced
- `77` - Closed

### MO Status (WHST)
- `20` - Released
- `30` - Started
- `50` - Reported
- `90` - Closed

### MOP Status (PSTS)
- `10` - Planning
- `20` - Approved
- `30` - Confirmed
- `77` - Closed

## Date Field Formats

All date fields in M3 Data Fabric are integers in YYYYMMDD format:
- `20240115` = January 15, 2024
- Convert to PostgreSQL DATE: `TO_DATE(LMDT::text, 'YYYYMMDD')`

## Best Practices

### 1. Always Filter on `deleted`
```sql
WHERE deleted = 'false'  -- STRING, not boolean!
```

### 2. Use LMDT for Incremental Loads
```sql
-- Store last sync LMDT in snapshot_metadata table
-- Next sync: WHERE LMDT >= last_sync_lmdt
```

### 3. Index Reference Order Fields
```sql
CREATE INDEX idx_entity_rorc_rorn ON table_name(rorc, rorn, rorl, rorx);
```

### 4. Query JSONB Efficiently
```sql
-- Use GIN indexes
CREATE INDEX idx_attributes ON table_name USING GIN(attributes);

-- Use -> for object navigation, ->> for text extraction
WHERE attributes->'builtin_string'->>'ATV6' = 'Value'
```

### 5. Partition Large Tables
```sql
-- Partition by LMDT or CONO for better performance
CREATE TABLE co_lines_2024 PARTITION OF customer_order_lines
  FOR VALUES FROM (20240101) TO (20250101);
```

## Field Mapping Reference

| M3 Table | M3 Field | PostgreSQL Table | PostgreSQL Column | Type |
|----------|----------|------------------|-------------------|------|
| OOLINE | ORNO | customer_order_lines | order_number | VARCHAR(50) |
| OOLINE | PONR | customer_order_lines | line_number | VARCHAR(10) |
| OOLINE | RORC | customer_order_lines | rorc | INTEGER |
| OOLINE | RORN | customer_order_lines | rorn | VARCHAR(50) |
| OOLINE | ATV1-ATV5 | customer_order_lines | attributes->'builtin_numeric' | JSONB |
| MWOHED | MFNO | manufacturing_orders | mo_number | VARCHAR(50) |
| MWOHED | RORC | manufacturing_orders | rorc | INTEGER |
| MMOPLP | PLPN | planned_manufacturing_orders | mop_number | VARCHAR(50) |
| MMOPLP | MSG1-4 | planned_manufacturing_orders | messages | JSONB |

## Example: Complete Data Flow

1. **Fetch from Compass Data Fabric**
   ```sql
   SELECT CONO, ORNO, PONR, RORC, RORN, ATV1, ATV6, ...
   FROM OOLINE WHERE deleted = 'false' AND LMDT >= 20240101
   ```

2. **Transform and Insert**
   ```sql
   INSERT INTO customer_order_lines (
     cono, order_number, line_number,
     rorc, rorn, rorl, rorx,
     attributes, lmdt, is_deleted
   ) VALUES (
     100, 'CO12345', '1',
     3, 'CO12345', 1, 0,
     '{"builtin_numeric": {"ATV1": 123.45}}'::jsonb,
     '2024-01-15', false
   );
   ```

3. **Query with Joins**
   ```sql
   SELECT
     co.order_number,
     co.line_number,
     mo.mo_number,
     mo.status,
     co.attributes->'builtin_string'->>'ATV6' as color
   FROM customer_order_lines co
   LEFT JOIN manufacturing_orders mo
     ON mo.rorc = 3
     AND mo.rorn = co.order_number
     AND mo.rorl = CAST(co.line_number AS INTEGER)
   WHERE co.is_deleted = false
     AND co.lmdt >= '2024-01-01';
   ```
