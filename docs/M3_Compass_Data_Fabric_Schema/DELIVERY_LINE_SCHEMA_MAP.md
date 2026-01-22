# Customer Order Delivery Line (ODLINE) - Schema Map

## Table Overview
**M3 Table**: ODLINE
**Description**: Delivery Customer Order Line - contains line-level details for each delivery
**Record Count**: 102 fields
**Alternative Name**: "1/(UB)" in M3 documentation

---

## Core Identifiers

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| CONO | integer | Company Number | Company identifier (max: 999) |
| DIVI | string | Division | Division code within company |
| ORNO | string | Order Number | Customer order number |
| PONR | integer | Line Number | Order line number (max: 99999) |
| POSX | integer | Line Suffix | Line suffix (max: 999) |
| DLIX | integer | Delivery Number | Delivery identifier (max: 99999999999) |

**Primary Key**: CONO + DLIX + ORNO + PONR + POSX

---

## Facility and Warehouse

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| FACI | string | Facility | Facility where goods were picked |
| WHLO | string | Warehouse | Warehouse for this delivery line |

---

## Line Type

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| LTYP | string | Line Type | Type of order line |

**Common Values:**
- Empty or "1" = Regular inventory item
- "2" = Text line
- "3" = Charge line
- "4" = Package line

---

## Item Information

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| ITNO | string | Item Number | Product/item delivered |

---

## Quantities (Basic U/M)

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| DLQT | number | Delivered Quantity | Quantity delivered in basic U/M |
| IVQT | number | Invoiced Quantity | Quantity invoiced in basic U/M |
| CHQT | number | Quantity Difference | Difference between delivered and invoiced |
| RTQT | number | Returned Quantity | Quantity returned by customer |

---

## Quantities (Alternate U/M)

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| DLQA | number | Delivered Qty (Alt U/M) | Delivered in alternate unit |
| IVQA | number | Invoiced Qty (Alt U/M) | Invoiced in alternate unit |
| RTQA | number | Returned Qty (Alt U/M) | Returned in alternate unit |

---

## Quantities (Sales Price U/M)

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| DLQS | number | Delivered Qty (Price U/M) | Delivered in sales price unit |
| IVQS | number | Invoiced Qty (Price U/M) | Invoiced in sales price unit |
| CHQS | number | Qty Difference (Price U/M) | Difference in price unit |

---

## Unit of Measure

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| ALUN | string | Alternate U/M | Alternate unit of measure |
| SPUN | string | Sales Price U/M | Unit of measure for pricing |

---

## Pricing

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| SAPR | number | Sales Price | Unit sales price |
| NEPR | number | Net Price | Net price after discounts |
| SACD | integer | Sales Price Quantity | Quantity basis for price (max: 99999) |
| PRMO | string | Price Origin | Source of pricing |
| PRPR | integer | Preliminary Price | Preliminary pricing flag (max: 9) |

---

## Discounts (Status)

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| DIC1 | integer | Discount 1 Status | Status code for discount 1 (max: 9) |
| DIC2 | integer | Discount 2 Status | Status code for discount 2 (max: 9) |
| DIC3 | integer | Discount 3 Status | Status code for discount 3 (max: 9) |
| DIC4 | integer | Discount 4 Status | Status code for discount 4 (max: 9) |
| DIC5 | integer | Discount 5 Status | Status code for discount 5 (max: 9) |
| DIC6 | integer | Discount 6 Status | Status code for discount 6 (max: 9) |
| DIC7 | integer | Discount 7 Status | Status code for discount 7 (max: 9) |
| DIC8 | integer | Discount 8 Status | Status code for discount 8 (max: 9) |

---

## Discounts (Percentage)

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| DIP1 | number | Discount 1 % | Discount percentage 1 (max: 999.99) |
| DIP2 | number | Discount 2 % | Discount percentage 2 (max: 999.99) |
| DIP3 | number | Discount 3 % | Discount percentage 3 (max: 999.99) |
| DIP4 | number | Discount 4 % | Discount percentage 4 (max: 999.99) |
| DIP5 | number | Discount 5 % | Discount percentage 5 (max: 999.99) |
| DIP6 | number | Discount 6 % | Discount percentage 6 (max: 999.99) |
| DIP7 | number | Discount 7 % | Discount percentage 7 (max: 999.99) |
| DIP8 | number | Discount 8 % | Discount percentage 8 (max: 999.99) |

