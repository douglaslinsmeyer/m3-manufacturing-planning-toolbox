# Customer Order Line (OOLINE) - Schema Map

## Table Overview
**M3 Table**: OOLINE
**Description**: Customer Order Line file - contains all line-level details for customer orders
**Record Count**: 303 fields

---

## Core Identifiers

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| CONO | integer | Company Number | Company identifier (max: 999) |
| DIVI | string | Division | Division code within company |
| ORNO | string | Order Number | Customer order number (unique identifier) |
| PONR | integer | Line Number | Order line number (max: 99999) |
| POSX | integer | Line Suffix | Line suffix for sub-lines (max: 999) |
| LTYP | string | Line Type | Type of order line (item, text, charge, etc.) |

---

## Item Information

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| ITNO | string | Item Number | Product/item identifier |
| REPI | string | Replacement Item | Item number that replaced this item |
| ITDS | string | Item Description | Product name/description |
| TEDS | string | Extended Description | Additional description line 1 |
| POPN | string | Alias Number | Alternative item identifier |
| ALWT | integer | Alias Category | Alias type category (max: 99) |
| ALWQ | string | Alias Qualifier | Qualifier for alias |

---

## Quantities (Basic U/M)

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| ORQT | number | Ordered Quantity | Original ordered quantity in basic unit |
| RNQT | number | Remaining Quantity | Quantity not yet delivered |
| ALQT | number | Allocated Quantity | Quantity reserved from inventory |
| PLQT | number | Picking List Quantity | Quantity on picking lists |
| DLQT | number | Delivered Quantity | Total quantity delivered to customer |
| IVQT | number | Invoiced Quantity | Total quantity invoiced |
| RTQT | number | Returned Quantity | Quantity returned by customer |
| CNQT | number | Confirmed Quantity | Quantity confirmed by customer |
| CAWE | number | Catch Weight | Actual weight for catch weight items |

---

## Quantities (Alternate U/M)

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| ORQA | number | Ordered Qty (Alt U/M) | Ordered quantity in alternate unit |
| RNQA | number | Remaining Qty (Alt U/M) | Remaining in alternate unit |
| ALQA | number | Allocated Qty (Alt U/M) | Allocated in alternate unit |
| PLQA | number | Picking List Qty (Alt U/M) | Picking list in alternate unit |
| DLQA | number | Delivered Qty (Alt U/M) | Delivered in alternate unit |
| IVQA | number | Invoiced Qty (Alt U/M) | Invoiced in alternate unit |
| RTQA | number | Returned Qty (Alt U/M) | Returned in alternate unit |

---

## Unit of Measure

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| DCCD | integer | Decimal Places (Basic) | Number of decimals for basic U/M (max: 9) |
| ALUN | string | Alternate Unit | Alternate unit of measure code |
| DCCA | integer | Decimal Places (Alt) | Decimals for alternate U/M |
| COFA | number | Conversion Factor | Factor to convert basic to alternate |
| DMCF | integer | Conversion Method | Method for unit conversion (max: 9) |
| SPUN | string | Sales Price Unit | Unit of measure for pricing |
| PCOF | number | Price Adjustment Factor | Factor for price unit conversion |
| DCCS | integer | Decimal Places (Price) | Decimals for sales price U/M |
| COFS | number | Conversion Factor (Price) | Sales price unit conversion |
| DMCS | integer | Conversion Method (Price) | Price unit conversion method |
| MNIN | string | Main U/M Indicator | Indicates if basic or alt is primary |

---

## Pricing

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| SAPR | number | Sales Price | Unit sales price |
| NEPR | number | Net Price | Net price after discounts |
| SACD | integer | Sales Price Quantity | Quantity basis for price (max: 99999) |
| PRMO | string | Price Origin | Source of pricing (manual, list, agreement) |
| LNAM | number | Line Amount | Total line amount in order currency |
| LNA2 | number | Line Amount 2 | Alternate line amount calculation |
| EDFP | number | EDIFACT Price | Price from EDI transmission |
| INPR | number | Internal Transfer Price | Transfer price for inter-company |
| CUCT | string | Internal Transfer Currency | Currency for transfer pricing |
| NTCD | integer | Net Price Used | Flag if net pricing active (max: 9) |
| PRPR | integer | Preliminary Price | Flag for preliminary pricing (max: 9) |

---

