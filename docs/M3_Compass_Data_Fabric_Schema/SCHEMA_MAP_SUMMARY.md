# M3 Schema Maps - Summary Guide

## Overview

This directory contains comprehensive schema maps for eight critical M3 supply chain tables:

### Order Management
1. **CO_HEADER_SCHEMA_MAP.md** - Customer Order Header (OOHEAD) - 163 fields
2. **CO_LINE_SCHEMA_MAP.md** - Customer Order Lines (OOLINE) - 303 fields
3. **ATTRIBUTE_SCHEMA_MAP.md** - Order Attributes (MOATTR) - 52 fields

### Manufacturing
4. **MO_SCHEMA_MAP.md** - Manufacturing Orders (MWOHED) - 149 fields
5. **MOP_SCHEMA_MAP.md** - Planned Manufacturing Orders (MMOPLP) - 91 fields

### Supply Chain Linkage ⭐
6. **PREALLOCATION_SCHEMA_MAP.md** - Pre-Allocation (MPREAL) - 44 fields
   - **Critical linking table** between demand (CO) and supply (MO/MOP/PO)

### Delivery & Fulfillment
7. **DELIVERY_HEADER_SCHEMA_MAP.md** - Delivery Header (ODHEAD) - 87 fields
8. **DELIVERY_LINE_SCHEMA_MAP.md** - Delivery Lines (ODLINE) - 102 fields

---

## Purpose

These schema maps provide:
- **Human-readable field names** for cryptic M3 codes
- **Business descriptions** for each field
- **Data types and constraints** from M3 Data Catalog
- **Relationships and foreign keys** for data modeling
- **Common query patterns** for each table
- **Categorized field lists** for easier navigation

---

## How to Use These Maps

### 1. For Database Schema Design
- Use the "Key Fields for Data Modeling" section in each map
- Reference the field categories to decide core vs. JSONB storage
- Follow the primary/foreign key definitions for relationships

### 2. For SQL Query Writing
- Find M3 field codes by searching human-readable names
- Use the "Common Query Patterns" sections as templates
- Reference the relationship sections for JOIN conditions

### 3. For ETL Development
- Use "Critical for Incremental Load" fields (LMDT, LMTS)
- Reference the audit fields section for change tracking
- Note the Data Lake metadata fields (deleted, timestamp)

### 4. For Business Analysis
- Field categories help identify business areas
- Descriptions explain business meaning
- Status value tables clarify codes

---

## Optimal Data Harvesting Strategy

### Overview

For building an external system to download and store M3 data, follow this comprehensive harvesting approach:

### Phase 1: Initial Full Load

Load data in this order to satisfy foreign key dependencies:

```sql
-- 1. Order Headers (no dependencies)
SELECT * FROM OOHEAD WHERE deleted = 'false' AND ORDT >= 20240101;

-- 2. Order Lines (depends on OOHEAD)
SELECT * FROM OOLINE WHERE deleted = 'false' AND LMDT >= 20240101;

-- 3. Order Attributes (depends on OOLINE)
SELECT * FROM MOATTR WHERE deleted = 'false' AND ORCA = '3' AND LMDT >= 20240101;

-- 4. Manufacturing Orders (depends on OOLINE for demand reference)
SELECT * FROM MWOHED WHERE deleted = 'false' AND LMDT >= 20240101;

-- 5. Planned Manufacturing Orders
SELECT * FROM MMOPLP WHERE deleted = 'false' AND LMDT >= 20240101;

-- 6. Pre-Allocations (depends on all above) ⭐ CRITICAL
SELECT * FROM MPREAL WHERE deleted = 'false' AND LMDT >= 20240101;

-- 7. Delivery Headers (depends on OOHEAD)
SELECT * FROM ODHEAD WHERE deleted = 'false' AND LMDT >= 20240101;

-- 8. Delivery Lines (depends on ODHEAD and OOLINE)
SELECT * FROM ODLINE WHERE deleted = 'false' AND LMDT >= 20240101;
```

### Phase 2: Incremental Updates

After initial load, use LMDT for efficient incremental updates:

```sql
-- Run daily/hourly to capture changes
SELECT * FROM [TABLE]
WHERE deleted = 'false'
  AND LMDT >= [LAST_SYNC_DATE]
ORDER BY LMDT, LMTS
LIMIT 10000;  -- Batch in chunks
```

### Optimal Batch Sizes

Based on field counts and typical volumes:

| Table | Recommended Batch Size | Reason |
|-------|----------------------|--------|
| OOHEAD | 5,000 records | Moderate field count (163) |
| OOLINE | 2,000 records | Very wide (303 fields) |
| MOATTR | 10,000 records | Narrow table (52 fields) |
| MWOHED | 3,000 records | Wide table (149 fields) |
| MMOPLP | 5,000 records | Medium width (91 fields) |
| MPREAL | 10,000 records | Narrow linking table (44 fields) |
| ODHEAD | 5,000 records | Medium width (87 fields) |
| ODLINE | 3,000 records | Wide table (102 fields) |

### Critical Filters for Initial Load

```sql
-- CO Headers: Last 2 years of orders
WHERE deleted = 'false'
  AND ORDT >= 20240101
  AND ORST < '90'  -- Exclude closed

-- CO Lines: Active lines only
WHERE deleted = 'false'
  AND LMDT >= 20240101
  AND ORST < '77'  -- Not confirmed/closed

-- MOs: Active and recent
WHERE deleted = 'false'
  AND LMDT >= 20240101
  AND WHST < '90'  -- Not closed

-- MOPs: Active proposals only
WHERE deleted = 'false'
  AND PSTS IN ('10', '20')  -- New, Acknowledged

-- MPREAL: Active allocations
WHERE deleted = 'false'
  AND STSB = '10'  -- Active

-- Deliveries: Recent deliveries
WHERE deleted = 'false'
  AND DLDT >= 20240101
```

### Parallel Harvesting Strategy

Harvest independent tables in parallel for performance:

**Group 1 (Parallel)**: No dependencies
- OOHEAD

**Group 2 (Parallel)**: Depend on Group 1
- OOLINE
- ODHEAD

**Group 3 (Parallel)**: Depend on Group 2
- MOATTR
- MWOHED
- MMOPLP
- ODLINE

**Group 4 (Sequential)**: Depends on Group 3
- MPREAL (requires MOs, MOPs, and CO lines to exist)

### Sample Harvesting Queries