---

## Discounts (Amount)

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| DIA1 | number | Discount 1 Amount | Discount amount 1 |
| DIA2 | number | Discount 2 Amount | Discount amount 2 |
| DIA3 | number | Discount 3 Amount | Discount amount 3 |
| DIA4 | number | Discount 4 Amount | Discount amount 4 |
| DIA5 | number | Discount 5 Amount | Discount amount 5 |
| DIA6 | number | Discount 6 Amount | Discount amount 6 |
| DIA7 | number | Discount 7 Amount | Discount amount 7 |
| DIA8 | number | Discount 8 Amount | Discount amount 8 |

---

## Line Amount

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| LNAM | number | Line Amount | Total line amount in order currency |

---

## Cost Information

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| UCOS | number | Unit Cost | Standard cost per unit |
| UCCD | integer | Cost Code | Standard cost code (max: 9) |
| DCOS | number | Issued Cost Amount | Actual issued cost |
| APBA | integer | Material Price Method | Material pricing method (max: 9) |

---

## Inventory Accounting

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| STCD | integer | Inventory Accounting | Inventory accounting flag (max: 9) |
| PRCH | integer | Price Adjustment Line | Price adjustment indicator (max: 9) |

---

## Campaign and Promotion

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| CMNO | string | Sales Campaign | Sales campaign identifier |
| PIDE | string | Promotion | Promotion identifier |

---

## Supplier Information (Drop Ship)

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| SUNO | string | Supplier Number | Supplier for drop ship |
| ISUN | string | Internal Supplier | Internal supplier number |

---

## Internal Transfer

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| INPR | number | Internal Transfer Price | Transfer price for inter-company |
| CUCT | string | Internal Transfer Currency | Currency for transfer pricing |

---

## Product Configuration

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| HDPR | string | Main Product | Main product in configuration |

---

## Invoice Information

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| IVNO | integer | Invoice Number | Invoice number (max: 999999999) |
| YEA4 | integer | Year | Invoice year (max: 9999) |
| INPX | string | Invoice Prefix | Invoice number prefix |
| EXIN | string | Extended Invoice Number | Extended/alternate invoice number |

---

## Customer Reference

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| CUOR | string | Customer Order Number | Customer's PO number |
| DECU | string | Delivery Customer | Customer receiving delivery |

---

## Text and Documents

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| PRTX | integer | Pre-Text Identity | Text before line |
| POTX | integer | Post-Text Identity | Text after line |
| DTID | integer | Document Identity | Document reference |

---

## Project Management

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| PROJ | string | Project Number | Project identifier |
| ELNO | string | Project Element | Project element/WBS code |

---

## Warranty

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| WATP | string | Warranty Type | Type of warranty |
| GWTP | string | Granted Warranty Type | Warranty type granted |

---

## Supplier Rebate

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| PCLA | number | Supplier Rebate | Supplier rebate amount |
| DPCL | number | Pending Rebate Claim | Pending rebate claim amount |
| SCLB | number | Supplier Rebate Base | Base amount for rebate |
| RAGN | string | Rebate Agreement | Supplier rebate agreement |
| CLAT | string | Rebate Reference Type | Type of rebate reference |
| CLRT | integer | Retrospective Rebate Invoice | Retrospective rebate flag (max: 9) |

---

## Payment Terms

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| TEPY | string | Payment Terms | Payment terms for this line |

---

## Line Classification

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| LNCL | integer | Line Classification | Classification code (max: 9) |

---

## Catch Weight

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| CAWE | number | Catch Weight | Actual weight for catch weight items |
| CICW | number | Corrective Invoice Catch Weight | Catch weight for corrections |

---

## Deposit

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| DAMT | number | Deposit Amount | Deposit amount |
| DVAT | number | Deposit VAT | VAT on deposit |

---

