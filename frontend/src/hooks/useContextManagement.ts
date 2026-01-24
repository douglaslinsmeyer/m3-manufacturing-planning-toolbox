import { useState, useCallback } from 'react';
import { api } from '../services/api';
import type {
  EffectiveContext,
  M3Company,
  M3Division,
  M3Facility,
  M3Warehouse,
  TemporaryOverride,
} from '../types';

export const useContextManagement = () => {
  const [effectiveContext, setEffectiveContext] = useState<EffectiveContext | null>(null);
  const [companies, setCompanies] = useState<M3Company[]>([]);
  const [divisions, setDivisions] = useState<M3Division[]>([]);
  const [facilities, setFacilities] = useState<M3Facility[]>([]);
  const [warehouses, setWarehouses] = useState<M3Warehouse[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Load effective context
  const loadEffectiveContext = useCallback(async () => {
    try {
      setLoading(true);
      const context = await api.getEffectiveContext();
      setEffectiveContext(context);
      setError(null);
    } catch (err) {
      setError('Failed to load context');
      console.error('Failed to load effective context:', err);
    } finally {
      setLoading(false);
    }
  }, []);

  // Load companies
  const loadCompanies = useCallback(async () => {
    try {
      setLoading(true);
      const data = await api.listCompanies();
      setCompanies(data);
      setError(null);
    } catch (err) {
      setError('Failed to load companies');
      console.error('Failed to load companies:', err);
    } finally {
      setLoading(false);
    }
  }, []);

  // Load divisions for a company
  const loadDivisions = useCallback(async (companyNumber: string) => {
    try {
      setLoading(true);
      const data = await api.listDivisions(companyNumber);
      setDivisions(data);
      setError(null);
    } catch (err) {
      setError('Failed to load divisions');
      console.error('Failed to load divisions:', err);
    } finally {
      setLoading(false);
    }
  }, []);

  // Load facilities
  const loadFacilities = useCallback(async () => {
    try {
      setLoading(true);
      const data = await api.listFacilities();
      setFacilities(data);
      setError(null);
    } catch (err) {
      setError('Failed to load facilities');
      console.error('Failed to load facilities:', err);
    } finally {
      setLoading(false);
    }
  }, []);

  // Load warehouses with optional filters
  const loadWarehouses = useCallback(async (
    companyNumber: string,
    division?: string,
    facility?: string
  ) => {
    try {
      setLoading(true);
      const data = await api.listWarehouses(companyNumber, division, facility);
      setWarehouses(data);
      setError(null);
    } catch (err) {
      setError('Failed to load warehouses');
      console.error('Failed to load warehouses:', err);
    } finally {
      setLoading(false);
    }
  }, []);

  // Set temporary override
  const setTemporaryOverride = useCallback(async (override: TemporaryOverride) => {
    try {
      setLoading(true);
      const context = await api.setTemporaryOverride(override);
      setEffectiveContext(context);
      setError(null);
      return context;
    } catch (err) {
      setError('Failed to set override');
      console.error('Failed to set temporary override:', err);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  // Clear temporary overrides
  const clearTemporaryOverrides = useCallback(async () => {
    try {
      setLoading(true);
      const context = await api.clearTemporaryOverrides();
      setEffectiveContext(context);
      setError(null);
      return context;
    } catch (err) {
      setError('Failed to clear overrides');
      console.error('Failed to clear temporary overrides:', err);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  // Retry loading context
  const retryLoadContext = useCallback(async () => {
    try {
      setLoading(true);
      const context = await api.retryLoadContext();
      setEffectiveContext(context);
      setError(null);
    } catch (err) {
      setError('Failed to load context. Please try again.');
      console.error('Failed to retry load context:', err);
    } finally {
      setLoading(false);
    }
  }, []);

  return {
    effectiveContext,
    companies,
    divisions,
    facilities,
    warehouses,
    loading,
    error,
    loadEffectiveContext,
    loadCompanies,
    loadDivisions,
    loadFacilities,
    loadWarehouses,
    setTemporaryOverride,
    clearTemporaryOverrides,
    retryLoadContext,
  };
};
