// Authentication types
export interface UserContext {
  company?: string;
  division?: string;
  facility?: string;
  warehouse?: string;
}

export interface AuthStatus {
  authenticated: boolean;
  environment?: 'TRN' | 'PRD';
  userContext?: UserContext;
  userProfile?: UserProfile;
}

// User Profile types
export interface UserProfile {
  id: string;
  userName: string;
  displayName: string;
  email?: string;
  title?: string;
  department?: string;
  groups?: UserProfileGroup[];
  m3Info?: M3UserInfo;
}

export interface UserProfileGroup {
  value: string;
  display: string;
  type: string; // "Security Role" | "Accounting Entity" | "Distribution Group"
}

export interface M3UserInfo {
  userId: string;
  fullName: string;
  defaultCompany: string;
  defaultDivision: string;
  defaultFacility: string;
  defaultWarehouse: string;
  languageCode: string;
  dateFormat: string;
  dateSeparator: string;
  timeSeparator: string;
  timeZone: string;
}

// M3 Organizational Hierarchy Types
export interface M3Company {
  companyNumber: string;
  companyName: string;
  currency: string;
}

export interface M3Division {
  companyNumber: string;
  division: string;
  divisionName: string;
  facility?: string;
  warehouse?: string;
}

export interface M3Facility {
  companyNumber: string;
  facility: string;
  facilityName: string;
  division?: string;
  warehouse?: string;
}

export interface M3Warehouse {
  companyNumber: string;
  warehouse: string;
  warehouseName: string;
  division?: string;
  facility?: string;
}

// Context State Types
export interface EffectiveContext {
  company: string;
  division: string;
  facility: string;
  warehouse: string;
  hasTemporaryOverrides: boolean;
  userDefaults: UserContext;
  loadError?: string;
}

export interface TemporaryOverride {
  company?: string;
  division?: string;
  facility?: string;
  warehouse?: string;
}

// Production Order types
export interface ProductionOrder {
  id: number;
  orderNumber: string;
  orderType: 'MO' | 'MOP';
  itemNumber: string;
  itemDescription?: string;
  facility: string;
  warehouse?: string;
  plannedStartDate: string;
  plannedFinishDate: string;
  orderedQuantity: number;
  quantityUnit?: string;
  status: string;
  moId?: number;
  mopId?: number;
}

// Manufacturing Order (full details)
export interface ManufacturingOrder {
  id: number;
  facility: string;
  moNumber: string;
  productNumber: string;
  itemNumber: string;
  itemDescription?: string;
  orderedQuantity: number;
  manufacturedQuantity: number;
  scrappedQuantity: number;
  quantityUnit?: string;
  plannedStartDate?: string;
  plannedFinishDate?: string;
  actualStartDate?: string;
  actualFinishDate?: string;
  status: string;
  orderType?: string;
  priority?: string;
  warehouse?: string;
  responsiblePerson?: string;
  planner?: string;
  referenceOrderCategory?: string;
  referenceOrderNumber?: string;
  referenceOrderLine?: string;
  operations?: MOOperation[];
  materials?: MOMaterial[];
}

export interface MOOperation {
  id: number;
  operationNumber: string;
  workCenter?: string;
  operationDescription?: string;
  setupTime?: number;
  runTimePerUnit?: number;
  totalRunTime?: number;
  completedQuantity: number;
  scrappedQuantity: number;
  status?: string;
  plannedStartDate?: string;
  plannedFinishDate?: string;
  actualStartDate?: string;
  actualFinishDate?: string;
}

export interface MOMaterial {
  id: number;
  componentNumber: string;
  componentDescription?: string;
  requiredQuantity: number;
  allocatedQuantity: number;
  issuedQuantity: number;
  quantityUnit?: string;
  plannedIssueDate?: string;
  actualIssueDate?: string;
  warehouse?: string;
  location?: string;
}

// Planned Manufacturing Order (full details)
export interface PlannedManufacturingOrder {
  id: number;
  mopNumber: string;
  facility: string;
  itemNumber: string;
  itemDescription?: string;
  plannedQuantity: number;
  quantityUnit?: string;
  plannedOrderDate?: string;
  plannedStartDate?: string;
  plannedFinishDate?: string;
  requirementDate?: string;
  status: string;
  proposalStatus?: string;
  demandOrderCategory?: string;
  demandOrderNumber?: string;
  demandOrderLine?: string;
  orderPolicy?: string;
  lotSize?: number;
  safetyStock?: number;
  reorderPoint?: number;
  warehouse?: string;
  buyer?: string;
  planner?: string;
}

