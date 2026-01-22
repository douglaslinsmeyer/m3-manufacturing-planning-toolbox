# Pre-Allocation (MPREAL) - Schema Map

## Table Overview
**M3 Table**: MPREAL
**Description**: Pre-allocation - links demand orders to acquisition/supply orders
**Record Count**: 44 fields
**Critical Purpose**: **This is THE linking table between customer orders and their supply sources (MOs, MOPs, POs, DOs)**

---

## Core Purpose

MPREAL creates the linkage between:
- **Demand Side**: Customer orders (or other demand sources)
- **Supply Side**: Manufacturing orders, planned MOs, purchase orders, distribution orders

**Think of it as**: "This quantity from this supply order is reserved for this demand order"

---

## Core Identifiers

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| CONO | integer | Company Number | Company identifier (max: 999) |
| WHLO | string | Warehouse | Warehouse for allocation |
| ITNO | string | Item Number | Item being allocated |

---

## Demand Side (Customer Order)

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| DOCA | string | Demand Order Category | Type of demand order |
| DRDN | string | Demand Order Number | Demand order number |
| DRDL | integer | Demand Order Line | Demand order line (max: 999999) |
| DRDX | integer | Demand Line Suffix | Demand line suffix (max: 999) |
| DLPS | integer | Demand Order Sub Number | Demand sub number (max: 999) |
| DLP2 | integer | Demand Order Sub Number 2 | Demand sub number 2 (max: 999) |

**DOCA (Demand Order Category) Values:**
- **3** = Customer Order (OOLINE) - most common
- **2** = Manufacturing Order (component for parent MO)
- **4** = Distribution Order
- **6** = Warehouse Order

**Linking to OOLINE**: DOCA='3', DRDN=ORNO, DRDL=PONR, DRDX=POSX

---

## Supply Side (Acquisition Order)

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| AOCA | string | Acquisition Order Category | Type of supply order |
| ARDN | string | Acquisition Order Number | Supply order number |
| ARDL | integer | Acquisition Order Line | Supply order line (max: 999999) |
| ARDX | integer | Acquisition Line Suffix | Supply line suffix (max: 999) |
| ALPS | integer | Acquisition Order Sub Number | Supply sub number (max: 999) |
| ALP2 | integer | Acquisition Order Sub Number 2 | Supply sub number 2 (max: 999) |

**AOCA (Acquisition Order Category) Values:**
- **1** = Purchase Order (MPHEAD/MPLINE)
- **2** = Manufacturing Order (MWOHED)
- **4** = Distribution Order
- **5** = Planned MO (MMOPLP)
- **6** = Warehouse Order

**Linking to MWOHED**: AOCA='2', ARDN=MFNO
**Linking to MMOPLP**: AOCA='5', ARDN=PLPN (as string)

---

## Allocation Quantities

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| PQTY | number | Preallocated Quantity | Quantity preallocated (basic U/M) |
| PQTR | number | Preallocated Quantity Reserve | Reserved/remaining quantity |

**Note**: PQTY is the key field showing how much of the supply is reserved for the demand.

---

## Status and Type

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| PATY | string | Preallocation Type | Type of preallocation |
| STSB | string | Preallocation Status | Status code |

**Common PATY Values:**
- **1** = Manual preallocation
- **2** = Automatic (system-generated)

**Common STSB Values:**
- **10** = Active
- **90** = Closed/Completed

---

## Notification Flags

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| NTFC | integer | Notify When Changed | Notify responsible when allocation changed (max: 9) |
| NTFD | integer | Notify When Deleted | Notify responsible when allocation deleted (max: 9) |
| NTDC | integer | Notify Demand Changed | Notify if demand order changed (max: 9) |
| NTDD | integer | Notify Demand Deleted | Notify if demand order deleted (max: 9) |
| NTAC | integer | Notify Acquisition Changed | Notify if supply order changed (max: 9) |
| NTAD | integer | Notify Acquisition Deleted | Notify if supply order deleted (max: 9) |

**Use Case**: Alert planner when linked orders change or are deleted.

---

## Backorder Flags

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| ARBD | integer | Acquisition Backorder | Supply is on backorder (max: 9) |
| DRBD | integer | Demand Backorder | Demand is on backorder (max: 9) |

---

## Responsible Person

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| RESP | string | Responsible | Person/planner responsible |

---

## Supply Chain Details

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| SCNB | string | Supply Chain Number | Supply chain identifier |
| SCPO | string | Supply Chain Policy | Supply chain policy code |

---