#### Complete Order with All Dependencies
```sql
-- Single query to get order with all related data
WITH order_data AS (
  SELECT * FROM OOHEAD WHERE ORNO = 'CO123456' AND deleted = 'false'
),
line_data AS (
  SELECT * FROM OOLINE WHERE ORNO = 'CO123456' AND deleted = 'false'
),
attr_data AS (
  SELECT * FROM MOATTR
  WHERE ORCA = '3' AND RIDN = 'CO123456' AND deleted = 'false'
),
alloc_data AS (
  SELECT * FROM MPREAL
  WHERE DOCA = '3' AND DRDN = 'CO123456' AND deleted = 'false'
),
mo_data AS (
  SELECT mo.* FROM MWOHED mo
  JOIN alloc_data pa ON mo.MFNO = pa.ARDN AND pa.AOCA = '2'
  WHERE mo.deleted = 'false'
),
delivery_data AS (
  SELECT * FROM ODHEAD WHERE ORNO = 'CO123456' AND deleted = 'false'
),
delivery_line_data AS (
  SELECT * FROM ODLINE WHERE ORNO = 'CO123456' AND deleted = 'false'
)
SELECT
  'OOHEAD' as table_name, COUNT(*) as record_count FROM order_data
UNION ALL
SELECT 'OOLINE', COUNT(*) FROM line_data
UNION ALL
SELECT 'MOATTR', COUNT(*) FROM attr_data
UNION ALL
SELECT 'MPREAL', COUNT(*) FROM alloc_data
UNION ALL
SELECT 'MWOHED', COUNT(*) FROM mo_data
UNION ALL
SELECT 'ODHEAD', COUNT(*) FROM delivery_data
UNION ALL
SELECT 'ODLINE', COUNT(*) FROM delivery_line_data;
```

#### Incremental Harvest with Pagination
```sql
-- Efficient incremental load with cursor-based pagination
SELECT *
FROM OOLINE
WHERE deleted = 'false'
  AND (LMDT > @last_sync_date
       OR (LMDT = @last_sync_date AND LMTS > @last_sync_timestamp))
ORDER BY LMDT, LMTS
LIMIT 2000;
```

### Data Quality Checks During Harvest

Run these validation queries to ensure data integrity:

```sql
-- Check 1: Orphaned order lines (no header)
SELECT ol.*
FROM OOLINE ol
LEFT JOIN OOHEAD oh ON oh.ORNO = ol.ORNO AND oh.deleted = 'false'
WHERE ol.deleted = 'false'
  AND oh.ORNO IS NULL;

-- Check 2: Allocations with missing supply orders
SELECT pa.*
FROM MPREAL pa
LEFT JOIN MWOHED mo ON mo.MFNO = pa.ARDN AND pa.AOCA = '2' AND mo.deleted = 'false'
WHERE pa.deleted = 'false'
  AND pa.AOCA = '2'
  AND mo.MFNO IS NULL;

-- Check 3: Attributes with missing order lines
SELECT attr.*
FROM MOATTR attr
LEFT JOIN OOLINE ol
  ON ol.ORNO = attr.RIDN
  AND ol.PONR = attr.RIDL
  AND ol.POSX = attr.RIDX
  AND ol.deleted = 'false'
WHERE attr.deleted = 'false'
  AND attr.ORCA = '3'
  AND ol.ORNO IS NULL;

-- Check 4: Delivery lines without headers
SELECT dl.*
FROM ODLINE dl
LEFT JOIN ODHEAD dh ON dh.DLIX = dl.DLIX AND dh.deleted = 'false'
WHERE dl.deleted = 'false'
  AND dh.DLIX IS NULL;
```

### Recommended Sync Frequency

| Table | Sync Frequency | Reason |
|-------|---------------|--------|
| OOHEAD | Every 15 min | Orders change frequently |
| OOLINE | Every 15 min | Line updates are common |
| MOATTR | Hourly | Attributes rarely change after creation |
| MWOHED | Every 15 min | Status updates throughout day |
| MMOPLP | Every 30 min | MRP runs periodically |
| MPREAL | Every 15 min | Allocations change with planning |
| ODHEAD | Every 10 min | Real-time delivery tracking |
| ODLINE | Every 10 min | Real-time delivery tracking |

### Storage Size Estimates

Approximate storage per record (compressed):

| Table | Bytes/Record | Rationale |
|-------|-------------|-----------|
| OOHEAD | 8 KB | Many text fields, user-defined |
| OOLINE | 12 KB | Widest table with attributes |
| MOATTR | 1 KB | Compact linking records |
| MWOHED | 6 KB | Manufacturing details |
| MMOPLP | 4 KB | Planning data |
| MPREAL | 500 bytes | Simple linking table |
| ODHEAD | 3 KB | Delivery details |
| ODLINE | 4 KB | Delivery line details |

**Example**: 100,000 order lines = ~1.2 GB (OOLINE alone)

---

## Critical Relationships

### Complete Supply Chain Flow

```
Order Entry
    ↓
OOHEAD (Order Header)
    ├─ CUNO, ORDT, RLDT, CUCD
    │
    ├─→ OOLINE (Order Lines)
    │       ├─ ITNO, ORQT, SAPR
    │       │
    │       ├─→ MOATTR (Attributes)
    │       │       ├─ ATNR, ATID
    │       │       └─ ATAV/ATAN (values)
    │       │
    │       └─→ MPREAL (Pre-Allocation) ⭐ CRITICAL LINKING TABLE
    │               ├─ DOCA='3', DRDN=ORNO, DRDL=PONR (Demand Side)
    │               ├─ AOCA='2'/'5', ARDN=MFNO/PLPN (Supply Side)
    │               └─ PQTY (allocated quantity)
    │
    ├─→ Supply Planning (linked via MPREAL)
    │       │
    │       ├─→ MMOPLP (Planned MOs)
    │       │       ├─ AOCA='5' in MPREAL
    │       │       ├─ RORC=3 → legacy link to OOLINE
    │       │       └─ Converts to MWOHED
    │       │
    │       └─→ MWOHED (Manufacturing Orders)
    │               ├─ AOCA='2' in MPREAL
    │               ├─ RORC=3 → legacy link to OOLINE
    │               ├─ RORC=2 → links to parent MO
    │               └─ Multi-level hierarchy
    │
    └─→ Fulfillment
            │
            └─→ ODHEAD (Delivery Header)
                    ├─ DLIX, DLDT, CONN
                    │
                    └─→ ODLINE (Delivery Lines)
                            ├─ Links: DLIX + ORNO + PONR + POSX
                            ├─ DLQT, IVQT
                            └─ IVNO (invoice reference)
```

**Key Note**: MPREAL is the **primary** linking mechanism between customer orders and their supply sources. The RORC/RORN fields in MWOHED/MMOPLP provide legacy linking but MPREAL offers more flexibility and detail.

### Table Relationships Summary

