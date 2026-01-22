# Customer Order Header (OOHEAD) - Schema Map

## Table Overview
**M3 Table**: OOHEAD
**Description**: Customer Order Header - contains all header-level details for customer orders
**Record Count**: 163 fields

---

## Core Identifiers

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| CONO | integer | Company Number | Company identifier (max: 999) |
| DIVI | string | Division | Division code within company |
| ORNO | string | Order Number | Customer order number (unique identifier) |
| ORTP | string | Order Type | Customer order type code |

---

## Facility and Warehouse

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| FACI | string | Facility | Default facility for order |
| WHLO | string | Warehouse | Default warehouse for order |

---

## Status

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| ORST | string | Highest Order Status | Highest status across all order lines |
| ORSL | string | Lowest Order Status | Lowest status across all order lines |

**Common Status Values:**
- 15 = Entered
- 20 = Released
- 33 = Picked
- 44 = Delivered
- 66 = Invoiced
- 77 = Confirmed
- 90 = Closed

---

## Business Chain Hierarchy

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| CHL1 | string | Business Chain Level 1 | Business chain hierarchy level 1 |
| CHL2 | string | Business Chain Level 2 | Business chain hierarchy level 2 |
| CHL3 | string | Business Chain Level 3 | Business chain hierarchy level 3 |
| CHL4 | string | Business Chain Level 4 | Business chain hierarchy level 4 |
| CHL5 | string | Business Chain Level 5 | Business chain hierarchy level 5 |
| CHL6 | string | Business Chain Level 6 | Business chain hierarchy level 6 |
| CHL7 | string | Business Chain Level 7 | Business chain hierarchy level 7 |
| CHL8 | string | Business Chain Level 8 | Business chain hierarchy level 8 |
| CHL9 | string | Business Chain Level 9 | Business chain hierarchy level 9 |

**Note**: Business chains represent organizational hierarchies (e.g., Region → District → Store)

---

## Customer Information

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| CUNO | string | Customer Number | Customer placing the order |
| DECU | string | Delivery Customer | Customer receiving the delivery |
| PYNO | string | Payer | Customer responsible for payment |
| INRC | string | Invoice Recipient | Customer receiving the invoice |

**Note**: These can all be different for drop-ship or third-party billing scenarios

---

## Dates (Header Level)

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| ORDT | integer | Order Date | Date order was placed (YYYYMMDD) |
| CUDT | integer | Customer PO Date | Date on customer's purchase order (YYYYMMDD) |
| RLDT | integer | Requested Delivery Date | Customer requested delivery date (YYYYMMDD) |
| RLHM | integer | Requested Delivery Time | Customer requested time (HHMM) |
| RLDZ | integer | Requested Delivery Date (TZ) | Requested date with timezone (YYYYMMDD) |
| RLHZ | integer | Requested Delivery Time (TZ) | Requested time with timezone (HHMM) |
| TIZO | string | Time Zone | Time zone code for dates/times |
| DMDT | integer | Manual Due Date | Manually set due date (YYYYMMDD) |
| CURD | integer | Value Date | Currency valuation date (YYYYMMDD) |
| FDDT | integer | Earliest Delivery Date | Earliest possible delivery date (YYYYMMDD) |
| FDED | integer | First Delivery Date | Date of first partial delivery (YYYYMMDD) |
| LDED | integer | Last Delivery Date | Date of last delivery (YYYYMMDD) |

---

## Priority and Control

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| OPRI | integer | Order Priority | Priority level (1-9, lower = higher) |
| OBLC | integer | Order Stop Code | Customer order hold/stop flag (max: 9) |
| HOCD | integer | Order Entry in Progress | Order being entered flag (max: 9) |

---

## Payment and Financial Terms

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| TEPY | string | Payment Terms | Payment terms code |
| PYCD | string | Payment Method | AR payment method |
| TECD | string | Cash Discount Term | Cash discount terms |
| BKID | string | Bank Account ID | Bank account identifier |
| PYRE | string | Payment Request Reference | Reference for payment request |

---

## Delivery Terms

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| MODL | string | Delivery Method | Method of delivery (truck, ship, air, etc.) |
| TEDL | string | Delivery Terms | Incoterms (FOB, CIF, etc.) |
| TEL2 | string | Delivery Terms Text | Additional delivery terms description |
| TEPA | string | Packaging Terms | Packaging terms code |