## Migration Status

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| MIGI | string | Internal Migration Status | Data migration status flag |

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
| Core Identifiers | 6 | Line identification |
| Facility/Warehouse | 2 | Location info |
| Line Type | 1 | Type of line |
| Item | 1 | Item reference |
| Quantities (Basic) | 4 | Basic U/M quantities |
| Quantities (Alt) | 3 | Alternate U/M quantities |
| Quantities (Price) | 3 | Price U/M quantities |
| Unit of Measure | 2 | U/M codes |
| Pricing | 5 | Prices and price control |
| Discounts (Status) | 8 | Discount status flags |
| Discounts (%) | 8 | Discount percentages |
| Discounts (Amount) | 8 | Discount amounts |
| Line Amount | 1 | Total line value |
| Cost | 4 | Cost information |
| Inventory | 2 | Accounting control |
| Campaign/Promo | 2 | Marketing programs |
| Supplier | 2 | Drop ship suppliers |
| Internal Transfer | 2 | Inter-company pricing |
| Configuration | 1 | Product configuration |
| Invoice | 4 | Invoice references |
| Customer | 2 | Customer references |
| Text/Documents | 3 | Text references |
| Project | 2 | Project references |
| Warranty | 2 | Warranty information |
| Supplier Rebate | 6 | Rebate tracking |
| Payment | 1 | Payment terms |
| Classification | 1 | Line classification |
| Catch Weight | 2 | Weight tracking |
| Deposit | 2 | Deposit handling |
| Migration | 1 | Migration status |
| Audit/Metadata | 11 | Change tracking + Data Lake |
| **Total** | **102** | All fields |

---

## Key Fields for Data Modeling

### Primary Key
- CONO + DLIX + ORNO + PONR + POSX

### Foreign Keys / Relationships
- **To Delivery Header**: DLIX (links to ODHEAD)
- **To Order Line**: ORNO + PONR + POSX (links to OOLINE)
- **To Item**: ITNO
- **To Invoice**: YEA4 + IVNO
- **To Facility**: FACI
- **To Warehouse**: WHLO
- **To Supplier**: SUNO

### Critical for Incremental Load
- LMDT (Change Date)
- LMTS (Timestamp)
- deleted (must use string comparison: 'false')

### Most Commonly Used Fields
1. DLIX, ORNO, PONR, POSX (identification)
2. ITNO (item)
3. DLQT, IVQT (quantities)
4. SAPR, NEPR, LNAM (pricing)
5. IVNO, YEA4 (invoice reference)
6. LMDT (change tracking)

---

## Common Query Patterns

### Get Delivery Lines with Header
```sql
SELECT
  dh.DLIX,
  dh.ORNO,
  dh.DLDT,
  dl.PONR,
  dl.POSX,
  dl.ITNO,
  dl.DLQT,
  dl.LNAM
FROM ODHEAD dh
JOIN ODLINE dl
  ON dl.DLIX = dh.DLIX
  AND dl.deleted = 'false'
WHERE dh.deleted = 'false'
  AND dh.DLIX = 123456789
ORDER BY dl.PONR, dl.POSX
```

### Compare Delivered vs Invoiced Quantities
```sql
SELECT
  DLIX,
  ORNO,
  PONR,
  ITNO,
  DLQT as delivered,
  IVQT as invoiced,
  CHQT as difference,
  CASE
    WHEN CHQT != 0 THEN 'Variance'
    ELSE 'Match'
  END as status
FROM ODLINE
WHERE deleted = 'false'
  AND DLIX = 123456789
ORDER BY PONR, POSX
```

### Delivery Line Summary by Item
```sql
SELECT
  ITNO,
  COUNT(*) as line_count,
  SUM(DLQT) as total_delivered,
  SUM(IVQT) as total_invoiced,
  SUM(LNAM) as total_value
FROM ODLINE
WHERE deleted = 'false'
  AND DLIX >= 100000000
  AND DLIX <= 200000000
GROUP BY ITNO
ORDER BY total_value DESC
```

### Lines with Discounts
```sql
SELECT
  DLIX,
  ORNO,
  PONR,
  ITNO,
  SAPR,
  NEPR,
  DIP1, DIP2, DIP3,
  LNAM
FROM ODLINE
WHERE deleted = 'false'
  AND (DIP1 > 0 OR DIP2 > 0 OR DIP3 > 0)
ORDER BY DLIX, PONR
```

---

## Relationship to Order Lines