| From Table | To Table | Link Fields | Relationship Type |
|------------|----------|-------------|-------------------|
| OOHEAD | OOLINE | ORNO | One-to-Many (header to lines) |
| OOLINE | MOATTR | ORCA=3, RIDN=ORNO, RIDL=PONR | One-to-Many (line to attributes) |
| **OOLINE** | **MPREAL** | **DOCA=3, DRDN=ORNO, DRDL=PONR** | **One-to-Many (demand to allocations)** ⭐ |
| **MPREAL** | **MWOHED** | **AOCA=2, ARDN=MFNO** | **Many-to-Many (allocations to MOs)** ⭐ |
| **MPREAL** | **MMOPLP** | **AOCA=5, ARDN=PLPN** | **Many-to-Many (allocations to MOPs)** ⭐ |
| OOLINE | MWOHED | RORC=3, RORN=ORNO, RORL=PONR | One-to-Many (legacy demand link) |
| OOLINE | MMOPLP | RORC=3, RORN=ORNO, RORL=PONR | One-to-Many (legacy demand link) |
| MWOHED | MWOHED | RORC=2, RORN=parent MFNO | One-to-Many (parent to child MOs) |
| MMOPLP | MWOHED | Converts via release | One-to-One (proposal to MO) |
| OOHEAD | ODHEAD | ORNO | One-to-Many (order to deliveries) |
| ODHEAD | ODLINE | DLIX | One-to-Many (delivery to lines) |
| OOLINE | ODLINE | ORNO + PONR + POSX | One-to-Many (line to delivery lines) |

**⭐ = Critical for understanding demand-supply linkage**

### Join Relationships

#### CO Line → MO (via MPREAL - RECOMMENDED)
```sql
-- Recommended approach using MPREAL
SELECT
  co.*,
  pa.PQTY as allocated_qty,
  mo.*
FROM OOLINE co
JOIN MPREAL pa
  ON pa.DOCA = '3'                  -- Customer Order
  AND pa.DRDN = co.ORNO             -- Order Number
  AND pa.DRDL = co.PONR             -- Line Number
  AND pa.DRDX = co.POSX             -- Line Suffix
  AND pa.AOCA = '2'                 -- Manufacturing Order
  AND pa.deleted = 'false'
JOIN MWOHED mo
  ON mo.MFNO = pa.ARDN              -- MO Number
  AND mo.deleted = 'false'
WHERE co.deleted = 'false'
```

#### CO Line → MO (Legacy RORC Method)
```sql
-- Legacy approach using RORC fields in MO
SELECT co.*, mo.*
FROM OOLINE co
JOIN MWOHED mo
  ON mo.RORC = 3                    -- Customer Order
  AND mo.RORN = co.ORNO             -- Order Number
  AND mo.RORL = co.PONR             -- Line Number
  AND mo.RORX = co.POSX             -- Line Suffix
WHERE co.deleted = 'false'
  AND mo.deleted = 'false'
```

**Note**: MPREAL method is preferred as it:
- Shows allocated quantities (PQTY)
- Handles many-to-many relationships
- Works with MOPs, POs, and DOs
- Provides allocation status

#### CO Line → MOP (via MPREAL - RECOMMENDED)
```sql
-- Recommended approach using MPREAL
SELECT
  co.*,
  pa.PQTY as allocated_qty,
  mop.*
FROM OOLINE co
JOIN MPREAL pa
  ON pa.DOCA = '3'                  -- Customer Order
  AND pa.DRDN = co.ORNO             -- Order Number
  AND pa.DRDL = co.PONR             -- Line Number
  AND pa.DRDX = co.POSX             -- Line Suffix
  AND pa.AOCA = '5'                 -- Planned MO
  AND pa.deleted = 'false'
JOIN MMOPLP mop
  ON mop.PLPN = CAST(pa.ARDN AS INTEGER)  -- PLPN is integer
  AND mop.deleted = 'false'
WHERE co.deleted = 'false'
  AND mop.PSTS IN ('10', '20')      -- Active proposals
```

#### CO Line → MOP (Legacy RORC Method)
```sql
-- Legacy approach using RORC fields in MOP
SELECT co.*, mop.*
FROM OOLINE co
JOIN MMOPLP mop
  ON mop.RORC = 3                   -- Customer Order
  AND mop.RORN = co.ORNO            -- Order Number
  AND mop.RORL = co.PONR            -- Line Number
  AND mop.RORX = co.POSX            -- Line Suffix
WHERE co.deleted = 'false'
  AND mop.deleted = 'false'
  AND mop.PSTS IN ('10', '20')      -- Active proposals
```

#### Multi-Level MO Structure
```sql
-- Parent-Child MO Hierarchy
SELECT
  parent.MFNO as parent_mo,
  parent.PRNO as parent_product,
  child.MFNO as child_mo,
  child.PRNO as child_product,
  child.LEVL as level_in_structure
FROM MWOHED child
LEFT JOIN MWOHED parent
  ON parent.MFNO = child.MFLO       -- Parent MO
  AND parent.FACI = child.FACI
WHERE child.MFHL = 'MO-TOP-12345'   -- Top-level MO
  AND child.deleted = 'false'
ORDER BY child.LEVL, child.LVSQ
```

#### Order Header → Lines → Attributes
```sql
-- Complete order with lines and attributes
SELECT
  oh.ORNO,
  oh.CUNO,
  oh.ORDT,
  ol.PONR,
  ol.ITNO,
  ol.ORQT,
  ol.ATNR,
  attr.ATID,
  attr.ATAV,
  attr.ATAN
FROM OOHEAD oh
JOIN OOLINE ol
  ON ol.ORNO = oh.ORNO
  AND ol.deleted = 'false'
LEFT JOIN MOATTR attr
  ON attr.ORCA = '3'
  AND attr.RIDN = ol.ORNO
  AND attr.RIDL = ol.PONR
  AND attr.RIDX = ol.POSX
  AND attr.deleted = 'false'
WHERE oh.deleted = 'false'
  AND oh.ORNO = 'CO123456'
ORDER BY ol.PONR, attr.ANSQ
```

#### Order → Delivery Flow
```sql
-- Order to delivery relationship
SELECT
  oh.ORNO,
  oh.CUNO,
  oh.ORDT,
  ol.PONR,
  ol.ITNO,
  ol.ORQT as ordered,
  ol.DLQT as total_delivered,
  dh.DLIX,
  dh.DLDT,
  dl.DLQT as delivery_qty,
  dl.IVNO
FROM OOHEAD oh
JOIN OOLINE ol
  ON ol.ORNO = oh.ORNO
  AND ol.deleted = 'false'
LEFT JOIN ODLINE dl
  ON dl.ORNO = ol.ORNO
  AND dl.PONR = ol.PONR
  AND dl.POSX = ol.POSX
  AND dl.deleted = 'false'
LEFT JOIN ODHEAD dh
  ON dh.DLIX = dl.DLIX
  AND dh.deleted = 'false'
WHERE oh.deleted = 'false'
  AND oh.ORNO = 'CO123456'
ORDER BY ol.PONR, dh.DLIX
```