## Demand Structure Details

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| DMSQ | integer | Demand Material Sequence | Material sequence in demand (max: 9999) |
| DOPN | integer | Demand Operation Number | Operation number in demand (max: 9999) |
| DSSQ | integer | Demand Structure Sequence | Structure sequence number (max: 9999999999) |

**Use Case**: For complex multi-level structures and operations.

---

## M3 Audit Fields

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| RGDT | integer | Entry Date | Record creation date (YYYYMMDD) |
| RGTM | integer | Entry Time | Record creation time (HHMMSS) |
| LMDT | integer | Change Date | Last modification date (YYYYMMDD) |
| CHNO | integer | Change Number | Sequential change counter (max: 999) |
| CHID | string | Changed By | User who last modified |
| LMTS | integer | Timestamp | Last modification timestamp (microseconds) |

---

## Data Lake Metadata

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| accountingEntity | string | Accounting Entity | Record accounting entity |
| variationNumber | integer | Variation Number | Record modification sequence |
| timestamp | string (datetime) | Modification Timestamp | Record modification time (ISO format) |
| deleted | boolean | Is Deleted | Record deletion flag (STRING: "true"/"false") |
| archived | boolean | Is Archived | Record archive flag |

---

## Field Count by Category

| Category | Count | Description |
|----------|-------|-------------|
| Core Identifiers | 3 | Company, warehouse, item |
| Demand Side | 6 | **Links to customer orders** |
| Supply Side | 6 | **Links to MOs/MOPs/POs** |
| Quantities | 2 | Preallocated amounts |
| Status/Type | 2 | Allocation status |
| Notifications | 6 | Alert configurations |
| Backorder | 2 | Backorder status |
| Responsible | 1 | Ownership |
| Supply Chain | 2 | Supply chain info |
| Demand Structure | 3 | Structure details |
| Audit/Metadata | 11 | Change tracking + Data Lake |
| **Total** | **44** | All fields |

---

## Key Fields for Data Modeling

### Primary Key (Composite)
- CONO + WHLO + ITNO + DOCA + DRDN + DRDL + DRDX + AOCA + ARDN + ARDL + ARDX

**Note**: This is a many-to-many linking table with a complex composite key.

### Foreign Keys / Relationships
- **To Customer Order Line**: DOCA='3', DRDN=ORNO, DRDL=PONR, DRDX=POSX (links to OOLINE)
- **To Manufacturing Order**: AOCA='2', ARDN=MFNO (links to MWOHED)
- **To Planned MO**: AOCA='5', ARDN=PLPN (links to MMOPLP)
- **To Purchase Order**: AOCA='1', ARDN=PUNO (links to MPHEAD/MPLINE)
- **To Item**: ITNO
- **To Warehouse**: WHLO

### Critical for Incremental Load
- LMDT (Change Date)
- LMTS (Timestamp)
- deleted (must use string comparison: 'false')

---

## Common Query Patterns

### 1. Find Supply for a Customer Order Line

```sql
-- What MOs/MOPs are supplying this customer order line?
SELECT
  pa.AOCA as supply_type,
  pa.ARDN as supply_order,
  pa.PQTY as allocated_qty,
  pa.STSB as status,
  CASE pa.AOCA
    WHEN '1' THEN 'Purchase Order'
    WHEN '2' THEN 'Manufacturing Order'
    WHEN '4' THEN 'Distribution Order'
    WHEN '5' THEN 'Planned MO'
    ELSE 'Other'
  END as supply_description
FROM MPREAL pa
WHERE pa.deleted = 'false'
  AND pa.DOCA = '3'           -- Customer Order
  AND pa.DRDN = 'CO123456'    -- Order Number
  AND pa.DRDL = 1             -- Line Number
  AND pa.DRDX = 0             -- Line Suffix
ORDER BY pa.AOCA, pa.ARDN
```

### 2. Find Demand for a Manufacturing Order

```sql
-- What customer orders is this MO fulfilling?
SELECT
  pa.DOCA as demand_type,
  pa.DRDN as demand_order,
  pa.DRDL as demand_line,
  pa.PQTY as allocated_qty,
  pa.STSB as status
FROM MPREAL pa
WHERE pa.deleted = 'false'
  AND pa.AOCA = '2'           -- Manufacturing Order
  AND pa.ARDN = 'MO-123456'   -- MO Number
ORDER BY pa.DRDN, pa.DRDL
```

### 3. Complete CO Line to MO Linkage

