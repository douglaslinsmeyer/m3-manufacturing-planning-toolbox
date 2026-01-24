import React, { createContext, useContext, useEffect, ReactNode } from 'react';
import { useContextManagement as useContextManagementHook } from '../hooks/useContextManagement';
import { useAuth } from './AuthContext';
import type {
  EffectiveContext,
  M3Company,
  M3Division,
  M3Facility,
  M3Warehouse,
  TemporaryOverride,
} from '../types';

interface ContextManagementContextType {
  effectiveContext: EffectiveContext | null;
  companies: M3Company[];
  divisions: M3Division[];
  facilities: M3Facility[];
  warehouses: M3Warehouse[];
  loading: boolean;
  error: string | null;
  loadEffectiveContext: () => Promise<void>;
  loadCompanies: () => Promise<void>;
  loadDivisions: (companyNumber: string) => Promise<void>;
  loadFacilities: () => Promise<void>;
  loadWarehouses: (companyNumber: string, division?: string, facility?: string) => Promise<void>;
  setTemporaryOverride: (override: TemporaryOverride) => Promise<EffectiveContext>;
  clearTemporaryOverrides: () => Promise<EffectiveContext>;
  retryLoadContext: () => Promise<void>;
}

const ContextManagementContext = createContext<ContextManagementContextType | undefined>(undefined);

export const ContextManagementProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const { isAuthenticated, loading: authLoading } = useAuth();
  const contextManagement = useContextManagementHook();

  // Automatically load effective context when user is authenticated
  useEffect(() => {
    if (authLoading || !isAuthenticated) {
      return;
    }
    contextManagement.loadEffectiveContext();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isAuthenticated, authLoading]);

  return (
    <ContextManagementContext.Provider value={contextManagement}>
      {children}
    </ContextManagementContext.Provider>
  );
};

export const useContextManagement = () => {
  const context = useContext(ContextManagementContext);
  if (context === undefined) {
    throw new Error('useContextManagement must be used within a ContextManagementProvider');
  }
  return context;
};
