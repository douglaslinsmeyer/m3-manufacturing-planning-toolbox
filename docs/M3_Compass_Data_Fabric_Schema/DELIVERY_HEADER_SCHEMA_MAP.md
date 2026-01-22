# Customer Order Delivery Header (ODHEAD) - Schema Map

## Table Overview
**M3 Table**: ODHEAD
**Description**: Customer Order Delivery Header - contains header-level details for each delivery
**Record Count**: 87 fields

---

## Core Identifiers

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| CONO | integer | Company Number | Company identifier (max: 999) |
| DIVI | string | Division | Division code within company |
| ORNO | string | Order Number | Customer order number |
| DLIX | integer | Delivery Number | Unique delivery identifier (max: 99999999999) |

**Note**: DLIX is the primary key for deliveries. One order (ORNO) can have multiple deliveries.

---

## Facility and Warehouse

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| FACI | string | Facility | Facility where goods were picked |
| WHLO | string | Warehouse | Warehouse for this delivery |

---

## Order Information

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| ORTP | string | Order Type | Customer order type |
| ORST | string | Highest Status | Highest status of delivery lines |

**Common ORST Values:**
- 33 = Picked
- 44 = Delivered
- 55 = Invoiced
- 66 = Confirmed

---

## Customer Information

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| CUNO | string | Customer Number | Ordering customer |
| DECU | string | Delivery Customer | Customer receiving delivery |
| PYNO | string | Payer | Customer paying for delivery |
| INRC | string | Invoice Recipient | Customer receiving invoice |

---

## Delivery Dates and Times

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| DLDT | integer | Planned Delivery Date | Planned/actual delivery date (YYYYMMDD) |
| DLTM | integer | Delivery Time | Time of delivery (HHMMSS) |
| RELD | integer | Release Date | Date delivery was released (YYYYMMDD) |

---

## Invoice Information

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| IVNO | integer | Invoice Number | Invoice number for delivery (max: 999999999) |
| YEA4 | integer | Invoice Year | Year of invoice (max: 9999) |
| IVDT | integer | Invoice Date | Date invoice created (YYYYMMDD) |
| INTM | integer | Invoicing Time | Time invoice created (HHMMSS) |
| ACDT | integer | Accounting Date | Date for accounting entries (YYYYMMDD) |

---

## Invoice Prefix and Extended Number

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| INPX | string | Invoice Prefix | Invoice number prefix |
| EXIN | string | Extended Invoice Number | Extended/alternate invoice number |

---

## Currency and Exchange

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| CUCD | string | Currency | Currency for this delivery |
| RAIN | number | Invoice Exchange Rate | Exchange rate used for invoicing |
| FECN | string | Future Rate Agreement | Forward exchange contract reference |

---

## Financial Amounts

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| BRAM | number | Gross Amount | Delivery value before discounts |
| NTAM | number | Net Amount | Net delivery value after discounts |

---

## Weights and Measures

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| GRWE | number | Gross Weight | Total gross weight of delivery |
| NEWE | number | Net Weight | Total net weight of delivery |
| VOL3 | number | Volume | Total volume of delivery |

---

## Delivery Terms

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| MODL | string | Delivery Method | Method of delivery (truck, ship, air) |
| TEDL | string | Delivery Terms | Incoterms (FOB, CIF, etc.) |

---

## Address and Route

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| ADID | string | Address Number | Delivery address identifier |
| ROUT | string | Route | Delivery route code |
| RODN | integer | Route Departure | Route departure number (max: 999) |

---

## Shipment and Wave

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| CONN | integer | Shipment Number | Shipment identifier (max: 9999999) |
| PLRI | string | Wave Number | Pick wave number |
| DNNO | string | Delivery Note Number | Delivery note/packing slip number |
| CDNU | integer | Chronological Delivery Note | Sequential delivery note number (max: 9999999999) |
| CDDE | integer | Delivery Note Creation Date | Date chrono number created (YYYYMMDD) |

---

## Priority

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| OPRI | integer | Priority | Delivery priority (1-9, lower = higher) |

---

## Invoicing Control

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| AICD | integer | Summary Invoice | Create summary invoice (max: 9) |
| IVGP | string | Invoicing Group | Grouping code for invoicing |

---

## Statistics Updates

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| UPST | integer | Update Sales Statistics | Update stats flag (max: 9) |
| UPIS | integer | Update Intrastat | Update Intrastat flag (max: 9) |

---

## Approval

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| APBY | string | Approved By | User who approved delivery |
| APDT | integer | Approval Date | Date delivery approved (YYYYMMDD) |
| ORS1 | integer | Delivery Approval Required | Requires approval flag (max: 9) |