```sql
-- Compare order line to delivery line
SELECT
  ol.ORNO,
  ol.PONR,
  ol.ITNO,
  ol.ORQT as ordered,
  ol.DLQT as delivered_on_order,
  dl.DLQT as delivered_this_delivery,
  dl.DLIX,
  ol.ORQT - ol.DLQT as remaining_to_deliver
FROM OOLINE ol
LEFT JOIN ODLINE dl
  ON dl.ORNO = ol.ORNO
  AND dl.PONR = ol.PONR
  AND dl.POSX = ol.POSX
  AND dl.deleted = 'false'
WHERE ol.deleted = 'false'
  AND ol.ORNO = 'CO123456'
ORDER BY ol.PONR, dl.DLIX
```

---

## Multi-Delivery Tracking

A single order line can be delivered across multiple deliveries:

```sql
-- All deliveries for an order line
SELECT
  DLIX,
  DLQT,
  IVQT,
  IVNO,
  YEA4,
  LNAM
FROM ODLINE
WHERE deleted = 'false'
  AND ORNO = 'CO123456'
  AND PONR = 1
  AND POSX = 0
ORDER BY DLIX
```

---

## Quantity Reconciliation

Understanding the quantity fields:

- **DLQT**: Physical quantity delivered
- **IVQT**: Quantity on invoice (usually = DLQT)
- **CHQT**: Difference (DLQT - IVQT)
- **RTQT**: Quantity returned after delivery

**Normal Flow**: DLQT = IVQT, CHQT = 0
**Under Invoice**: DLQT > IVQT, CHQT > 0
**Over Invoice**: DLQT < IVQT, CHQT < 0

---

## Discount Cascade

Discounts apply in sequence:

```
Sales Price (SAPR)
  - DIP1 (Discount 1)
  - DIP2 (Discount 2)
  - DIP3 (Discount 3)
  ...
  = Net Price (NEPR)
  × Quantity (DLQT)
  = Line Amount (LNAM)
```

---

## Cost vs Price

- **UCOS**: Standard cost (what it costs us)
- **SAPR**: Sales price (what we charge)
- **NEPR**: Net price after discounts
- **DCOS**: Actual issued cost from inventory
- **LNAM**: Total line revenue

**Margin** = LNAM - (DCOS × DLQT)

---

## Catch Weight Items

For variable weight items (meat, produce):

```sql
SELECT
  ITNO,
  DLQT as units_delivered,
  CAWE as actual_weight,
  SAPR as price_per_unit,
  LNAM as extended_amount
FROM ODLINE
WHERE deleted = 'false'
  AND CAWE IS NOT NULL
  AND CAWE > 0
```

**Use Case**: Ordered 10 kg, actual delivered weight 10.3 kg → invoice based on actual.

---

## Supplier Rebate Tracking

For lines with supplier rebates:

```sql
SELECT
  ITNO,
  SUNO,
  DLQT,
  LNAM as sales_value,
  SCLB as rebate_base,
  PCLA as rebate_amount,
  DPCL as pending_claim,
  RAGN as agreement
FROM ODLINE
WHERE deleted = 'false'
  AND PCLA > 0
ORDER BY SUNO, ITNO
```

---

## Drop Ship Lines

Identify drop-ship lines:

```sql
SELECT
  ORNO,
  PONR,
  ITNO,
  SUNO as supplier,
  DLQT,
  DECU as ship_to_customer
FROM ODLINE
WHERE deleted = 'false'
  AND SUNO IS NOT NULL
  AND SUNO != ''
ORDER BY SUNO
```

---

## Internal Transfer Pricing

For inter-company transfers:

```sql
SELECT
  ITNO,
  DLQT,
  SAPR as external_price,
  INPR as transfer_price,
  CUCT as transfer_currency,
  (SAPR - INPR) * DLQT as markup_amount
FROM ODLINE
WHERE deleted = 'false'
  AND INPR IS NOT NULL
  AND INPR > 0
```

---

## Price Adjustment Lines

Lines marked as price adjustments:

```sql
SELECT
  DLIX,
  ORNO,
  PONR,
  ITNO,
  LNAM as adjustment_amount
FROM ODLINE
WHERE deleted = 'false'
  AND PRCH = 1  -- Price adjustment line
```

**Note**: These don't affect inventory, only invoicing.

---

## Warranty Lines

Lines with warranty:

```sql
SELECT
  ITNO,
  WATP as warranty_offered,
  GWTP as warranty_granted,
  DLQT,
  LNAM
FROM ODLINE
WHERE deleted = 'false'
  AND (WATP IS NOT NULL OR GWTP IS NOT NULL)
```

---

## Preliminary Pricing