---

## Language and Communication

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| LNCD | string | Language | Language code for documents |
| WCON | string | Contact Method | Preferred contact method |

---

## Addresses and Routing

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| ADID | string | Address Number | Delivery address identifier |
| ROUT | string | Route | Delivery route code |
| RODN | integer | Route Departure | Route departure number (max: 999) |
| DLSP | string | Delivery Specification | Delivery specification code |
| DSTX | string | Delivery Description | Description/instructions for delivery |

---

## References

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| OREF | string | Our Reference | Internal reference contact |
| YREF | string | Your Reference 1 | Customer's reference contact |
| CUOR | string | Customer Order Number | Customer's PO number |
| NREF | string | Reference Number | Additional reference number |

---

## Sales and Marketing

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| SMCD | string | Salesperson | Sales representative code |
| OFNO | string | Quotation Number | Quote that was converted to order |
| VRCD | string | Business Type (TST) | Trade statistics business type |
| ECLC | string | Labor Code (TST) | Trade statistics labor code |
| FRE1 | string | Statistics Identity 1 | Statistics classification 1 |

---

## Project Management

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| PROJ | string | Project Number | Project identifier |
| ELNO | string | Project Element | Project element/WBS code |

---

## Agreements and Campaigns

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| AGNO | string | Blanket Agreement Number | Blanket order agreement |
| BAGC | string | Agreement Customer | Customer on blanket agreement |
| BAGD | integer | Agreement Start Date | Blanket agreement start (YYYYMMDD) |
| BREC | string | Recipient B/C Agreement | Bonus/commission agreement recipient |
| AGNT | string | Agreement Type 1 | Commission agreement type |

---

## Pricing and Discounts

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| PLTB | string | Price List Table | Price list table code |
| DISY | string | Discount Model | Discount model code |
| EXCD | string | Service Charge | Service charge code |
| CHSY | string | Line Charge Model | Line charge model |
| CMPN | string | Discount Campaign | Discount campaign identifier |
| DICD | integer | Discount Origin | Source of discount (max: 9) |
| OTDP | number | Order Total Discount % | Percentage for order-level discount |
| OTBA | number | Order Total Discount Base | Base amount for order discount |

---

## Currency and Exchange

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| LOCD | string | Local Currency | Local currency code |
| CUCD | string | Order Currency | Currency for this order |
| DCCD | integer | Decimal Places | Number of decimals for currency (max: 9) |
| CRTP | integer | Exchange Rate Type | Type of exchange rate (max: 99) |
| FECN | string | Future Rate Agreement | FRA number for hedging |
| ARAT | number | Exchange Rate | Exchange rate used |
| DMCU | integer | Currency Conversion Method | Method for conversion (max: 9) |

---

## Tax and VAT

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| TINC | integer | VAT Included | VAT included in prices flag (max: 9) |
| ECTT | integer | EU Triangular Trade | EU triangular trade flag (max: 9) |
| TAXC | string | Tax Code | Customer/address tax code |
| TXAP | integer | Tax Applicable | Tax applicable flag (max: 9) |
| VTCD | integer | VAT Code | Value added tax code (max: 99) |

---

## Amounts and Totals

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| BRAM | number | Order Value Gross | Total order value before discounts |
| BRLA | number | Total Order Value Gross | Total gross value |
| NTAM | number | Net Order Value | Net order value after discounts |
| NTLA | number | Total Order Value Net | Total net value |
| COAM | number | Total Order Cost | Total cost of order |
| TOPR | number | Total Price | Total price |
| TBLG | number | Not Used | Legacy field (not used) |

---

## Advance Invoicing

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| RPIV | number | Remains to Invoice in Advance | Remaining advance invoice amount |
| IPIV | number | Amount of Advance Invoice | Advance invoice amount |
| IAPD | number | Invoice Amount - Previous | Invoice amount from prior deliveries |
| VAPD | number | Remaining VAT to Invoice | Remaining VAT amount |
| PRP2 | integer | Prepayment Process | Prepayment process flag (max: 9) |

---

## Weights and Measures

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| GRWE | number | Gross Weight | Total gross weight of order |
| NEWE | number | Net Weight | Total net weight of order |
| VOL3 | number | Volume | Total volume of order |

