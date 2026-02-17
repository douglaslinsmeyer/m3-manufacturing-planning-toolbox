import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './contexts/AuthContext';
import { useAuth } from './contexts/AuthContext';
import { ContextManagementProvider } from './contexts/ContextManagementContext';
import Login from './pages/Login';
import Dashboard from './pages/Dashboard';
import ManufacturingOrderDetail from './pages/ManufacturingOrderDetail';
import PlannedOrderDetail from './pages/PlannedOrderDetail';
import Issues from './pages/Issues';
import Anomalies from './pages/Anomalies';
import Settings from './pages/Settings';
import Profile from './pages/Profile';
import AuditLogs from './pages/AuditLogs';

// Protected route wrapper
const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isAuthenticated, loading } = useAuth();

  // Don't render anything until auth check completes
  if (loading) {
    return (
      <div className="flex justify-center items-center h-screen text-lg">
        Loading...
      </div>
    );
  }

  // After loading completes, redirect to login if not authenticated
  // This ensures protected components only mount when authenticated
  return isAuthenticated ? <>{children}</> : <Navigate to="/login" replace />;
};

// Wrapper for protected routes that includes ContextManagementProvider
const ProtectedRouteWithContext: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  return (
    <ProtectedRoute>
      <ContextManagementProvider>
        {children}
      </ContextManagementProvider>
    </ProtectedRoute>
  );
};

function App() {
  return (
    <AuthProvider>
      <Router future={{ v7_startTransition: true, v7_relativeSplatPath: true }}>
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route
            path="/"
            element={
              <ProtectedRouteWithContext>
                <Dashboard />
              </ProtectedRouteWithContext>
            }
          />
          <Route
            path="/manufacturing-orders/:id"
            element={
              <ProtectedRouteWithContext>
                <ManufacturingOrderDetail />
              </ProtectedRouteWithContext>
            }
          />
          <Route
            path="/planned-orders/:id"
            element={
              <ProtectedRouteWithContext>
                <PlannedOrderDetail />
              </ProtectedRouteWithContext>
            }
          />
          <Route
            path="/issues"
            element={
              <ProtectedRouteWithContext>
                <Issues />
              </ProtectedRouteWithContext>
            }
          />
          <Route
            path="/anomalies"
            element={
              <ProtectedRouteWithContext>
                <Anomalies />
              </ProtectedRouteWithContext>
            }
          />
          <Route
            path="/settings"
            element={
              <ProtectedRouteWithContext>
                <Settings />
              </ProtectedRouteWithContext>
            }
          />
          <Route
            path="/profile"
            element={
              <ProtectedRouteWithContext>
                <Profile />
              </ProtectedRouteWithContext>
            }
          />
          <Route
            path="/audit-logs"
            element={
              <ProtectedRouteWithContext>
                <AuditLogs />
              </ProtectedRouteWithContext>
            }
          />
        </Routes>
      </Router>
    </AuthProvider>
  );
}

export default App;
