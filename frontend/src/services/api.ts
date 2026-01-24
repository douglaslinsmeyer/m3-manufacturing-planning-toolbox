import axios, { AxiosInstance } from 'axios';
import type {
  AuthStatus,
  UserContext,
  ProductionOrder,
  ManufacturingOrder,
  PlannedManufacturingOrder,
  CustomerOrder,
  Delivery,
  Inconsistency,
  SnapshotStatus,
  SnapshotSummary,
  EffectiveContext,
  TemporaryOverride,
  M3Company,
  M3Division,
  M3Facility,
  M3Warehouse,
} from '../types';

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

  // Production Orders
  async listProductionOrders(params?: {
    facility?: string;
    startDate?: string;
    endDate?: string;
    status?: string;
    type?: 'MO' | 'MOP';
  }): Promise<ProductionOrder[]> {
    const response = await this.client.get('/production-orders', { params });
    return response.data;
  }

  async getProductionOrder(id: number): Promise<ProductionOrder> {
    const response = await this.client.get(`/production-orders/${id}`);
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

  // Customer Orders
  async listCustomerOrders(params?: {
    customerNumber?: string;
    startDate?: string;
    endDate?: string;
    status?: string;
  }): Promise<CustomerOrder[]> {
    const response = await this.client.get('/customer-orders', { params });
    return response.data;
  }

  async getCustomerOrder(id: number): Promise<CustomerOrder> {
    const response = await this.client.get(`/customer-orders/${id}`);
    return response.data;
  }

  // Deliveries
  async listDeliveries(params?: {
    orderNumber?: string;
    startDate?: string;
    endDate?: string;
    status?: string;
  }): Promise<Delivery[]> {
    const response = await this.client.get('/deliveries', { params });
    return response.data;
  }

  async getDelivery(id: number): Promise<Delivery> {
    const response = await this.client.get(`/deliveries/${id}`);
    return response.data;
  }

  // Analysis
  async listInconsistencies(params?: {
    severity?: string;
    type?: string;
  }): Promise<Inconsistency[]> {
    const response = await this.client.get('/analysis/inconsistencies', { params });
    return response.data;
  }

  async getTimeline(params?: {
    startDate?: string;
    endDate?: string;
    facility?: string;
  }): Promise<any> {
    const response = await this.client.get('/analysis/timeline', { params });
    return response.data;
  }
}

export const api = new ApiService();
export default api;