---

## Invoicing

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| AICD | integer | Summary Invoice | Create summary invoice flag (max: 9) |

---

## Statistics

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| OT38 | integer | Update Order Received Stats | Update statistics flag (max: 9) |

---

## Customs and Export

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| CPRE | string | Customs Procedure - Export | Export customs procedure code |
| HAFE | string | Harbor or Airport | Export port/airport code |

---

## Text and Documents

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| TXID | integer | Text Identity | Header text reference |
| PRTX | integer | Pre-Text Identity | Text before order |
| POTX | integer | Post-Text Identity | Text after order |
| DTID | integer | Document Identity | Document reference |

---

## Job/Batch Processing

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| JNA | string | Job Name | Batch job name |
| JNU | integer | Job Number | Batch job number (max: 999999) |

---

## Supply Model

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| SPLM | string | Supply Model | Supply model code |
| BLRO | number | Backlog Rounding | Backlog rounding amount |
| ABNO | integer | Abnormal Demand | Abnormal demand flag (max: 9) |
| SCED | integer | Delivery Regrouping | Delivery regrouping flag (max: 9) |

---

## Line Counts

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| NBNS | integer | Non-Inventory Lines Count | Number of non-stock lines (max: 99999) |

---

## CTP (Capable to Promise)

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| VCTP | integer | Validate CTP | Validate capable to promise (max: 9) |

---

## Responsible Person

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| RESP | string | Responsible | Person responsible for order |

---

## Supplier Rebate

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| PCLA | number | Supplier Rebate | Supplier rebate amount |

---

## Customer Channel

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| CUCH | string | Customer Channel ID | Customer channel identifier |
| CCAC | string | Activity | Activity code |

---

## Rail Transport

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| RASN | string | Rail Station | Railway station code |

---

## Original Invoice Reference

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| OIVR | string | Original Invoice Reference | Reference to original invoice |
| OYEA | integer | Original Year | Year of original invoice (max: 9999) |

---

## Migration Status

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| MIGI | string | Internal Migration Status | Data migration status flag |

---

## Internal Transfer

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| ICTR | integer | Internal Sales | Internal transfer flag (max: 9) |

---

## Trade Agreement

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| TAGY | string | Trade Agreement Model | Trade agreement model code |

---

## User-Defined Fields (Alpha)

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| UCA1 | string | User Alpha Field 1 | Custom text field 1 |
| UCA2 | string | User Alpha Field 2 | Custom text field 2 |
| UCA3 | string | User Alpha Field 3 | Custom text field 3 |
| UCA4 | string | User Alpha Field 4 | Custom text field 4 |
| UCA5 | string | User Alpha Field 5 | Custom text field 5 |
| UCA6 | string | User Alpha Field 6 | Custom text field 6 |
| UCA7 | string | User Alpha Field 7 | Custom text field 7 |
| UCA8 | string | User Alpha Field 8 | Custom text field 8 |
| UCA9 | string | User Alpha Field 9 | Custom text field 9 |
| UCA0 | string | User Alpha Field 10 | Custom text field 10 |

---

## User-Defined Fields (Numeric)

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| UDN1 | number | User Numeric Field 1 | Custom number field 1 |
| UDN2 | number | User Numeric Field 2 | Custom number field 2 |
| UDN3 | number | User Numeric Field 3 | Custom number field 3 |
| UDN4 | number | User Numeric Field 4 | Custom number field 4 |
| UDN5 | number | User Numeric Field 5 | Custom number field 5 |
| UDN6 | number | User Numeric Field 6 | Custom number field 6 |

---

## User-Defined Fields (Dates)

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| UID1 | integer | User Date Field 1 | Custom date field 1 (YYYYMMDD) |
| UID2 | integer | User Date Field 2 | Custom date field 2 (YYYYMMDD) |
| UID3 | integer | User Date Field 3 | Custom date field 3 (YYYYMMDD) |

---

## User-Defined Fields (Text)

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| UCT1 | string | User Text Field 1 | Custom long text field 1 |

---

## Finance Reason Code

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| FRSC | string | Finance Reason Code | Financial reason/classification code |

---