#### Complete Supply Chain View (via MPREAL)
```sql
-- Order → Pre-Allocation → MO/MOP → Delivery comprehensive view
SELECT
  oh.ORNO,
  oh.CUNO,
  oh.ORDT,
  ol.PONR,
  ol.ITNO,
  ol.ORQT,
  ol.ORST as line_status,
  pa.AOCA as supply_type,
  pa.ARDN as supply_order,
  pa.PQTY as allocated_qty,
  mo.MFNO,
  mo.WHST as mo_status,
  mo.ORQT as mo_qty,
  mo.MAQT as manufactured,
  dh.DLIX,
  dh.DLDT,
  dl.DLQT,
  dl.IVNO
FROM OOHEAD oh
JOIN OOLINE ol
  ON ol.ORNO = oh.ORNO
  AND ol.deleted = 'false'
LEFT JOIN MPREAL pa
  ON pa.DOCA = '3'
  AND pa.DRDN = ol.ORNO
  AND pa.DRDL = ol.PONR
  AND pa.DRDX = ol.POSX
  AND pa.deleted = 'false'
LEFT JOIN MWOHED mo
  ON mo.MFNO = pa.ARDN
  AND pa.AOCA = '2'  -- MO only
  AND mo.deleted = 'false'
LEFT JOIN ODLINE dl
  ON dl.ORNO = ol.ORNO
  AND dl.PONR = ol.PONR
  AND dl.POSX = ol.POSX
  AND dl.deleted = 'false'
LEFT JOIN ODHEAD dh
  ON dh.DLIX = dl.DLIX
  AND dh.deleted = 'false'
WHERE oh.deleted = 'false'
  AND oh.ORNO = 'CO123456'
ORDER BY ol.PONR, pa.ARDN, dh.DLIX
```

---

## Reference Order Category (RORC) Values

This field is **CRITICAL** for understanding relationships:

| RORC | Description | Example Use Case |
|------|-------------|------------------|
| 1 | Purchase Order | Direct purchase for demand |
| 2 | Manufacturing Order | Component for parent MO (multi-level) |
| 3 | Customer Order | Demand from customer (most common) |
| 4 | Distribution Order | Transfer between warehouses |
| 5 | MRP Proposal | Planned order (MOP) |
| 6 | Warehouse Order | Internal warehouse work |

---

## Demand-Supply Linking: RORC vs MPREAL

M3 provides **two methods** for linking demand to supply:

### Method 1: MPREAL (Pre-Allocation) - RECOMMENDED ⭐

**Table**: MPREAL
**Approach**: Explicit many-to-many linking table

**Advantages**:
- ✅ Supports many-to-many (one CO line → multiple MOs/MOPs)
- ✅ Shows allocated quantities (PQTY)
- ✅ Works with all supply types (MO, MOP, PO, DO)
- ✅ Provides allocation status
- ✅ Enables detailed supply chain visibility
- ✅ More flexible for complex scenarios

**Link Pattern**:
```sql
FROM OOLINE ol
JOIN MPREAL pa
  ON pa.DOCA = '3' AND pa.DRDN = ol.ORNO
  AND pa.DRDL = ol.PONR AND pa.DRDX = ol.POSX
JOIN MWOHED mo
  ON mo.MFNO = pa.ARDN AND pa.AOCA = '2'
```

**Use When**:
- Building supply chain visibility systems
- Need to track allocated quantities
- Multiple supply sources per demand
- Detailed planning and promising

### Method 2: RORC Fields (Legacy) - SIMPLER

**Tables**: MWOHED, MMOPLP (RORC, RORN, RORL, RORX fields)
**Approach**: Reference fields in supply order point to demand

**Advantages**:
- ✅ Simpler queries (direct join)
- ✅ Single source of truth in supply record
- ✅ Works for simple one-to-one scenarios
- ✅ Standard M3 approach

**Limitations**:
- ❌ One-to-one only (one MO → one CO line)
- ❌ No quantity allocation detail
- ❌ Only works with MO/MOP, not PO/DO
- ❌ Less flexible

**Link Pattern**:
```sql
FROM OOLINE ol
JOIN MWOHED mo
  ON mo.RORC = 3 AND mo.RORN = ol.ORNO
  AND mo.RORL = ol.PONR AND mo.RORX = ol.POSX
```

**Use When**:
- Simple one-to-one demand-supply
- Quick queries without allocation detail
- Legacy system compatibility

### Recommendation

**Use MPREAL when**:
- Building a comprehensive supply chain system
- Need full visibility into allocations
- Handling complex make-to-order scenarios
- Require allocation quantities and status

**Use RORC when**:
- Simple queries for basic linking
- Performance is critical (fewer joins)
- One-to-one relationships are guaranteed

**Best Practice**: Query **both** and compare for validation:
```sql
-- Validate RORC against MPREAL
SELECT
  mo.MFNO,
  mo.RORC, mo.RORN, mo.RORL,  -- RORC method
  pa.DOCA, pa.DRDN, pa.DRDL,  -- MPREAL method
  pa.PQTY
FROM MWOHED mo
LEFT JOIN MPREAL pa
  ON pa.AOCA = '2' AND pa.ARDN = mo.MFNO
WHERE mo.deleted = 'false'
  AND mo.RORC = 3  -- Should match pa.DOCA
```

---

## Key Field Mappings Across Tables

### Identifiers

| Concept | OOLINE (CO) | MWOHED (MO) | MMOPLP (MOP) |
|---------|-------------|-------------|--------------|
| Primary Key 1 | ORNO | MFNO | PLPN |
| Primary Key 2 | PONR | FACI | PLPS |
| Primary Key 3 | POSX | CONO | FACI |
| Item Number | ITNO | ITNO | ITNO |
| Product | - | PRNO | PRNO |
| Facility | FACI | FACI | FACI |
| Warehouse | WHLO | WHLO | WHLO |

### Status

| Concept | OOLINE (CO) | MWOHED (MO) | MMOPLP (MOP) |
|---------|-------------|-------------|--------------|
| Status | ORST | WHST | PSTS |
| Status Change Date | - | SLDT | - |
| Action Message | - | ACTP | ACTP |

### Quantities