```sql
-- Customer order line with its manufacturing orders
SELECT
  ol.ORNO,
  ol.PONR,
  ol.ITNO,
  ol.ITDS,
  ol.ORQT as ordered,
  pa.PQTY as allocated_to_mo,
  pa.ARDN as mo_number,
  mo.WHST as mo_status,
  mo.ORQT as mo_quantity,
  mo.MAQT as manufactured
FROM OOLINE ol
LEFT JOIN MPREAL pa
  ON pa.DOCA = '3'
  AND pa.DRDN = ol.ORNO
  AND pa.DRDL = ol.PONR
  AND pa.DRDX = ol.POSX
  AND pa.AOCA = '2'           -- MO
  AND pa.deleted = 'false'
LEFT JOIN MWOHED mo
  ON mo.MFNO = pa.ARDN
  AND mo.deleted = 'false'
WHERE ol.deleted = 'false'
  AND ol.ORNO = 'CO123456'
ORDER BY ol.PONR, pa.ARDN
```

### 4. MO to CO Line Linkage (Reverse)

```sql
-- Manufacturing order with its customer order demands
SELECT
  mo.MFNO,
  mo.PRNO,
  mo.ORQT as mo_qty,
  mo.WHST as mo_status,
  pa.PQTY as allocated_qty,
  pa.DRDN as customer_order,
  pa.DRDL as co_line,
  ol.CUNO,
  ol.ITNO,
  ol.ORQT as ordered_qty
FROM MWOHED mo
LEFT JOIN MPREAL pa
  ON pa.AOCA = '2'
  AND pa.ARDN = mo.MFNO
  AND pa.DOCA = '3'           -- Customer Order
  AND pa.deleted = 'false'
LEFT JOIN OOLINE ol
  ON ol.ORNO = pa.DRDN
  AND ol.PONR = pa.DRDL
  AND ol.POSX = pa.DRDX
  AND ol.deleted = 'false'
WHERE mo.deleted = 'false'
  AND mo.MFNO = 'MO-123456'
ORDER BY pa.DRDN, pa.DRDL
```

### 5. Planned MO (MOP) Allocations

```sql
-- Customer orders linked to planned MOs
SELECT
  ol.ORNO,
  ol.PONR,
  ol.ITNO,
  ol.ORQT,
  pa.PQTY as allocated_to_mop,
  pa.ARDN as mop_number,
  mop.PSTS as mop_status,
  mop.PPQT as planned_qty,
  mop.ACTP as action_message
FROM OOLINE ol
JOIN MPREAL pa
  ON pa.DOCA = '3'
  AND pa.DRDN = ol.ORNO
  AND pa.DRDL = ol.PONR
  AND pa.DRDX = ol.POSX
  AND pa.AOCA = '5'           -- Planned MO
  AND pa.deleted = 'false'
JOIN MMOPLP mop
  ON mop.PLPN = CAST(pa.ARDN AS INTEGER)
  AND mop.deleted = 'false'
WHERE ol.deleted = 'false'
  AND ol.ORNO = 'CO123456'
ORDER BY ol.PONR, pa.ARDN
```

### 6. Allocation Coverage Analysis

```sql
-- How much of the order is covered by allocations?
SELECT
  ol.ORNO,
  ol.PONR,
  ol.ITNO,
  ol.ORQT as ordered_qty,
  ol.ALQT as allocated_from_stock,
  COALESCE(SUM(pa.PQTY), 0) as preallocated_from_supply,
  ol.ORQT - ol.ALQT - COALESCE(SUM(pa.PQTY), 0) as uncovered_qty
FROM OOLINE ol
LEFT JOIN MPREAL pa
  ON pa.DOCA = '3'
  AND pa.DRDN = ol.ORNO
  AND pa.DRDL = ol.PONR
  AND pa.DRDX = ol.POSX
  AND pa.deleted = 'false'
WHERE ol.deleted = 'false'
  AND ol.ORNO = 'CO123456'
GROUP BY ol.ORNO, ol.PONR, ol.ITNO, ol.ORQT, ol.ALQT
ORDER BY ol.PONR
```

### 7. Supply Chain Linkage Report

```sql
-- Complete supply chain view with allocations
SELECT
  oh.ORNO,
  oh.CUNO,
  oh.ORDT,
  ol.PONR,
  ol.ITNO,
  ol.ORQT,
  pa.AOCA,
  pa.ARDN,
  pa.PQTY,
  CASE pa.AOCA
    WHEN '1' THEN 'PO'
    WHEN '2' THEN 'MO'
    WHEN '4' THEN 'DO'
    WHEN '5' THEN 'MOP'
    ELSE 'Unknown'
  END as supply_type,
  pa.STSB as allocation_status
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
WHERE oh.deleted = 'false'
  AND oh.ORDT >= 20260101
ORDER BY oh.ORNO, ol.PONR, pa.AOCA, pa.ARDN
```