## Discounts (Percentage)

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| DIP1 | number | Discount 1 % | Discount percentage 1 (max: 999.99) |
| DIP2 | number | Discount 2 % | Discount percentage 2 |
| DIP3 | number | Discount 3 % | Discount percentage 3 |
| DIP4 | number | Discount 4 % | Discount percentage 4 |
| DIP5 | number | Discount 5 % | Discount percentage 5 |
| DIP6 | number | Discount 6 % | Discount percentage 6 |
| DIP7 | number | Discount 7 % | Discount percentage 7 |
| DIP8 | number | Discount 8 % | Discount percentage 8 |

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

## Discount Configuration

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| DIC1 | integer | Discount 1 Status | Status code for discount 1 (max: 9) |
| DIC2 | integer | Discount 2 Status | Status code for discount 2 |
| DIC3 | integer | Discount 3 Status | Status code for discount 3 |
| DIC4 | integer | Discount 4 Status | Status code for discount 4 |
| DIC5 | integer | Discount 5 Status | Status code for discount 5 |
| DIC6 | integer | Discount 6 Status | Status code for discount 6 |
| DIC7 | integer | Discount 7 Status | Status code for discount 7 |
| DIC8 | integer | Discount 8 Status | Status code for discount 8 |
| CMP1 | string | Discount 1 Statistics ID | Statistical identity for discount 1 |
| CMP2 | string | Discount 2 Statistics ID | Statistical identity for discount 2 |
| CMP3 | string | Discount 3 Statistics ID | Statistical identity for discount 3 |
| CMP4 | string | Discount 4 Statistics ID | Statistical identity for discount 4 |
| CMP5 | string | Discount 5 Statistics ID | Statistical identity for discount 5 |
| CMP6 | string | Discount 6 Statistics ID | Statistical identity for discount 6 |
| CMP7 | string | Discount 7 Statistics ID | Statistical identity for discount 7 |
| CMP8 | string | Discount 8 Statistics ID | Statistical identity for discount 8 |
| DIBE | string | Discount Base | Base for discount calculation |
| DIRE | string | Discount Relation | Discount relationship type |
| DCHA | integer | Manually Changeable Discount | Can discount be changed manually (max: 9) |
| DPST | integer | Discount Statistics Field | Stats field for discount (max: 9) |
| DDSU | integer | Discount Presentation | How discount is displayed (max: 9) |
| IDSC | integer | Internal Discount | Internal discount flag (max: 9) |
| OTDI | integer | Order Total Discount Gen | Generates order total discount (max: 9) |

---

## Dates and Times (Requested)

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| DWDT | integer | Requested Delivery Date | Customer requested delivery date (YYYYMMDD) |
| DWHM | integer | Requested Delivery Time | Customer requested time (HHMM) |
| DWDZ | integer | Requested Delivery Date (TZ) | Requested date with timezone |
| DWHZ | integer | Requested Delivery Time (TZ) | Requested time with timezone |
| TIZO | string | Time Zone | Time zone code |

---

## Dates and Times (Confirmed)

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| CODT | integer | Confirmed Delivery Date | Promised delivery date to customer |
| COHM | integer | Confirmed Delivery Time | Promised delivery time |
| CODZ | integer | Confirmed Delivery Date (TZ) | Confirmed date with timezone |
| COHZ | integer | Confirmed Delivery Time (TZ) | Confirmed time with timezone |
| FDED | integer | First Delivery Date | Date of first partial delivery |
| LDED | integer | Last Delivery Date | Date of last delivery |

---

## Planning Dates

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| PLDT | integer | Planning Date | MRP planning date (YYYYMMDD) |
| PLHM | integer | Planning Time | MRP planning time (HHMM) |
| DLTS | integer | Delivery Time Slot | Assigned delivery time slot |

---

## Status and Control

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| ORST | string | Order Line Status | Highest status of order line |
| STCD | integer | Inventory Accounting | Inventory accounting flag (max: 9) |
| PRCH | integer | Price Adjustment Line | Is this a price adjustment (max: 9) |
| REND | integer | Manual Completion | Manual completion flag (max: 9) |
| SMCC | integer | Contribution Margin Check | CM check status (max: 9) |
| BNCD | integer | Bonus Generating | Generates bonus commission (max: 9) |
| PRAC | integer | Commission Generating | Generates sales commission (max: 9) |
| DELS | integer | Delivery Schedule Order | Scheduled delivery flag (max: 9) |
| ABNO | integer | Abnormal Demand | Abnormal demand flag (max: 9) |
| BPST | integer | Buying Pattern Status | Status of buying pattern (max: 9) |
| CINA | integer | Create Cost Accounting | Create cost accounting trans (max: 9) |
| OLSC | integer | Order Line Stop Code | Line stop/hold code (max: 9) |
| PIKD | integer | Picked | Line has been picked (max: 9) |
| UPAV | integer | Update Material Plan | Update MRP flag (max: 9) |
| LNCL | integer | Line Classification | Classification code (max: 9) |