| Concept | OOLINE (CO) | MWOHED (MO) | MMOPLP (MOP) |
|---------|-------------|-------------|--------------|
| Ordered Qty | ORQT | ORQT | PPQT |
| Remaining Qty | RNQT | - | - |
| Received Qty | - | RVQT | - |
| Manufactured Qty | - | MAQT | - |
| Unit of Measure | ALUN | MAUN | MAUN |

### Dates

| Concept | OOLINE (CO) | MWOHED (MO) | MMOPLP (MOP) |
|---------|-------------|-------------|--------------|
| Requested Date | DWDT | - | - |
| Confirmed Date | CODT | - | - |
| Start Date | - | STDT | STDT |
| Finish Date | - | FIDT | FIDT |
| Release Date | - | - | RELD |
| Planning Date | PLDT | - | PLDT |

### Reference Orders (CRITICAL!)

| Concept | OOLINE (CO) | MWOHED (MO) | MMOPLP (MOP) |
|---------|-------------|-------------|--------------|
| Reference Category | RORC | RORC | RORC |
| Reference Order | RORN | RORN | RORN |
| Reference Line | RORL | RORL | RORL |
| Reference Suffix | RORX | RORX | RORX |

### Hierarchy

| Concept | OOLINE (CO) | MWOHED (MO) | MMOPLP (MOP) |
|---------|-------------|-------------|--------------|
| Top Level | - | PRHL, MFHL | PLHL |
| Parent Level | - | PRLO, MFLO | PLLO |
| Level Number | - | LEVL | - |
| Level Sequence | - | LVSQ | - |

### Attributes

| Concept | OOLINE (CO) | MWOHED (MO) | MMOPLP (MOP) |
|---------|-------------|-------------|--------------|
| Attribute Number | ATNR | ATNR | ATNR |
| Config Number | CFIN | CFIN | CFIN |
| Built-in Attrs | ATV1-ATV0 | - | - |
| User-Defined | UCA1-UCA0 | - | - |
| User Numeric | UDN1-UDN6 | - | - |
| User Dates | UID1-UID3 | - | - |

### Change Tracking

| Concept | OOLINE (CO) | MWOHED (MO) | MMOPLP (MOP) |
|---------|-------------|-------------|--------------|
| Change Date | LMDT | LMDT | LMDT |
| Timestamp | LMTS | LMTS | LMTS |
| Changed By | CHID | CHID | CHID |
| Change Number | CHNO | CHNO | CHNO |
| Entry Date | RGDT | RGDT | RGDT |
| Deleted Flag | deleted | deleted | deleted |

---

## Data Lake Considerations

### Critical: The 'deleted' Column Issue

**IMPORTANT**: The Data Lake metadata indicates `deleted` is boolean, but the **actual data type is STRING**:

```sql
-- WRONG (will fail or return incorrect results)
WHERE deleted = false

-- CORRECT (always use string comparison)
WHERE deleted = 'false'
```

### Incremental Load Strategy

For efficient incremental loads, use:

```sql
SELECT *
FROM [TABLE]
WHERE deleted = 'false'
  AND LMDT >= 20260101        -- Change date filter
ORDER BY LMDT, LMTS           -- Ensure consistent ordering
```

### Timestamp Fields

Each table has multiple timestamp concepts:

| Timestamp Type | Purpose | Format |
|----------------|---------|--------|
| LMDT | M3 change date | Integer YYYYMMDD |
| LMTS | M3 change timestamp | Integer (microseconds) |
| timestamp | Data Lake sync time | ISO 8601 string |
| RGDT | M3 entry date | Integer YYYYMMDD |
| RGTM | M3 entry time | Integer HHMMSS |

**Best Practice**: Use LMDT + LMTS for incremental loads (M3 change time)

---

## Recommended Field Selection

### Minimal Fields for Each Table

#### CO Line (OOLINE) - Minimal Set
```sql
SELECT
  CONO, DIVI, ORNO, PONR, POSX,         -- Keys
  ITNO, ITDS,                            -- Item
  ORST,                                  -- Status
  ORQT, RNQT, DLQT, IVQT,               -- Quantities
  SAPR, NEPR, LNAM,                      -- Pricing
  DWDT, CODT,                            -- Dates
  RORC, RORN, RORL, RORX,               -- Links (CRITICAL!)
  ATNR,                                  -- Attributes
  LMDT, LMTS,                            -- Change tracking
  deleted                                -- Deletion flag
FROM OOLINE
WHERE deleted = 'false'
```

#### Manufacturing Order (MWOHED) - Minimal Set
```sql
SELECT
  CONO, FACI, MFNO,                      -- Keys
  PRNO, ITNO,                            -- Product/Item
  WHST, WHHS,                            -- Status
  ORQT, RVQT, MAQT,                      -- Quantities
  STDT, FIDT, RSDT, REFD,               -- Dates
  RORC, RORN, RORL, RORX,               -- Links (CRITICAL!)
  PRHL, MFHL, PRLO, MFLO, LEVL,         -- Hierarchy
  RESP, PLGR,                            -- Planning
  ATNR,                                  -- Attributes
  LMDT, LMTS,                            -- Change tracking
  deleted                                -- Deletion flag
FROM MWOHED
WHERE deleted = 'false'
```

#### Planned MO (MMOPLP) - Minimal Set
```sql
SELECT
  CONO, FACI, PLPN, PLPS,               -- Keys
  PRNO, ITNO,                            -- Product/Item
  PSTS, ACTP,                            -- Status/Action
  PPQT,                                  -- Quantity
  RELD, STDT, FIDT,                      -- Dates
  RORC, RORN, RORL, RORX,               -- Links (CRITICAL!)
  PLHL, PLLO,                            -- Hierarchy
  MSPM,                                  -- Material shortage
  MSG1, MSG2, MSG3, MSG4,                -- Warnings
  RESP, PLGR,                            -- Planning
  ATNR,                                  -- Attributes
  LMDT, LMTS,                            -- Change tracking
  deleted                                -- Deletion flag
FROM MMOPLP
WHERE deleted = 'false'
```

#### Pre-Allocation (MPREAL) - Critical Linking Table ⭐
```sql
SELECT
  CONO, WHLO, ITNO,                      -- Keys
  DOCA, DRDN, DRDL, DRDX,               -- Demand side (CO)
  AOCA, ARDN, ARDL, ARDX,               -- Supply side (MO/MOP)
  PQTY, PQTR,                            -- Allocated quantities
  PATY, STSB,                            -- Type and status
  RESP,                                  -- Responsible
  SCNB,                                  -- Supply chain number
  LMDT, LMTS,                            -- Change tracking
  deleted                                -- Deletion flag
FROM MPREAL
WHERE deleted = 'false'
```