Lines with preliminary prices that may be updated:

```sql
SELECT
  DLIX,
  ORNO,
  PONR,
  ITNO,
  SAPR,
  PRPR as is_preliminary
FROM ODLINE
WHERE deleted = 'false'
  AND PRPR = 1
```

---

## Return Lines

Identify return/credit lines:

```sql
SELECT
  DLIX,
  ORNO,
  PONR,
  ITNO,
  RTQT as returned_quantity,
  LNAM as credit_amount
FROM ODLINE
WHERE deleted = 'false'
  AND RTQT > 0
ORDER BY DLIX, PONR
```

---

## Incremental Load Strategy

```sql
SELECT *
FROM ODLINE
WHERE deleted = 'false'
  AND LMDT >= 20260101  -- Change date filter
ORDER BY LMDT, LMTS
```

---

## Performance Considerations

### Recommended Indexes

```sql
-- Primary key
CREATE INDEX idx_odline_pk ON odline(CONO, DLIX, ORNO, PONR, POSX);

-- Delivery lookup
CREATE INDEX idx_odline_dlix ON odline(DLIX);

-- Order line lookup
CREATE INDEX idx_odline_order ON odline(ORNO, PONR, POSX);

-- Item lookup
CREATE INDEX idx_odline_itno ON odline(ITNO);

-- Invoice lookup
CREATE INDEX idx_odline_invoice ON odline(YEA4, IVNO);

-- Change tracking
CREATE INDEX idx_odline_lmdt ON odline(LMDT);
```

---

## Data Quality Checks

### Quantity Validation
```sql
-- Find lines with quantity mismatches
SELECT *
FROM ODLINE
WHERE deleted = 'false'
  AND CHQT != 0
  AND IVNO > 0  -- Already invoiced
```

### Cost Validation
```sql
-- Lines with zero cost (potential issue)
SELECT *
FROM ODLINE
WHERE deleted = 'false'
  AND STCD = 1  -- Inventory item
  AND (UCOS = 0 OR UCOS IS NULL)
```

### Discount Validation
```sql
-- Lines with excessive discounts
SELECT *
FROM ODLINE
WHERE deleted = 'false'
  AND (DIP1 + DIP2 + DIP3 + DIP4 + DIP5 + DIP6) > 50  -- >50% total discount
```

---

## Best Practices

1. **Always join with ODHEAD**: Get delivery-level context
2. **Check quantity fields carefully**: DLQT vs IVQT vs CHQT
3. **Handle multi-delivery**: Same order line can be on multiple deliveries
4. **Validate discounts**: Check both DIC status and DIP percentage
5. **Cost tracking**: Use DCOS for actual cost, UCOS for standard
6. **Catch weight**: Check CAWE for variable weight items
7. **Returns**: Track RTQT separately from deliveries
8. **Rebates**: Monitor PCLA and DPCL for supplier rebate claims

---

## Integration Flow

```
Order Line (OOLINE)
    ↓
Pick & Pack
    ↓
Delivery Created (ODHEAD)
    ↓
Delivery Line (ODLINE)
    ↓ DLQT populated
Goods Issued
    ↓ DCOS populated
Invoice Created
    ↓ IVNO/YEA4 populated, IVQT = DLQT
Invoice Posted
    ↓
Customer Payment
```

---

## Full Delivery Analysis Query

```sql
SELECT
  dh.DLIX,
  dh.ORNO,
  dh.DLDT,
  dh.CONN as shipment,
  dl.PONR,
  dl.ITNO,
  dl.DLQT,
  dl.SAPR,
  dl.NEPR,
  dl.LNAM,
  dl.UCOS,
  dl.DCOS,
  dl.LNAM - (dl.DCOS * dl.DLQT) as line_margin,
  dl.IVNO,
  dl.YEA4,
  CASE
    WHEN dl.IVNO > 0 THEN 'Invoiced'
    WHEN dh.ORST = '44' THEN 'Delivered'
    WHEN dh.ORST = '33' THEN 'Picked'
    ELSE 'Other'
  END as status
FROM ODHEAD dh
JOIN ODLINE dl
  ON dl.DLIX = dh.DLIX
  AND dl.deleted = 'false'
WHERE dh.deleted = 'false'
  AND dh.DLDT >= 20260101
ORDER BY dh.DLDT, dh.DLIX, dl.PONR
```