---

## Location and Logistics

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| FACI | string | Facility | Manufacturing/supply facility |
| WHLO | string | Warehouse | Warehouse code |
| WHSL | string | Storage Location | Specific location in warehouse |
| BANO | string | Lot Number | Lot/batch number for traceability |
| SERN | string | Serial Number | Serial number for serialized items |
| CAMU | string | Container | Container identifier |
| ADID | string | Address Number | Delivery address identifier |
| AETY | integer | Address Type | Type of address (max: 9) |

---

## Routing and Delivery

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| ROUT | string | Route | Delivery route code |
| RODN | integer | Route Departure | Route departure number (max: 999) |
| DSDT | integer | Departure Date | Scheduled departure date |
| DSHM | integer | Departure Time | Scheduled departure time |
| DSTX | string | Description | Delivery description/instructions |
| MODL | string | Delivery Method | Method of delivery (truck, ship, etc.) |
| TEDL | string | Delivery Terms | Incoterms for delivery |
| TEL2 | string | Delivery Terms Text | Additional delivery terms text |

---

## Splitting and Scheduling

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| JDCD | string | Joint Delivery Code | Joint delivery grouping code |
| DLSP | string | Delivery Split Rule | How to split deliveries |
| DLBU | integer | Shipping Period | Delivery bucket/period (max: 9) |
| SPOS | integer | Sequence Position | Line sequence number (max: 999) |
| SPLC | string | Delivery Split Rule | Delivery split logic code |
| SRCD | integer | Reservation Level | Inventory reservation level (max: 9) |
| DEFC | integer | Demand Factor | Demand multiplier factor (max: 999) |

---

## Reference Orders (Critical for Linking)

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| RORC | integer | Reference Order Category | Type of reference order (max: 9) |
| RORN | string | Reference Order Number | Order number being referenced |
| RORL | integer | Reference Order Line | Line number in reference order (max: 999999) |
| RORX | integer | Reference Line Suffix | Suffix of reference line (max: 999) |

**RORC Values:**
- 1 = Purchase Order
- 2 = Manufacturing Order
- 3 = Customer Order
- 4 = Distribution Order
- 5 = MRP Order Proposal

---

## Cost Information

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| UCOS | number | Unit Cost Price | Standard cost per unit |
| COCD | integer | Cost Quantity | Quantity basis for cost (max: 99999) |
| UCCD | integer | Cost Code | Standard cost code (max: 9) |
| SCMO | string | Costing Model | Sales price costing model |
| ACRF | string | Accounting Control Object | User-defined accounting object |

---

## Customer Information

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| CUNO | string | Customer Number | Customer placing the order |
| DECU | string | Delivery Customer | Customer receiving delivery |
| CUOR | string | Customer Order Number | Customer's PO number |
| CUPO | integer | Customer Line Number | Line on customer's PO (max: 9999999) |
| CUSX | integer | Customer Line Suffix | Suffix on customer line (max: 999) |

---

## Campaign and Agreements

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| CMNO | string | Sales Campaign | Sales campaign identifier |
| AGNO | string | Blanket Agreement Number | Blanket order agreement |
| BAGC | string | Agreement Customer | Customer on blanket agreement |
| BAGD | integer | Agreement Start Date | Blanket agreement start (YYYYMMDD) |
| AGLN | integer | Agreement Sequence | Sequence in agreement (max: 9999999) |
| PRRF | string | Price List | Price list used |
| OFNO | string | Quotation Number | Quote converted to order |
| VERS | integer | Quotation Version | Version of quotation (max: 99) |
| OFNR | integer | Quotation Line | Line from quotation (max: 99999) |
| OFSX | integer | Quotation Line Suffix | Quotation line suffix (max: 999) |

---

## Product Configuration

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| HDPR | string | Main Product | Main product in configuration |
| PRHL | string | Product Highest Level | Top-level product in structure |
| CFIN | integer | Configuration Number | Unique config instance (max: 9999999999) |
| ECVS | integer | Simulation Round | Configuration simulation (max: 999) |
| CFGL | string | Configuration Position | Position in configuration |

---

## Attributes (Built-in Numeric)

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| ATV1 | number | Attribute Value 1 | Numeric attribute field 1 |
| ATV2 | number | Attribute Value 2 | Numeric attribute field 2 |
| ATV3 | number | Attribute Value 3 | Numeric attribute field 3 |
| ATV4 | number | Attribute Value 4 | Numeric attribute field 4 |
| ATV5 | number | Attribute Value 5 | Numeric attribute field 5 |

