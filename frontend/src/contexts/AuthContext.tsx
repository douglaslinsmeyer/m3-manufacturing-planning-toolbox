import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { api } from '../services/api';
import type { AuthStatus, UserContext, UserProfile } from '../types';

interface AuthContextType {
  isAuthenticated: boolean;
  environment?: 'TRN' | 'PRD';
  userContext?: UserContext;
  userProfile?: UserProfile;
  loading: boolean;
  login: (environment: 'TRN' | 'PRD') => Promise<void>;
  logout: () => Promise<void>;
  setUserContext: (context: UserContext) => Promise<void>;
  refreshProfile: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [environment, setEnvironment] = useState<'TRN' | 'PRD' | undefined>();
  const [userContext, setUserContextState] = useState<UserContext | undefined>();
  const [userProfile, setUserProfile] = useState<UserProfile | undefined>();
  const [loading, setLoading] = useState(true);

  // Check authentication status on mount
  useEffect(() => {
    checkAuthStatus();
  }, []);

  const checkAuthStatus = async () => {
    try {
      const status: AuthStatus = await api.getAuthStatus();
      setIsAuthenticated(status.authenticated);
      setEnvironment(status.environment);
      setUserContextState(status.userContext);
      setUserProfile(status.userProfile);
    } catch (error) {
      console.error('Failed to check auth status:', error);
      setIsAuthenticated(false);
    } finally {
      setLoading(false);
    }
  };

  const login = async (env: 'TRN' | 'PRD') => {
    try {
      const { authUrl } = await api.login(env);
      // Redirect to OAuth provider
      window.location.href = authUrl;
    } catch (error) {
      console.error('Login failed:', error);
      throw error;
    }
  };

  const logout = async () => {
    try {
      await api.logout();
      setIsAuthenticated(false);
      setEnvironment(undefined);
      setUserContextState(undefined);
      setUserProfile(undefined);
    } catch (error) {
      console.error('Logout failed:', error);
      throw error;
    }
  };

  const setUserContext = async (context: UserContext) => {
    try {
      const updated = await api.setContext(context);
      setUserContextState(updated);
    } catch (error) {
      console.error('Failed to set user context:', error);
      throw error;
    }
  };

  const refreshProfile = async () => {
    try {
      const updated = await api.refreshProfile();
      setUserProfile(updated);
    } catch (error) {
      console.error('Failed to refresh profile:', error);
      throw error;
    }
  };

  return (
    <AuthContext.Provider
      value={{
        isAuthenticated,
        environment,
        userContext,
        userProfile,
        loading,
        login,
        logout,
        setUserContext,
        refreshProfile,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