#### CO Header (OOHEAD) - Minimal Set
```sql
SELECT
  CONO, ORNO,                            -- Keys
  CUNO, DECU, PYNO, INRC,               -- Customers
  ORST, ORSL,                            -- Status
  ORDT, RLDT,                            -- Dates
  CUCD, ARAT,                            -- Currency
  BRAM, NTAM,                            -- Amounts
  TEPY, MODL, TEDL,                      -- Terms
  LMDT, LMTS,                            -- Change tracking
  deleted                                -- Deletion flag
FROM OOHEAD
WHERE deleted = 'false'
```

#### Order Attributes (MOATTR) - Minimal Set
```sql
SELECT
  CONO, ATNR, ANSQ,                      -- Keys
  ORCA, RIDN, RIDL, RIDX,               -- Order reference
  ITNO,                                  -- Item
  ATID,                                  -- Attribute ID
  ATAV, ATAN,                            -- Target values
  AALF, AALT,                            -- Text range
  ANUF, ANUT,                            -- Numeric range
  ATMO,                                  -- Model
  STSB,                                  -- Status
  LMDT, LMTS,                            -- Change tracking
  deleted                                -- Deletion flag
FROM MOATTR
WHERE deleted = 'false'
  AND ORCA = '3'  -- Customer orders
```

#### Delivery Header (ODHEAD) - Minimal Set
```sql
SELECT
  CONO, DLIX,                            -- Keys
  ORNO,                                  -- Order reference
  DLDT,                                  -- Delivery date
  ORST,                                  -- Status
  IVNO, YEA4, IVDT,                      -- Invoice
  CONN, PLRI,                            -- Shipment/wave
  BRAM, NTAM,                            -- Amounts
  CUNO, DECU,                            -- Customers
  LMDT, LMTS,                            -- Change tracking
  deleted                                -- Deletion flag
FROM ODHEAD
WHERE deleted = 'false'
```

#### Delivery Line (ODLINE) - Minimal Set
```sql
SELECT
  CONO, DLIX,                            -- Keys
  ORNO, PONR, POSX,                      -- Order line reference
  ITNO,                                  -- Item
  DLQT, IVQT, CHQT,                      -- Quantities
  SAPR, NEPR, LNAM,                      -- Pricing
  IVNO, YEA4,                            -- Invoice
  UCOS, DCOS,                            -- Cost
  LMDT, LMTS,                            -- Change tracking
  deleted                                -- Deletion flag
FROM ODLINE
WHERE deleted = 'false'
```

---

## Extended Fields for Deep Analysis

### When to Include Extended Fields

#### For Financial Analysis (CO Lines)
- All discount fields (DIP1-DIP8, DIA1-DIA8, DIC1-DIC8)
- Cost fields (UCOS, COCD, UCCD)
- Currency (CUCD)
- Payment terms (TEPY)

#### For Supply Chain Performance
- All date fields (requested, confirmed, actual)
- Time fields (DWHM, COHM, MSTI, MFTI)
- Route/delivery (ROUT, RODN, MODL, TEDL)
- Status change dates (SLDT, HSDT)

#### For Attribute-Heavy Products
- ATV1-ATV0 (built-in attributes)
- UCA1-UCA0 (user alpha fields)
- UDN1-UDN6 (user numeric fields)
- UID1-UID3 (user date fields)
- ATMO, ATPR (attribute model/pricing)

#### For Multi-Level Manufacturing
- Full hierarchy (PRHL, MFHL, PRLO, MFLO, LEVL, LVSQ)
- Structure fields (SCOM, BDCD, STRT, ECVE)
- Parent references (PLLO, PLHL for MOPs)

---

## Query Performance Tips

### Indexing Strategy

**Primary Indexes (Required)**:
```sql
-- CO Lines
CREATE INDEX idx_ooline_pk ON ooline(CONO, ORNO, PONR, POSX);
CREATE INDEX idx_ooline_rorc ON ooline(RORC, RORN, RORL, RORX);
CREATE INDEX idx_ooline_lmdt ON ooline(LMDT);

-- Manufacturing Orders
CREATE INDEX idx_mwohed_pk ON mwohed(CONO, FACI, MFNO);
CREATE INDEX idx_mwohed_rorc ON mwohed(RORC, RORN, RORL, RORX);
CREATE INDEX idx_mwohed_hierarchy ON mwohed(MFHL, LEVL, LVSQ);
CREATE INDEX idx_mwohed_lmdt ON mwohed(LMDT);

-- Planned MOs
CREATE INDEX idx_mmoplp_pk ON mmoplp(CONO, FACI, PLPN, PLPS);
CREATE INDEX idx_mmoplp_rorc ON mmoplp(RORC, RORN, RORL, RORX);
CREATE INDEX idx_mmoplp_lmdt ON mmoplp(LMDT);
```

**Secondary Indexes (Recommended)**:
```sql
-- By item
CREATE INDEX idx_ooline_itno ON ooline(ITNO);
CREATE INDEX idx_mwohed_itno ON mwohed(ITNO);
CREATE INDEX idx_mmoplp_itno ON mmoplp(ITNO);

-- By status
CREATE INDEX idx_ooline_orst ON ooline(ORST);
CREATE INDEX idx_mwohed_whst ON mwohed(WHST);
CREATE INDEX idx_mmoplp_psts ON mmoplp(PSTS);

-- By facility
CREATE INDEX idx_ooline_faci ON ooline(FACI);
CREATE INDEX idx_mwohed_faci ON mwohed(FACI);
CREATE INDEX idx_mmoplp_faci ON mmoplp(FACI);
```

### Query Optimization

1. **Always filter deleted records first**:
   ```sql
   WHERE deleted = 'false'  -- Most selective filter
     AND [other conditions]
   ```

2. **Use LMDT ranges for incremental loads**:
   ```sql
   WHERE deleted = 'false'
     AND LMDT >= 20260101
     AND LMDT <= 20260131
   ```

3. **Leverage composite indexes for joins**:
   ```sql
   -- Good: Uses composite index
   WHERE RORC = 3 AND RORN = 'CO123' AND RORL = 1

   -- Bad: Partial index usage
   WHERE RORN = 'CO123'
   ```

4. **Order matters for timestamp consistency**:
   ```sql
   ORDER BY LMDT, LMTS  -- Ensures deterministic ordering
   ```

---

## Common Analysis Queries

### 1. Full Supply Chain View (CO → MPREAL → MO → Status)