---

## Attributes (Built-in String)

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| ATV6 | string | Attribute Value 6 | Text attribute field 6 |
| ATV7 | string | Attribute Value 7 | Text attribute field 7 |
| ATV8 | string | Attribute Value 8 | Text attribute field 8 |
| ATV9 | string | Attribute Value 9 | Text attribute field 9 |
| ATV0 | string | Attribute Value 10 | Text attribute field 10 |

---

## Attribute Configuration

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| ATNR | integer | Attribute Number | Unique attribute identifier |
| ATPR | string | Attribute Pricing Rule | Pricing rule for attributes |
| ATMO | string | Attribute Model | Attribute model reference |
| CNNR | integer | Detailed Order Requirement | DOR number for attributes |

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

## Sales and Statistics

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| SMCD | string | Salesperson | Sales representative code |
| VRCD | string | Business Type (TST) | Trade statistics business type |
| ECLC | string | Labor Code (TST) | Trade statistics labor code |
| ORCO | string | Country of Origin | Manufacturing country of origin |

---

## Packaging

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| TEPA | string | Packaging Terms | Packaging terms code |
| PACT | string | Packaging Type | Type of packaging |
| CUPA | string | Customer Packaging ID | Customer's package identifier |
| D1QT | number | Standard Quantity | Standard packaging quantity |
| QTQL | integer | Demand Bucket | Demand planning bucket (max: 9) |

---

## EDI and External References

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| E0PA | string | EDI Partner | EDI partner code |
| DSGP | string | Delivery Schedule Group | Schedule grouping |
| MOYE | string | Model/Year | Product model and year |
| PUSN | string | Delivery Note Reference | Packing slip reference |
| PUTP | integer | Delivery Note Qualifier | Reference qualifier (max: 9) |

---

## Project Management

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| PROJ | string | Project Number | Project identifier |
| ELNO | string | Project Element | Project element/activity |

---

## Purchasing (for Drop Ship)

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| SUNO | string | Supplier Number | Supplier for drop ship |
| PUPR | number | Purchase Price | Purchase price from supplier |
| PUCD | integer | Purchase Price Quantity | Quantity basis for purchase price |
| PPUN | string | Purchase Price U/M | Unit of measure for purchase price |
| CUCD | string | Currency | Currency code |
| ODI1 | number | Order Discount 1 | Purchase discount 1 |
| ODI2 | number | Order Discount 2 | Purchase discount 2 |
| ODI3 | number | Order Discount 3 | Purchase discount 3 |
| OURR | string | Our Reference | Our reference number |
| OURT | string | Reference Type | Type of reference |
| BUYE | string | Buyer | Buyer/planner code |
| FWHL | string | From Warehouse | Source warehouse for transfer |

---

## Text and Documents

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| PRTX | integer | Pre-Text Identity | Text before line (ID) |
| POTX | integer | Post-Text Identity | Text after line (ID) |
| DTID | integer | Document Identity | Document reference (ID) |

---

## Discount in Kind

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| DINR | integer | Discount Line Number | Line giving discount in kind |
| DISX | integer | Discount Line Suffix | Suffix of discount line |

---

## Action Reasons

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| RSCD | string | Transaction Reason | Reason for transaction |
| ARST | integer | Action Reason Type | Type of action reason (max: 99) |
| RSC1 | string | Action Reason | Action reason code |
| RSC5 | string | Transaction Reason Price | Reason for price change |
| RSC6 | string | Transaction Reason Quantity | Reason for quantity change |
| RSC7 | string | Transaction Reason Time | Reason for time change |
| RSC8 | string | Transaction Reason Prod Restriction | Reason for product restriction |
| PRHC | integer | Product Restriction Stop | Product restriction stop code (max: 9) |

---

## Payment Terms

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| TEPY | string | Payment Terms | Payment terms code |
| PMOR | integer | Payment Terms Origin | Source of payment terms (max: 9) |

---

## Tax and VAT

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| VTCD | integer | VAT Code | Value added tax code (max: 99) |
| TECN | string | Tax Exemption Contract | Tax exemption contract number |

---

## Customs

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| CPRE | string | Customs Procedure Export | Export customs procedure |
| HAFE | string | Harbor or Airport | Export port/airport code |

---

## Buying Pattern

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| BPAT | string | Buying Pattern | Buying pattern code |
| BPST | integer | Buying Pattern Status | Status of pattern (max: 9) |
| NOAA | integer | Override Allocation Method | Allocation override (max: 9) |