---

## Understanding Pre-Allocation

### What is Pre-Allocation?

Pre-allocation is the **explicit linking** of supply to demand:

```
Customer Order Line (Demand)
    ↓ MPREAL links
Manufacturing Order (Supply)
```

### Why It's Critical

1. **Visibility**: Shows which MO is making goods for which customer
2. **Planning**: Helps MRP understand demand coverage
3. **Promising**: Enables accurate delivery date promises
4. **Tracking**: Follows supply through production to demand
5. **Prioritization**: Helps prioritize production based on customer orders

### How It Works

```
CO Line: ORNO="CO123", PONR=1, Qty=100
    ↓
MPREAL: DOCA='3', DRDN='CO123', DRDL=1, PQTY=60
    ↓
MO: MFNO="MO-456", Qty=60

MPREAL: DOCA='3', DRDN='CO123', DRDL=1, PQTY=40
    ↓
MOP: PLPN=789, Qty=40
```

The 100-unit order is covered by:
- 60 units from an active MO
- 40 units from a planned MO

---

## Order Category Reference

### DOCA (Demand Order Category)

| Code | Description | Table | Key Field |
|------|-------------|-------|-----------|
| 2 | Manufacturing Order | MWOHED | MFNO |
| 3 | Customer Order | OOLINE | ORNO + PONR |
| 4 | Distribution Order | - | - |
| 6 | Warehouse Order | - | - |

### AOCA (Acquisition Order Category)

| Code | Description | Table | Key Field |
|------|-------------|-------|-----------|
| 1 | Purchase Order | MPHEAD/MPLINE | PUNO |
| 2 | Manufacturing Order | MWOHED | MFNO |
| 4 | Distribution Order | - | - |
| 5 | Planned MO | MMOPLP | PLPN |
| 6 | Warehouse Order | - | - |

---

## Many-to-Many Relationships

MPREAL enables many-to-many:

### One CO Line → Many Supply Orders

```
CO Line (100 units)
    ├─→ MO-1 (40 units)
    ├─→ MO-2 (35 units)
    └─→ MOP-3 (25 units)
```

### One MO → Many CO Lines

```
MO-123 (1000 units)
    ├─→ CO-A Line 1 (300 units)
    ├─→ CO-A Line 2 (200 units)
    ├─→ CO-B Line 1 (400 units)
    └─→ Stock (100 units)
```

---

## Allocation Status Lifecycle

```
STSB = 10 (Active)
    ↓
Supply order progresses
    ↓
Goods delivered/completed
    ↓
STSB = 90 (Closed)
```

---

## Notification Use Cases

Set notification flags to alert planners:

- **NTDC=1**: Alert if customer changes their order
- **NTAC=1**: Alert if MO is changed
- **NTAD=1**: Alert if MO is deleted
- **NTDD=1**: Alert if customer cancels

---

## Allocation Type (PATY)

| Code | Description | When Used |
|------|-------------|-----------|
| 1 | Manual | Planner explicitly links demand to supply |
| 2 | Automatic | MRP/system automatically creates link |

---

## Data Model Recommendation

### Normalized Linking Table

```sql
CREATE TABLE preallocation_links (
    id BIGSERIAL PRIMARY KEY,

    -- Core
    company INTEGER,
    warehouse VARCHAR(10),
    item_number VARCHAR(20),

    -- Demand side
    demand_category VARCHAR(10),
    demand_order VARCHAR(20),
    demand_line INTEGER,
    demand_suffix INTEGER,

    -- Supply side
    supply_category VARCHAR(10),
    supply_order VARCHAR(20),
    supply_line INTEGER,
    supply_suffix INTEGER,

    -- Quantities
    preallocated_qty NUMERIC(15,6),
    preallocated_qty_reserve NUMERIC(15,6),

    -- Status
    preallocation_type VARCHAR(10),
    status VARCHAR(10),

    -- Control
    responsible VARCHAR(20),
    supply_chain_number VARCHAR(20),

    -- Metadata
    lmdt DATE,
    lmts BIGINT,
    deleted BOOLEAN,

    UNIQUE(company, warehouse, item_number,
           demand_category, demand_order, demand_line, demand_suffix,
           supply_category, supply_order, supply_line, supply_suffix)
);

-- Critical indexes
CREATE INDEX idx_prealloc_demand ON preallocation_links(
    demand_category, demand_order, demand_line, demand_suffix
);

CREATE INDEX idx_prealloc_supply ON preallocation_links(
    supply_category, supply_order, supply_line, supply_suffix
);

CREATE INDEX idx_prealloc_item ON preallocation_links(item_number);
CREATE INDEX idx_prealloc_status ON preallocation_links(status);
```