---

## Pricing

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| ORS2 | integer | Recalculate Preliminary Price | Recalc price flag (max: 9) |
| ORS3 | integer | Preliminary Charges | Has preliminary charges (max: 9) |

---

## Text and Documents

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| TXID | integer | Text Identity | Header text reference |
| PRTX | integer | Pre-Text Identity | Text before delivery |
| POTX | integer | Post-Text Identity | Text after delivery |
| DTID | integer | Document Identity | Document reference |

---

## Project Management

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| PROJ | string | Project Number | Project identifier |
| ELNO | string | Project Element | Project element/WBS code |

---

## Status History

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| PRST | string | Previous Status | Status before current |

---

## Payment Terms

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| TEPY | string | Payment Terms | Payment terms for this delivery |

---

## Reference Number

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| NREF | string | Reference Number | General reference number |

---

## 3rd Party Provider

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| 3RDP | string | 3rd Party Provider | Third-party logistics provider |

---

## Customer Channel

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| CUCH | string | Customer Channel ID | Customer channel identifier |
| CCAC | string | Activity | Activity code |

---

## VAT Reporting

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| VRGD | integer | VAT on GDNI Reported | VAT reported flag (max: 9) |

---

## Responsible Person

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| RLBY | string | Responsible | Person responsible for delivery |

---

## Return Information (Credit Deliveries)

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| RYEA | integer | Reference Year | Year of referenced invoice (max: 9999) |
| RIVN | integer | Reference Invoice Number | Referenced invoice number (max: 999999999) |
| RDLX | integer | Corrected Delivery Number | Original delivery being corrected (max: 99999999999) |
| ADLX | integer | Actual Delivery Number | Actual delivery number for corrections (max: 99999999999) |
| DBCR | string | Debit/Credit Code | Type of financial transaction |
| CIME | integer | Corrective Method | Method for corrections (max: 9) |
| RINP | string | Reference Invoice Prefix | Prefix of referenced invoice |
| RXIN | string | Reference Extended Invoice | Extended number of reference |

---

## Transaction Reason

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| RSCD | string | Transaction Reason | Reason code for delivery |

---

## Goods Responsibility

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| GRN0 | integer | Goods Responsibility Not Transferred | Ownership not yet transferred (max: 9) |

---

## Modified Tax Date

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| MTXD | integer | Modified Tax Date | Adjusted tax date (YYYYMMDD) |

---

## Return Warehouse

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| RWHL | string | Return Warehouse | Warehouse for returns |

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
| Core Identifiers | 4 | Delivery identification |
| Facility/Warehouse | 2 | Location info |
| Order Info | 2 | Order type and status |
| Customer Info | 4 | Customer references |
| Dates/Times | 3 | Delivery dates |
| Invoice Info | 5 | Invoice details |
| Invoice Extensions | 2 | Extended invoice numbers |
| Currency | 3 | Currency and exchange |
| Amounts | 2 | Financial totals |
| Weights/Measures | 3 | Physical dimensions |
| Delivery Terms | 2 | Delivery method and terms |
| Address/Route | 3 | Delivery destination |
| Shipment/Wave | 5 | Warehouse operations |
| Priority | 1 | Priority level |
| Invoicing Control | 2 | Invoice grouping |
| Statistics | 2 | Reporting flags |
| Approval | 3 | Approval tracking |
| Pricing | 2 | Price control |
| Text/Documents | 4 | Text references |
| Project | 2 | Project references |
| Status History | 1 | Previous status |
| Payment | 1 | Payment terms |
| Return Info | 8 | Credit memo details |
| Transaction Reason | 1 | Reason codes |
| Special Flags | 4 | Various control flags |
| Audit/Metadata | 11 | Change tracking + Data Lake |
| **Total** | **87** | All fields |

---

## Key Fields for Data Modeling

### Primary Key
- CONO + DLIX

### Foreign Keys / Relationships
- **To Order Header**: ORNO (links to OOHEAD)
- **To Order Lines**: ORNO (one delivery has many lines via ODLINE)
- **To Customer**: CUNO
- **To Delivery Customer**: DECU
- **To Payer**: PYNO
- **To Invoice Recipient**: INRC
- **To Shipment**: CONN
- **To Facility**: FACI
- **To Warehouse**: WHLO

### Critical for Incremental Load
- LMDT (Change Date)
- LMTS (Timestamp)
- deleted (must use string comparison: 'false')

### Most Commonly Used Fields
1. DLIX, ORNO (identification)
2. DLDT (delivery date)
3. ORST (status)
4. IVNO, YEA4 (invoice reference)
5. BRAM, NTAM (amounts)
6. CONN (shipment grouping)
7. LMDT (change tracking)