```sql
-- Using MPREAL for accurate allocation tracking
SELECT
  co.ORNO as customer_order,
  co.PONR as co_line,
  co.ITNO as item,
  co.ORQT as ordered_qty,
  co.DLQT as delivered_qty,
  co.ORST as co_status,
  co.DWDT as requested_date,
  co.CODT as confirmed_date,
  pa.AOCA as supply_type,
  pa.PQTY as allocated_qty,
  mo.MFNO as mfg_order,
  mo.WHST as mo_status,
  mo.ORQT as mo_qty,
  mo.MAQT as manufactured_qty,
  mo.STDT as mo_start,
  mo.FIDT as mo_finish,
  mo.REFD as mo_actual_finish
FROM OOLINE co
LEFT JOIN MPREAL pa
  ON pa.DOCA = '3'
  AND pa.DRDN = co.ORNO
  AND pa.DRDL = co.PONR
  AND pa.DRDX = co.POSX
  AND pa.AOCA = '2'  -- MO only
  AND pa.deleted = 'false'
LEFT JOIN MWOHED mo
  ON mo.MFNO = pa.ARDN
  AND mo.deleted = 'false'
WHERE co.deleted = 'false'
  AND co.ORST IN ('20', '30', '40')  -- Active CO lines
ORDER BY co.ORNO, co.PONR, pa.ARDN
```

### 2. Planning Coverage with Allocation Detail (CO → MPREAL → MOP/MO)

```sql
-- Complete supply coverage including allocated quantities
SELECT
  co.ORNO,
  co.PONR,
  co.ITNO,
  co.ORQT as demand_qty,
  co.RNQT as remaining_qty,
  co.DWDT as need_date,
  -- MOP allocations
  pa_mop.ARDN as mop_number,
  pa_mop.PQTY as mop_allocated_qty,
  mop.PSTS as mop_status,
  mop.ACTP as mop_action,
  mop.FIDT as mop_finish_date,
  -- MO allocations
  pa_mo.ARDN as mo_number,
  pa_mo.PQTY as mo_allocated_qty,
  mo.WHST as mo_status,
  mo.MAQT as manufactured_qty,
  -- Coverage summary
  COALESCE(pa_mop.PQTY, 0) + COALESCE(pa_mo.PQTY, 0) as total_allocated,
  co.ORQT - COALESCE(pa_mop.PQTY, 0) - COALESCE(pa_mo.PQTY, 0) as uncovered_qty,
  CASE
    WHEN mo.MFNO IS NOT NULL THEN 'Confirmed'
    WHEN mop.PLPN IS NOT NULL THEN 'Planned'
    ELSE 'No Coverage'
  END as supply_coverage
FROM OOLINE co
LEFT JOIN MPREAL pa_mop
  ON pa_mop.DOCA = '3'
  AND pa_mop.DRDN = co.ORNO
  AND pa_mop.DRDL = co.PONR
  AND pa_mop.DRDX = co.POSX
  AND pa_mop.AOCA = '5'  -- Planned MO
  AND pa_mop.deleted = 'false'
LEFT JOIN MMOPLP mop
  ON mop.PLPN = CAST(pa_mop.ARDN AS INTEGER)
  AND mop.deleted = 'false'
  AND mop.PSTS IN ('10', '20')
LEFT JOIN MPREAL pa_mo
  ON pa_mo.DOCA = '3'
  AND pa_mo.DRDN = co.ORNO
  AND pa_mo.DRDL = co.PONR
  AND pa_mo.DRDX = co.POSX
  AND pa_mo.AOCA = '2'  -- Manufacturing Order
  AND pa_mo.deleted = 'false'
LEFT JOIN MWOHED mo
  ON mo.MFNO = pa_mo.ARDN
  AND mo.deleted = 'false'
WHERE co.deleted = 'false'
  AND co.ORST < '90'
ORDER BY
  CASE WHEN uncovered_qty > 0 THEN 0 ELSE 1 END,  -- Uncovered first
  co.DWDT,  -- By due date
  co.ORNO, co.PONR
```

### 3. Multi-Level BOM Explosion

```sql
WITH RECURSIVE mo_hierarchy AS (
  -- Top level
  SELECT
    MFNO, PRNO, ITNO, LEVL, LVSQ,
    ORQT, WHST,
    MFHL, MFLO,
    RORC, RORN, RORL,
    CAST(MFNO AS VARCHAR(500)) as path
  FROM MWOHED
  WHERE MFHL = 'MO-TOP-123'  -- Top-level MO
    AND deleted = 'false'

  UNION ALL

  -- Children
  SELECT
    child.MFNO, child.PRNO, child.ITNO,
    child.LEVL, child.LVSQ,
    child.ORQT, child.WHST,
    child.MFHL, child.MFLO,
    child.RORC, child.RORN, child.RORL,
    CAST(parent.path || ' → ' || child.MFNO AS VARCHAR(500))
  FROM MWOHED child
  JOIN mo_hierarchy parent
    ON child.MFLO = parent.MFNO
    AND child.FACI = parent.FACI
  WHERE child.deleted = 'false'
)
SELECT * FROM mo_hierarchy
ORDER BY LEVL, LVSQ
```

---

## Attribute Handling Recommendations

### Option 1: Structured JSONB (Recommended)

Store all attributes in a single JSONB column:

```json
{
  "builtin_numeric": {
    "ATV1": 150.5,
    "ATV2": 200.0
  },
  "builtin_string": {
    "ATV6": "Red",
    "ATV7": "Large"
  },
  "user_alpha": {
    "UCA1": "CustomField1",
    "UCA2": "CustomField2"
  },
  "user_numeric": {
    "UDN1": 1500.00,
    "UDN2": 2000.00
  },
  "user_dates": {
    "UID1": "2026-01-15"
  }
}
```

**Query Example**:
```sql
SELECT *
FROM co_lines
WHERE attributes->'builtin_string'->>'ATV6' = 'Red'
  AND (attributes->'user_numeric'->>'UDN1')::numeric > 1000
```

### Option 2: Separate Columns for Common Attributes

For frequently queried attributes, create dedicated columns:

```sql
CREATE TABLE co_lines (
  -- ... standard fields ...

  -- Commonly used attributes as columns
  color VARCHAR(20),          -- From ATV6
  size VARCHAR(20),           -- From ATV7
  custom_priority NUMERIC,    -- From UDN1
  custom_date DATE,           -- From UID1

  -- All others in JSONB
  other_attributes JSONB
);
```

---

## File Structure