---

## Common Pitfalls

### 1. String vs Integer Keys

**CRITICAL**: ARDN and DRDN are **strings**, but some reference integer keys:
- MFNO (MO Number) is a string
- PLPN (Planned Order Number) is an integer stored as string
- PUNO (Purchase Order) is a string

**Solution**: Always cast when joining:
```sql
-- Correct for MMOPLP
WHERE mop.PLPN = CAST(pa.ARDN AS INTEGER)

-- Correct for MWOHED
WHERE mo.MFNO = pa.ARDN  -- Both strings
```

### 2. Suffix Fields

Don't forget DRDX and ARDX (line suffixes) when joining to OOLINE!

```sql
-- Correct
ON pa.DRDN = ol.ORNO
   AND pa.DRDL = ol.PONR
   AND pa.DRDX = ol.POSX

-- Wrong (missing suffix)
ON pa.DRDN = ol.ORNO
   AND pa.DRDL = ol.PONR
```

### 3. Multiple Allocations

Remember: One demand line can have **multiple** supply allocations. Always use LEFT JOIN or aggregate:

```sql
-- Correct
SELECT ol.*, SUM(pa.PQTY) as total_preallocated
FROM OOLINE ol
LEFT JOIN MPREAL pa ON ...
GROUP BY ol.ORNO, ol.PONR
```

---

## Incremental Load Strategy

```sql
SELECT *
FROM MPREAL
WHERE deleted = 'false'
  AND LMDT >= 20260101  -- Change date filter
ORDER BY LMDT, LMTS
```

---

## Performance Considerations

### Recommended Indexes

```sql
-- Demand lookup (most common)
CREATE INDEX idx_mpreal_demand ON mpreal(DOCA, DRDN, DRDL, DRDX);

-- Supply lookup
CREATE INDEX idx_mpreal_supply ON mpreal(AOCA, ARDN, ARDL, ARDX);

-- Item lookup
CREATE INDEX idx_mpreal_item ON mpreal(ITNO);

-- Warehouse lookup
CREATE INDEX idx_mpreal_whlo ON mpreal(WHLO);

-- Status
CREATE INDEX idx_mpreal_status ON mpreal(STSB);

-- Change tracking
CREATE INDEX idx_mpreal_lmdt ON mpreal(LMDT);

-- Composite for CO to MO
CREATE INDEX idx_mpreal_co_to_mo ON mpreal(
    DOCA, DRDN, DRDL, AOCA
) WHERE DOCA = '3' AND AOCA IN ('2', '5');
```

---

## Best Practices

1. **Always check STSB**: Only active allocations (STSB='10') are valid
2. **Sum quantities**: One demand can have multiple supply allocations
3. **Handle nulls**: Not all demand has preallocations (make-to-stock)
4. **Check both directions**: Query from demand→supply AND supply→demand
5. **Include ITNO**: Helps validate the link makes sense
6. **Use WHLO**: Allocations are warehouse-specific
7. **Cast carefully**: ARDN/DRDN are strings but may reference integer keys
8. **Don't forget suffixes**: DRDX and ARDX are part of the key

---

## Integration with Other Tables

```sql
-- Complete view: Order → Allocation → MO → Delivery
SELECT
  oh.ORNO,
  oh.CUNO,
  ol.PONR,
  ol.ITNO,
  ol.ORQT,
  pa.PQTY as preallocated,
  pa.ARDN as mo_number,
  mo.MAQT as manufactured,
  dh.DLIX,
  dl.DLQT as delivered
FROM OOHEAD oh
JOIN OOLINE ol ON ol.ORNO = oh.ORNO AND ol.deleted = 'false'
LEFT JOIN MPREAL pa
  ON pa.DOCA = '3'
  AND pa.DRDN = ol.ORNO
  AND pa.DRDL = ol.PONR
  AND pa.DRDX = ol.POSX
  AND pa.AOCA = '2'
  AND pa.deleted = 'false'
LEFT JOIN MWOHED mo
  ON mo.MFNO = pa.ARDN
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

## Summary

**MPREAL is the critical linking table** that connects:
- Customer orders to manufacturing orders
- Customer orders to planned MOs
- Customer orders to purchase orders
- Any demand to any supply

Without MPREAL, you cannot trace which production/purchase is fulfilling which customer order. This table enables:
- Supply chain visibility
- Order promising
- Production prioritization
- Demand-driven planning

**Key takeaway**: Always join through MPREAL when you need to understand the demand-supply relationship in M3.