## 3rd Party Provider

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| 3RDP | string | 3rd Party Provider | Third-party logistics provider |
| IPAD | string | IP Address | IP address for web orders |

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
| Core Identifiers | 4 | Basic order identification |
| Status | 2 | Order status levels |
| Business Chain | 9 | Organizational hierarchy |
| Customer Info | 4 | Customer, payer, invoice recipient |
| Dates | 12 | All date fields |
| Priority/Control | 3 | Priority and stop codes |
| Payment/Financial | 5 | Payment terms and methods |
| Delivery Terms | 4 | Delivery method and terms |
| References | 4 | Customer and internal references |
| Sales/Marketing | 5 | Sales rep, stats, quotation |
| Pricing/Discounts | 8 | Price lists and discount models |
| Currency | 7 | Currency and exchange rates |
| Tax/VAT | 5 | Tax and VAT handling |
| Amounts | 7 | Order value totals |
| Advance Invoicing | 5 | Prepayment and advances |
| Weights/Measures | 3 | Physical dimensions |
| Text/Documents | 4 | Text and document references |
| Project | 2 | Project references |
| Agreements | 5 | Blanket agreements and campaigns |
| User-Defined | 20 | Custom fields (alpha, numeric, date, text) |
| Supply Model | 4 | Supply and backlog management |
| Audit/Metadata | 11 | Change tracking + Data Lake |
| **Total** | **163** | All fields |

---

## Key Fields for Data Modeling

### Primary Key
- CONO + ORNO

### Foreign Keys / Relationships
- **To Customer**: CUNO
- **To Delivery Customer**: DECU
- **To Payer**: PYNO
- **To Invoice Recipient**: INRC
- **To Order Lines**: ORNO (one-to-many with OOLINE)
- **To Deliveries**: ORNO (one-to-many with ODHEAD)
- **To Facility**: FACI
- **To Warehouse**: WHLO

### Critical for Incremental Load
- LMDT (Change Date)
- LMTS (Timestamp)
- deleted (must use string comparison: 'false')

### Most Commonly Used Fields
1. ORNO (order identification)
2. CUNO, DECU (customer info)
3. ORST, ORSL (status)
4. ORDT, RLDT (dates)
5. CUCD, ARAT (currency)
6. BRAM, NTAM (amounts)
7. LMDT (change tracking)

---

## Common Query Patterns

### Get Order Header with Customer Info
```sql
SELECT
  oh.ORNO,
  oh.ORST,
  oh.CUNO,
  oh.DECU,
  oh.ORDT,
  oh.RLDT,
  oh.BRAM,
  oh.NTAM,
  oh.CUCD
FROM OOHEAD oh
WHERE oh.deleted = 'false'
  AND oh.ORST < '90'  -- Not closed
ORDER BY oh.ORDT DESC
```

### Orders by Status Range
```sql
SELECT *
FROM OOHEAD
WHERE deleted = 'false'
  AND ORST >= '20'  -- Released
  AND ORST < '66'   -- Not yet invoiced
  AND ORDT >= 20260101
ORDER BY RLDT, OPRI
```

### Customer Order Summary
```sql
SELECT
  CUNO,
  COUNT(*) as order_count,
  SUM(BRAM) as total_gross,
  SUM(NTAM) as total_net
FROM OOHEAD
WHERE deleted = 'false'
  AND ORST < '90'
  AND ORDT >= 20260101
GROUP BY CUNO
ORDER BY total_net DESC
```

### Orders with Blanket Agreements
```sql
SELECT
  oh.*,
  ol.line_count
FROM OOHEAD oh
LEFT JOIN (
  SELECT ORNO, COUNT(*) as line_count
  FROM OOLINE
  WHERE deleted = 'false'
  GROUP BY ORNO
) ol ON ol.ORNO = oh.ORNO
WHERE oh.deleted = 'false'
  AND oh.AGNO IS NOT NULL
  AND oh.AGNO != ''
ORDER BY oh.BAGD DESC
```

---

## Relationship to Order Lines

The header-to-line relationship:

```sql
-- Get complete order with all lines
SELECT
  h.ORNO,
  h.CUNO,
  h.ORST as header_status,
  h.ORDT,
  h.BRAM as header_gross_total,
  l.PONR,
  l.ITNO,
  l.ORQT,
  l.SAPR,
  l.LNAM,
  l.ORST as line_status
FROM OOHEAD h
JOIN OOLINE l
  ON l.ORNO = h.ORNO
  AND l.deleted = 'false'
WHERE h.deleted = 'false'
  AND h.ORNO = 'CO123456'
ORDER BY l.PONR, l.POSX
```

