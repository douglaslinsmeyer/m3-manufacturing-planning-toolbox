import axios, { AxiosInstance } from 'axios';
import type {
  AuthStatus,
  UserContext,
  UserProfile,
  ManufacturingOrder,
  PlannedManufacturingOrder,
  Issue,
  SnapshotStatus,
  SnapshotSummary,
  EffectiveContext,
  TemporaryOverride,
  M3Company,
  M3Division,
  M3Facility,
  M3Warehouse,
  UserSettings,
  SystemSettingsGrouped,
  CacheStatus,
  RefreshResult,
  AuditLog,
  AuditLogFilters,
} from '../types';

// IssueSummary represents aggregated issue counts from the backend
export interface IssueSummary {
  total: number;
  by_detector: Record<string, number>;
  by_facility: Record<string, number>;
  by_warehouse: Record<string, number>;
  by_facility_warehouse_detector: Record<string, Record<string, Record<string, number>>>;
}

// PaginatedResponse represents a paginated API response
export interface PaginatedResponse<T> {
  data: T[];
  pagination: {
    page: number;
    pageSize: number;
    totalCount: number;
    totalPages: number;
  };
}

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

class ApiService {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: `${API_BASE_URL}/api`,
      withCredentials: true,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Global 401 interceptor - redirect to login if session expires
    this.client.interceptors.response.use(
      (response) => response,
      (error) => {
        if (error.response?.status === 401) {
          // Only redirect if we're not already on the login page or auth endpoints
          const isAuthEndpoint = error.config?.url?.includes('/auth/');
          const isLoginPage = window.location.pathname === '/login';

          if (!isAuthEndpoint && !isLoginPage) {
            console.warn('Session expired, redirecting to login');
            window.location.href = '/login';
          }
        }
        return Promise.reject(error);
      }
    );
  }

  // Authentication
  async login(environment: 'TRN' | 'PRD'): Promise<{ authUrl: string }> {
    const response = await this.client.post('/auth/login', { environment });
    return response.data;
  }

  async logout(): Promise<void> {
    await this.client.post('/auth/logout');
  }

  async refreshProfile(): Promise<UserProfile> {
    const response = await this.client.post('/auth/profile/refresh');
    return response.data;
  }

  async getAuthStatus(): Promise<AuthStatus> {
    const response = await this.client.get('/auth/status');
    return response.data;
  }

  // User Context
  async getContext(): Promise<UserContext> {
    const response = await this.client.get('/auth/context');
    return response.data;
  }

  async setContext(context: UserContext): Promise<UserContext> {
    const response = await this.client.post('/auth/context', context);
    return response.data;
  }

  // M3 Context Management
  async getEffectiveContext(): Promise<EffectiveContext> {
    const response = await this.client.get('/context/effective');
    return response.data;
  }

  async retryLoadContext(): Promise<EffectiveContext> {
    const response = await this.client.post('/context/retry-load');
    return response.data;
  }

  async setTemporaryOverride(override: TemporaryOverride): Promise<EffectiveContext> {
    const response = await this.client.post('/context/temporary', override);
    return response.data;
  }

  async clearTemporaryOverrides(): Promise<EffectiveContext> {
    const response = await this.client.delete('/context/temporary');
    return response.data;
  }

  // Organizational Hierarchy
  async listCompanies(): Promise<M3Company[]> {
    const response = await this.client.get('/context/companies');
    return response.data;
  }

  async listDivisions(companyNumber: string): Promise<M3Division[]> {
    const response = await this.client.get('/context/divisions', {
      params: { company: companyNumber },
    });
    return response.data;
  }

  async listFacilities(): Promise<M3Facility[]> {
    const response = await this.client.get('/context/facilities');
    return response.data;
  }

  async listWarehouses(
    companyNumber: string,
    division?: string,
    facility?: string
  ): Promise<M3Warehouse[]> {
    const response = await this.client.get('/context/warehouses', {
      params: {
        company: companyNumber,
        ...(division && { division }),
        ...(facility && { facility }),
      },
    });
    return response.data;
  }

  // Context Cache Management
  async getContextCacheStatus(): Promise<CacheStatus[]> {
    const response = await this.client.get('/context/cache-status');
    return response.data;
  }

  async refreshContextCache(resourceType: 'all' | 'companies' | 'divisions' | 'facilities' | 'warehouses'): Promise<RefreshResult> {
    const response = await this.client.post(`/context/refresh/${resourceType}`);
    return response.data;
  }

  // Snapshot Management
  async refreshSnapshot(): Promise<{ jobId: string; status: string; message: string }> {
    const response = await this.client.post('/snapshot/refresh');
    return response.data;
  }

  async getSnapshotStatus(): Promise<SnapshotStatus> {
    const response = await this.client.get('/snapshot/status');
    return response.data;
  }

  async getSnapshotSummary(): Promise<SnapshotSummary> {
    const response = await this.client.get('/snapshot/summary');
    return response.data;
  }

  async cancelRefresh(jobId: string): Promise<{ status: string; message: string }> {
    const response = await this.client.post(`/snapshot/refresh/${jobId}/cancel`);
    return response.data;
  }

  async getActiveJob(): Promise<{ jobId: string | null; status?: string }> {
    const response = await this.client.get('/snapshot/active-job');
    return response.data;
  }

  // Manufacturing Orders (full details)
  async getManufacturingOrder(id: number): Promise<ManufacturingOrder> {
    const response = await this.client.get(`/manufacturing-orders/${id}`);
    return response.data;
  }

  // Planned Orders (full details)
  async getPlannedOrder(id: number): Promise<PlannedManufacturingOrder> {
    const response = await this.client.get(`/planned-orders/${id}`);
    return response.data;
  }

  // Issues
  async getIssueSummary(includeIgnored: boolean = false): Promise<IssueSummary> {
    const params = includeIgnored ? { include_ignored: 'true' } : {};
    const response = await this.client.get('/issues/summary', { params });
    return response.data;
  }

  async listIssues(params?: {
    severity?: string;
    type?: string;
    warehouse?: string;
    includeIgnored?: boolean;
    page?: number;
    pageSize?: number;
  }): Promise<PaginatedResponse<Issue>> {
    const queryParams: any = {};
    if (params?.severity) queryParams.severity = params.severity;
    if (params?.type) queryParams.detector_type = params.type;
    if (params?.warehouse) queryParams.warehouse = params.warehouse;
    if (params?.includeIgnored) queryParams.include_ignored = 'true';
    if (params?.page) queryParams.page = params.page;
    if (params?.pageSize) queryParams.page_size = params.pageSize;

    const response = await this.client.get('/issues', { params: queryParams });
    return response.data;
  }

  async ignoreIssue(issueId: number, notes?: string): Promise<void> {
    await this.client.post(`/issues/${issueId}/ignore`, { notes });
  }

  async unignoreIssue(issueId: number): Promise<void> {
    await this.client.post(`/issues/${issueId}/unignore`);
  }

  async deletePlannedMO(issueId: number): Promise<{ success: boolean; m3_response?: any }> {
    const response = await this.client.post(`/issues/${issueId}/delete-mop`);
    return response.data;
  }

  async deleteMO(issueId: number): Promise<{ success: boolean; m3_response?: any }> {
    const response = await this.client.post(`/issues/${issueId}/delete-mo`);
    return response.data;
  }

  async closeMO(issueId: number): Promise<{ success: boolean; m3_response?: any }> {
    const response = await this.client.post(`/issues/${issueId}/close-mo`);
    return response.data;
  }

  async alignEarliestMOs(issueId: number): Promise<{
    success: boolean;
    aligned_count: number;
    skipped_count: number;
    failed_count: number;
    total_orders: number;
    target_date: string;
    date_adjusted?: boolean;
    original_min_date?: string;
    failures?: Array<{ order: string; type: string; error: string }>;
  }> {
    const response = await this.client.post(`/issues/${issueId}/align-earliest`);
    return response.data;
  }

  async alignLatestMOs(issueId: number): Promise<{
    success: boolean;
    aligned_count: number;
    skipped_count: number;
    failed_count: number;
    total_orders: number;
    target_date: string;
    date_adjusted?: boolean;
    original_max_date?: string;
    failures?: Array<{ order: string; type: string; error: string }>;
  }> {
    const response = await this.client.post(`/issues/${issueId}/align-latest`);
    return response.data;
  }

  // Bulk Issue Operations
  async bulkDelete(issueIds: number[]): Promise<{
    total: number;
    successful: number;
    failed: number;
    results: Array<{
      issue_id: number;
      production_order: string;
      status: 'success' | 'error';
      message?: string;
      error?: string;
    }>;
  }> {
    const response = await this.client.post('/issues/bulk-delete', { issue_ids: issueIds });
    return response.data;
  }

  async bulkClose(issueIds: number[]): Promise<{
    total: number;
    successful: number;
    failed: number;
    results: Array<{
      issue_id: number;
      production_order: string;
      status: 'success' | 'error';
      message?: string;
      error?: string;
    }>;
  }> {
    const response = await this.client.post('/issues/bulk-close', { issue_ids: issueIds });
    return response.data;
  }

  async bulkReschedule(issueIds: number[], newDate: string): Promise<{
    total: number;
    successful: number;
    failed: number;
    results: Array<{
      issue_id: number;
      production_order: string;
      status: 'success' | 'error';
      message?: string;
      error?: string;
    }>;
  }> {
    const response = await this.client.post('/issues/bulk-reschedule', {
      issue_ids: issueIds,
      params: { new_date: newDate },
    });
    return response.data;
  }

  // Anomalies
  async getAnomalySummary(): Promise<{
    total: number;
    by_severity: Record<string, number>;
    by_detector: Record<string, number>;
  }> {
    const response = await this.client.get('/anomalies/summary');
    return response.data;
  }

  async listAnomalies(params?: {
    severity?: string;
    detectorType?: string;
    page?: number;
    pageSize?: number;
  }): Promise<PaginatedResponse<any>> {
    const queryParams: any = {};
    if (params?.severity) queryParams.severity = params.severity;
    if (params?.detectorType) queryParams.detector_type = params.detectorType;
    if (params?.page) queryParams.page = params.page;
    if (params?.pageSize) queryParams.page_size = params.pageSize;

    const response = await this.client.get('/anomalies', { params: queryParams });
    return response.data;
  }

  async acknowledgeAnomaly(anomalyId: number, notes?: string): Promise<void> {
    await this.client.post(`/anomalies/${anomalyId}/acknowledge`, { notes: notes || '' });
  }

  async resolveAnomaly(anomalyId: number, notes?: string): Promise<void> {
    await this.client.post(`/anomalies/${anomalyId}/resolve`, { notes: notes || '' });
  }

  async getTimeline(params?: {
    startDate?: string;
    endDate?: string;
    facility?: string;
  }): Promise<any> {
    const response = await this.client.get('/analysis/timeline', { params });
    return response.data;
  }

  // User Settings
  async getUserSettings(): Promise<UserSettings> {
    const response = await this.client.get('/settings/user');
    return response.data;
  }

  async updateUserSettings(settings: Partial<UserSettings>): Promise<void> {
    await this.client.put('/settings/user', settings);
  }

  // System Settings (admin only)
  async getSystemSettings(): Promise<SystemSettingsGrouped> {
    const response = await this.client.get('/settings/system');
    return response.data;
  }

  async updateSystemSettings(settings: Record<string, string>): Promise<void> {
    await this.client.put('/settings/system', { settings });
  }

  // Audit Logs
  async listAuditLogs(filters?: AuditLogFilters): Promise<PaginatedResponse<AuditLog>> {
    const params = new URLSearchParams();
    if (filters?.entityType) params.append('entity_type', filters.entityType);
    if (filters?.operation) params.append('operation', filters.operation);
    if (filters?.userId) params.append('user_id', filters.userId);
    if (filters?.facility) params.append('facility', filters.facility);
    if (filters?.startTime) params.append('start_time', filters.startTime);
    if (filters?.endTime) params.append('end_time', filters.endTime);
    if (filters?.page) params.append('page', filters.page.toString());
    if (filters?.pageSize) params.append('page_size', filters.pageSize.toString());

    const response = await this.client.get(`/audit-logs?${params}`);
    return response.data;
  }
}

export const api = new ApiService();
export default api;