// Customer Order types
export interface CustomerOrder {
  id: number;
  orderNumber: string;
  customerNumber: string;
  customerName?: string;
  orderType?: string;
  orderDate?: string;
  requestedDeliveryDate?: string;
  confirmedDeliveryDate?: string;
  status: string;
  currency?: string;
  totalAmount?: number;
  warehouse?: string;
  salesPerson?: string;
  lines?: CustomerOrderLine[];
}

export interface CustomerOrderLine {
  id: number;
  orderNumber: string;
  lineNumber: string;
  lineSuffix?: string;
  itemNumber: string;
  itemDescription?: string;
  orderedQuantity: number;
  deliveredQuantity: number;
  quantityUnit?: string;
  requestedDeliveryDate?: string;
  confirmedDeliveryDate?: string;
  actualDeliveryDate?: string;
  status: string;
  lineType?: string;
  unitPrice?: number;
  lineAmount?: number;
  warehouse?: string;
}

// Delivery types
export interface Delivery {
  id: number;
  deliveryNumber: string;
  orderNumber: string;
  lineNumber: string;
  deliveryType?: string;
  deliveryQuantity: number;
  quantityUnit?: string;
  plannedDeliveryDate?: string;
  confirmedDeliveryDate?: string;
  actualDeliveryDate?: string;
  status: string;
  warehouse?: string;
  shipmentNumber?: string;
}

// Inconsistency types
export interface Inconsistency {
  id: number;
  type: 'date_mismatch' | 'missing_link' | 'quantity_mismatch' | 'status_conflict';
  severity: 'low' | 'medium' | 'high' | 'critical';
  description: string;
  productionOrder?: ProductionOrder;
  customerOrder?: CustomerOrder;
  delivery?: Delivery;
  details: Record<string, any>;
  createdAt: string;
  isIgnored?: boolean;
}

// Snapshot types
export interface PhaseProgress {
  phase: string;                    // "mops" | "mos" | "cos"
  status: 'pending' | 'running' | 'completed' | 'failed';
  recordCount?: number;
  startTime?: string;
  endTime?: string;
  error?: string;
}

export interface DetectorProgress {
  detectorName: string;             // "unlinked_production_orders"
  displayLabel: string;             // "Unlinked Production Orders"
  status: 'pending' | 'running' | 'completed' | 'failed';
  issuesFound?: number;
  durationMs?: number;
  startTime?: string;
  endTime?: string;
  error?: string;
}

export interface SnapshotStatus {
  jobId?: string;
  status: 'idle' | 'running' | 'completed' | 'failed' | 'cancelled';
  lastUpdate?: string;
  progress: number;
  currentStep?: string;
  completedSteps?: number;
  totalSteps?: number;
  parallelPhases?: PhaseProgress[];      // Parallel data loading tracking
  parallelDetectors?: DetectorProgress[]; // Parallel detector tracking
  coLinesProcessed?: number;
  mosProcessed?: number;
  mopsProcessed?: number;
  recordsPerSecond?: number;           // Processing rate
  estimatedTimeRemaining?: number;     // Seconds remaining
  currentOperation?: string;           // Detailed operation description
  currentBatch?: number;               // Current batch number
  totalBatches?: number;               // Total batches
  error?: string;
}

export interface SnapshotSummary {
  totalProductionOrders: number;
  totalManufacturingOrders: number;
  totalPlannedOrders: number;
  totalCustomerOrderLines: number;
  totalDeliveries: number;
  lastRefresh?: string;
  inconsistenciesCount: number;
}

// Settings types
export interface UserSettings {
  userId: string;
  defaultWarehouse?: string;
  defaultFacility?: string;
  defaultDivision?: string;
  defaultCompany?: string;
}

export interface SystemSetting {
  key: string;
  value: string;
  type: 'string' | 'integer' | 'float' | 'boolean' | 'json';
  description?: string;
  category: string;
  constraints?: Record<string, any>;
}

export interface SystemSettingsGrouped {
  categories: Record<string, SystemSetting[]>;
}

export interface CacheStatus {
  resourceType: string;
  recordCount: number;
  lastRefresh: string;
  isStale: boolean;
}

export interface RefreshResult {
  status: 'success' | 'failed';
  message: string;
  companiesRefreshed?: number;
  divisionsRefreshed?: number;
  facilitiesRefreshed?: number;
  warehousesRefreshed?: number;
  durationMs?: number;
}