---

## Warranty (Service Orders)

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| WATP | string | Warranty Type | Type of warranty |
| GWTP | string | Granted Warranty Type | Warranty type granted |
| PRHW | string | Product Number (Warranty) | Product under warranty |
| SERW | string | Serial Number (Warranty) | Serial of warranted product |
| PWNR | integer | Warranty Order Line | Parent warranty line (max: 99999) |
| PWSX | integer | Warranty Line Suffix | Parent warranty suffix (max: 999) |
| EWST | integer | Extended Warranty Start | Extended warranty start flag (max: 9) |
| AGNB | string | Warranty Agreement | Warranty agreement number |
| CTNS | string | Service Agreement | Service agreement number |
| WATQ | integer | Goodwill Qualified | Qualifies for goodwill (max: 9) |

---

## Promotions and Rebates

| M3 Field | data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| PIDE | string | Promotion | Promotion identifier |
| INIP | integer | Initial Promotion | Initial promotion flag (max: 9) |
| MPID | integer | Manual Promotion | Manual promotion flag (max: 9) |
| PCLA | number | Supplier Rebate | Supplier rebate amount |
| SCLB | number | Supplier Rebate Base | Base amount for rebate calculation |
| RAGN | string | Rebate Agreement | Supplier rebate agreement |
| CLAT | string | Rebate Reference Type | Type of rebate reference |

---

## Supply Model

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| CBSM | integer | Created by Supply Model | Supply model created flag (max: 9) |
| PRNO | string | Product Number | Manufactured product number |
| SCHN | integer | Schedule Number | Supply chain schedule |
| REPN | integer | Receiving Number | Receiving point number |
| ENNO | string | Entitlement Number | Entitlement reference |

---

## Order Closure

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| QTBC | number | Quantity to be Closed | Qty for closure |
| BOP1 | integer | Automatic Closing | Auto-close flag (max: 9) |
| PRIO | integer | Priority | Line priority (max: 9) |

---

## Product Charge Lines

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| PCNR | integer | Parent Charge Line | Parent line for charges (max: 99999) |
| PCSU | integer | Parent Charge Suffix | Parent charge suffix (max: 999) |
| RELI | integer | Related Line | Related order line (max: 99999) |

---

## Core Charge

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| CCSA | number | Core Charge Sales Price | Sales price for core charge |

---

## Industry Application

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| INAP | string | Industry Application | Industry-specific application code |

---

## Rail Transport

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| RASN | string | Rail Station | Railway station code |

---

## Dispatch Policy

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| DPOL | string | Dispatch Policy | Warehouse dispatch policy |

---

## Manual Price Date

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| MPRD | integer | Manual Price Date | Date for manual pricing (YYYYMMDD) |

---

## Order Type

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| ORTY | string | Order Type | Customer order type code |

---

## Allocation

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| APBA | integer | Material Price Method | Material pricing method (max: 9) |
| TOMU | number | Issue Multiple | Issue quantity multiple |

---

## Migration

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| MIGI | string | Internal Migration Status | Data migration status |

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
| Core Identifiers | 6 | Basic record identification |
| Quantities | 17 | All quantity fields (basic + alternate) |
| Pricing | 11 | Prices and price-related |
| Discounts | 32 | All discount fields (%, amount, config) |
| Dates/Times | 14 | Delivery dates and times |
| Status/Control | 15 | Status flags and controls |
| Attributes | 21 | Built-in + user-defined attributes |
| Reference Orders | 4 | **Critical for CO→MO/MOP linking** |
| Customer Info | 5 | Customer references |
| Location/Logistics | 9 | Warehouse, location, routing |
| Audit/Metadata | 11 | Change tracking + Data Lake |
| **Total** | **303** | All fields |

---

## Key Fields for Data Modeling

### Primary Key
- CONO + ORNO + PONR + POSX

### Foreign Keys / Relationships
- **To Manufacturing Orders**: RORC=2, RORN=MFNO, RORL→MO line reference
- **To MO Proposals**: RORC=5, RORN=PLPN reference
- **To Customer**: CUNO
- **To Item**: ITNO
- **To Warehouse**: WHLO

### Critical for Incremental Load
- LMDT (Change Date)
- LMTS (Timestamp)
- deleted (must use string comparison: 'false')

### Most Commonly Used Fields
1. ORNO, PONR, POSX (identification)
2. ITNO, ORQT (item and quantity)
3. ORST (status)
4. DWDT, CODT (delivery dates)
5. RORC, RORN, RORL (supply chain links)
6. LMDT (change tracking)