---

## Status Rollup Logic

The header status fields represent aggregated line statuses:

- **ORST (Highest Status)**: Maximum status value across all lines
- **ORSL (Lowest Status)**: Minimum status value across all lines

This allows quick determination of order progress without querying all lines.

---

## Multi-Customer Scenario

A single order can involve multiple customer entities:

```
Order ORNO = "CO123456"
├─ CUNO = "CUST001"      (Ordering customer)
├─ DECU = "CUST002"      (Delivery customer - ship to)
├─ PYNO = "CUST003"      (Payer - bill to)
└─ INRC = "CUST004"      (Invoice recipient)
```

This is common in:
- Drop-ship scenarios
- Third-party billing
- Franchise operations
- Parent-subsidiary relationships

---

## Business Chain Usage

Business chains (CHL1-CHL9) represent organizational hierarchies:

**Example: Retail Chain**
```
CHL1 = "REGION-NORTH"
CHL2 = "DISTRICT-05"
CHL3 = "STORE-0234"
CHL4 = "DEPT-ELECTRONICS"
```

**Example: Manufacturing**
```
CHL1 = "BUSINESS-UNIT-A"
CHL2 = "PRODUCT-LINE-1"
CHL3 = "CUSTOMER-SEGMENT-ENTERPRISE"
```

Use these for:
- Hierarchical reporting
- Sales analysis by territory
- Commission calculations
- Performance tracking

---

## Currency Handling

For multi-currency orders:

1. **LOCD**: Company's local/base currency
2. **CUCD**: Order currency (customer's currency)
3. **ARAT**: Exchange rate at order entry
4. **FECN**: Forward exchange contract (hedging)
5. **CRTP**: Type of rate (spot, forward, etc.)

**Important**: Exchange rates are typically locked at order entry, not updated.

---

## Date Hierarchy

Understanding the date fields:

1. **ORDT**: When order was entered in system
2. **CUDT**: Date on customer's PO
3. **RLDT**: Customer's requested delivery date
4. **FDDT**: Earliest system can deliver
5. **DMDT**: Manual override due date
6. **FDED**: Actual first delivery date
7. **LDED**: Actual last delivery date

---

## User-Defined Fields Strategy

The 20 user-defined fields allow customization:

**Common Uses:**
- UCA1-UCA10: Sales region, territory, special instructions
- UDN1-UDN6: Margins, targets, custom calculations
- UID1-UID3: Campaign start/end, special event dates
- UCT1: Extended notes or JSON data

**Best Practice**: Document your usage in a data dictionary!

---

## Incremental Load Query

```sql
SELECT *
FROM OOHEAD
WHERE deleted = 'false'
  AND LMDT >= 20260101  -- Change date filter
ORDER BY LMDT, LMTS
```

---

## Performance Considerations

### Recommended Indexes

```sql
-- Primary key
CREATE INDEX idx_oohead_pk ON oohead(CONO, ORNO);

-- Status queries
CREATE INDEX idx_oohead_status ON oohead(ORST, ORSL);

-- Customer queries
CREATE INDEX idx_oohead_cuno ON oohead(CUNO);
CREATE INDEX idx_oohead_decu ON oohead(DECU);
CREATE INDEX idx_oohead_pyno ON oohead(PYNO);

-- Date queries
CREATE INDEX idx_oohead_ordt ON oohead(ORDT);
CREATE INDEX idx_oohead_rldt ON oohead(RLDT);

-- Change tracking
CREATE INDEX idx_oohead_lmdt ON oohead(LMDT);

-- Business chain
CREATE INDEX idx_oohead_chain ON oohead(CHL1, CHL2, CHL3);
```

---

## Integration Points

### Order Entry Flow
```
Customer Request
    ↓
OOHEAD Created (status 15)
    ↓
OOLINE Added
    ↓
Order Released (status 20)
    ↓
ORST/ORSL updated as lines progress
```

### Relationship to Deliveries
```
OOHEAD (Order Header)
    ↓ One order, multiple deliveries
ODHEAD (Delivery Header)
    ↓
ODLINE (Delivery Lines)
```

Each delivery references back to OOHEAD via ORNO.