---

## Delivery Lifecycle

```
Order Lines Created (OOLINE)
    ↓
Picking List Created
    ↓
Wave Released (PLRI)
    ↓
ODHEAD Created (Delivery Header)
    ↓
ODLINE Created (Delivery Lines)
    ↓
Goods Issued (ORST=44)
    ↓
Invoice Created (IVNO populated)
    ↓
Delivery Confirmed (ORST=66)
```

---

## Common Query Patterns

### Get Delivery with Order Info
```sql
SELECT
  dh.DLIX,
  dh.ORNO,
  dh.DLDT,
  dh.ORST,
  dh.IVNO,
  dh.YEA4,
  dh.BRAM,
  dh.NTAM,
  oh.CUNO,
  oh.CUOR as customer_po
FROM ODHEAD dh
JOIN OOHEAD oh
  ON oh.ORNO = dh.ORNO
  AND oh.deleted = 'false'
WHERE dh.deleted = 'false'
  AND dh.DLIX = 123456789
```

### Deliveries by Date Range
```sql
SELECT
  DLIX,
  ORNO,
  DLDT,
  ORST,
  CONN,
  BRAM,
  NTAM
FROM ODHEAD
WHERE deleted = 'false'
  AND DLDT >= 20260101
  AND DLDT <= 20260131
ORDER BY DLDT, DLIX
```

### Deliveries Not Yet Invoiced
```sql
SELECT
  DLIX,
  ORNO,
  DLDT,
  ORST,
  BRAM,
  NTAM
FROM ODHEAD
WHERE deleted = 'false'
  AND ORST = '44'        -- Delivered
  AND (IVNO IS NULL OR IVNO = 0)  -- Not invoiced
ORDER BY DLDT
```

### Shipment Summary
```sql
SELECT
  CONN as shipment_number,
  COUNT(*) as delivery_count,
  SUM(GRWE) as total_weight,
  SUM(VOL3) as total_volume,
  SUM(BRAM) as total_value
FROM ODHEAD
WHERE deleted = 'false'
  AND CONN > 0
  AND DLDT >= 20260101
GROUP BY CONN
ORDER BY shipment_number
```

### Deliveries Requiring Approval
```sql
SELECT
  DLIX,
  ORNO,
  DLDT,
  BRAM,
  RLBY as responsible
FROM ODHEAD
WHERE deleted = 'false'
  AND ORS1 = 1           -- Approval required
  AND (APBY IS NULL OR APBY = '')  -- Not approved
ORDER BY DLDT
```

---

## Relationship to Delivery Lines

```sql
-- Get delivery with all lines
SELECT
  dh.DLIX,
  dh.ORNO,
  dh.DLDT,
  dh.BRAM as header_amount,
  dl.PONR,
  dl.ITNO,
  dl.DLQT,
  dl.LNAM as line_amount
FROM ODHEAD dh
JOIN ODLINE dl
  ON dl.DLIX = dh.DLIX
  AND dl.deleted = 'false'
WHERE dh.deleted = 'false'
  AND dh.DLIX = 123456789
ORDER BY dl.PONR, dl.POSX
```

---

## Multiple Deliveries per Order

An order can have multiple deliveries:

```sql
SELECT
  ORNO,
  COUNT(*) as delivery_count,
  MIN(DLDT) as first_delivery,
  MAX(DLDT) as last_delivery,
  SUM(BRAM) as total_delivered_value
FROM ODHEAD
WHERE deleted = 'false'
  AND ORNO = 'CO123456'
GROUP BY ORNO
```

---

## Invoice to Delivery Relationship

```sql
-- Get all deliveries for an invoice
SELECT
  DLIX,
  ORNO,
  DLDT,
  BRAM,
  NTAM
FROM ODHEAD
WHERE deleted = 'false'
  AND YEA4 = 2026
  AND IVNO = 123456
ORDER BY DLIX
```

**Note**: Multiple deliveries can be on one invoice (AICD=1 for summary invoice).

---

## Shipment Consolidation

Deliveries are grouped into shipments via CONN:

```sql
-- Get all deliveries in a shipment
SELECT
  dh.DLIX,
  dh.ORNO,
  dh.CUNO,
  dh.ADID,
  dh.BRAM,
  dh.GRWE
FROM ODHEAD dh
WHERE dh.deleted = 'false'
  AND dh.CONN = 789456
ORDER BY dh.DLIX
```

**Use Case**: One truck/container carries multiple customer deliveries.

---

## Credit Deliveries (Returns)