```
schema-maps/
├── README.md                          ← Quick start guide
├── SCHEMA_MAP_SUMMARY.md              ← This file - comprehensive overview
├── RECOMMENDED_SCHEMA_DESIGN.md       ← Production-ready target schema ⭐
│
├── Order Management
│   ├── CO_HEADER_SCHEMA_MAP.md        ← Customer Order Header (OOHEAD)
│   ├── CO_LINE_SCHEMA_MAP.md          ← Customer Order Lines (OOLINE)
│   └── ATTRIBUTE_SCHEMA_MAP.md        ← Order Attributes (MOATTR)
│
├── Manufacturing
│   ├── MO_SCHEMA_MAP.md               ← Manufacturing Orders (MWOHED)
│   └── MOP_SCHEMA_MAP.md              ← Planned MOs (MMOPLP)
│
├── Supply Chain Linking
│   └── PREALLOCATION_SCHEMA_MAP.md    ← Pre-Allocation (MPREAL) ⭐ CRITICAL
│
└── Delivery & Fulfillment
    ├── DELIVERY_HEADER_SCHEMA_MAP.md  ← Delivery Header (ODHEAD)
    └── DELIVERY_LINE_SCHEMA_MAP.md    ← Delivery Lines (ODLINE)
```

---

## Quick Reference Card

### Most Critical Fields - Core Tables

| Purpose | CO Header | CO Line | MO | MOP |
|---------|-----------|---------|-----|-----|
| **Primary Key** | ORNO | ORNO+PONR+POSX | FACI+MFNO | FACI+PLPN+PLPS |
| **Item** | - | ITNO | ITNO, PRNO | ITNO, PRNO |
| **Status** | ORST, ORSL | ORST | WHST | PSTS, ACTP |
| **Quantity** | - | ORQT, RNQT | ORQT, MAQT | PPQT |
| **Key Date** | ORDT, RLDT | DWDT, CODT | STDT, FIDT | RELD, FIDT |
| **Customer** | CUNO, DECU | CUNO | - | - |
| **Link to Demand** | - | RORC=ref | RORC+RORN+RORL | RORC+RORN+RORL |
| **Change Tracking** | LMDT, LMTS | LMDT, LMTS | LMDT, LMTS | LMDT, LMTS |
| **Hierarchy** | - | - | MFHL, MFLO, LEVL | PLHL, PLLO |
| **Attributes** | UCA1-0, UDN1-6 | ATV1-0, UCA1-0 | ATNR, CFIN | ATNR, CFIN |

### Most Critical Fields - Supporting Tables

| Purpose | MOATTR | MPREAL | Delivery Header | Delivery Line |
|---------|--------|---------|-----------------|---------------|
| **Primary Key** | ATNR+ANSQ | DOCA+DRDN+DRDL+AOCA+ARDN | DLIX | DLIX+ORNO+PONR+POSX |
| **Links To** | ORCA+RIDN+RIDL | Demand+Supply sides | ORNO | ORNO+PONR, DLIX |
| **Key Field 1** | ATID (attr name) | DOCA, DRDN (demand) | DLDT | DLQT, IVQT |
| **Key Field 2** | ATAV, ATAN (value) | AOCA, ARDN (supply) | IVNO, YEA4 | LNAM |
| **Key Field 3** | ATNR (instance) | PQTY (allocated qty) | CONN (shipment) | CHQT (variance) |
| **Change Tracking** | LMDT, LMTS | LMDT, LMTS | LMDT, LMTS | LMDT, LMTS |

### Common MPREAL Linking Patterns (RECOMMENDED)

```sql
-- CO Line → MO allocations
SELECT * FROM MPREAL
WHERE DOCA = '3' AND DRDN = [ORNO] AND DRDL = [PONR] AND DRDX = [POSX]
  AND AOCA = '2'  -- Manufacturing Orders

-- CO Line → MOP allocations
SELECT * FROM MPREAL
WHERE DOCA = '3' AND DRDN = [ORNO] AND DRDL = [PONR] AND DRDX = [POSX]
  AND AOCA = '5'  -- Planned MOs

-- CO Line → PO allocations
SELECT * FROM MPREAL
WHERE DOCA = '3' AND DRDN = [ORNO] AND DRDL = [PONR] AND DRDX = [POSX]
  AND AOCA = '1'  -- Purchase Orders

-- MO → CO Lines it fulfills
SELECT * FROM MPREAL
WHERE AOCA = '2' AND ARDN = [MFNO]
  AND DOCA = '3'  -- Customer Orders
```

### Legacy RORC Patterns (Still Valid)

```sql
-- CO Line driven by customer
WHERE RORC IS NULL OR RORC = 0

-- MO for customer order
WHERE RORC = 3 AND RORN = [CO_NUMBER]

-- MO for parent MO (multi-level)
WHERE RORC = 2 AND RORN = [PARENT_MO]

-- MOP from MRP
WHERE RORC = 3 AND RORN = [CO_NUMBER]
```

---

## Additional Resources

### Data Catalog Access
```javascript
// Using infor-mcp tools
mcp__infor-m3__datacatalog_get_table_info({
  tableName: "OOLINE"  // or "OOHEAD", "MOATTR", "MWOHED", "MMOPLP", "MPREAL", "ODHEAD", "ODLINE"
})
```

### M3 API Programs
- **OOHEAD**: OIS100MI (GetOrderHead, etc.)
- **OOLINE**: OIS100MI (GetOrderLine, LstOrderLines, etc.)
- **MOATTR**: CRS980MI (Attribute APIs)
- **MWOHED**: PMS100MI (GetMO, LstMOs, etc.)
- **MMOPLP**: PPS280MI (GetProposal, LstProposals, etc.)
- **MPREAL**: MMS080MI (Pre-allocation APIs - GetPreAlloc, LstPreAlloc, etc.)
- **ODHEAD**: OIS350MI (GetDelivery, LstDeliveries, etc.)
- **ODLINE**: OIS350MI (GetDeliveryLine, etc.)

---

## Version History

| Date | Version | Changes |
|------|---------|---------|
| 2026-01-21 | 1.0 | Initial schema maps created (OOLINE, MWOHED, MMOPLP) |
| 2026-01-21 | 1.1 | Added order header, attributes, and delivery tables (OOHEAD, MOATTR, ODHEAD, ODLINE) |
| 2026-01-21 | 1.2 | Added MPREAL (Pre-Allocation) - critical demand→supply linking table |
| 2026-01-21 | 1.3 | Added RECOMMENDED_SCHEMA_DESIGN.md - complete production-ready database schema with DDL, ETL strategy, and performance tuning |

---

## Support and Feedback

For questions, corrections, or enhancements to these schema maps:
1. Reference the source Data Catalog metadata
2. Verify field usage in your M3 environment
3. Test queries in a non-production environment first
4. Document any tenant-specific customizations separately

---

**Remember**: These are standard M3 fields. Your implementation may have:
- Different field usage patterns
- Custom fields beyond standard M3
- Industry-specific attributes
- Modified workflows affecting status values

Always validate against your specific M3 configuration and business rules.