For credit memos and returns:

```sql
SELECT
  DLIX as credit_delivery,
  RDLX as original_delivery,
  RIVN as original_invoice,
  RYEA as original_year,
  DBCR as debit_credit,
  NTAM as credit_amount
FROM ODHEAD
WHERE deleted = 'false'
  AND RDLX > 0           -- References original delivery
  AND NTAM < 0           -- Negative amount (credit)
ORDER BY RDLX
```

---

## Wave Picking

Wave number (PLRI) groups deliveries for efficient picking:

```sql
-- Get all deliveries in a pick wave
SELECT
  DLIX,
  ORNO,
  DLDT,
  ORST
FROM ODHEAD
WHERE deleted = 'false'
  AND PLRI = 'WAVE-2026-001'
ORDER BY OPRI, DLIX
```

---

## Delivery Status Flow

```
33 = Picked         (goods allocated)
    ↓
44 = Delivered      (goods issued, can invoice)
    ↓
55 = Invoiced       (invoice created)
    ↓
66 = Confirmed      (customer acknowledged)
```

---

## Chronological Delivery Note

CDNU is a sequential number across all deliveries:

```sql
SELECT
  CDNU as sequential_number,
  CDDE as assigned_date,
  DLIX,
  ORNO,
  DNNO as delivery_note
FROM ODHEAD
WHERE deleted = 'false'
  AND CDDE >= 20260101
ORDER BY CDNU
```

**Use Case**: Legal requirement for sequential numbering of delivery notes.

---

## VAT on Delivery vs Invoice

```
VRGD = 0: VAT calculated on invoice
VRGD = 1: VAT already reported on delivery
```

Important for:
- Cross-border deliveries
- Tax reporting compliance
- EU Intrastat

---

## Incremental Load Strategy

```sql
SELECT *
FROM ODHEAD
WHERE deleted = 'false'
  AND LMDT >= 20260101  -- Change date filter
ORDER BY LMDT, LMTS
```

---

## Performance Considerations

### Recommended Indexes

```sql
-- Primary key
CREATE INDEX idx_odhead_pk ON odhead(CONO, DLIX);

-- Order reference
CREATE INDEX idx_odhead_orno ON odhead(ORNO);

-- Delivery date
CREATE INDEX idx_odhead_dldt ON odhead(DLDT);

-- Status
CREATE INDEX idx_odhead_orst ON odhead(ORST);

-- Invoice reference
CREATE INDEX idx_odhead_invoice ON odhead(YEA4, IVNO);

-- Shipment
CREATE INDEX idx_odhead_conn ON odhead(CONN);

-- Wave
CREATE INDEX idx_odhead_plri ON odhead(PLRI);

-- Customer
CREATE INDEX idx_odhead_cuno ON odhead(CUNO);

-- Change tracking
CREATE INDEX idx_odhead_lmdt ON odhead(LMDT);
```

---

## Delivery Approval Workflow

When ORS1 = 1:

```
Delivery Created
    ↓
Review Required
    ↓
APBY = Approver
    ↓
APDT = Approval Date
    ↓
Can Proceed to Invoice
```

---

## Summary Invoice Handling

When AICD = 1:

- Multiple deliveries combine into one invoice
- Group by IVGP (invoicing group)
- One IVNO shared across deliveries
- Summarized on customer invoice

```sql
-- Get summary invoice deliveries
SELECT
  IVNO,
  YEA4,
  IVGP,
  COUNT(*) as delivery_count,
  SUM(BRAM) as total_gross,
  SUM(NTAM) as total_net
FROM ODHEAD
WHERE deleted = 'false'
  AND AICD = 1
  AND YEA4 = 2026
  AND IVNO = 123456
GROUP BY IVNO, YEA4, IVGP
```

---

## Integration Points

### To Order Management
```
OOHEAD ← ORNO → ODHEAD
```

### To Invoicing
```
ODHEAD (IVNO, YEA4) → Invoice System
```

### To Warehouse
```
Pick Wave (PLRI) → ODHEAD ← CONN (Shipment)
```

### To Transportation
```
ODHEAD (CONN, ROUT, RODN) → Shipping/TMS
```

---

## Best Practices

1. **Always link via ORNO and DLIX**: These are your primary keys
2. **Check status (ORST)**: Determines what operations are allowed
3. **Validate invoice data**: IVNO may be 0 until invoiced
4. **Handle returns carefully**: Check RDLX and DBCR for credits
5. **Use shipment grouping**: CONN for logistics optimization
6. **Track approval status**: ORS1, APBY, APDT for workflow
7. **Consider multi-delivery**: One order can have many deliveries
