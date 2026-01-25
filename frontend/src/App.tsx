import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './contexts/AuthContext';
import { useAuth } from './contexts/AuthContext';
import { ContextManagementProvider } from './contexts/ContextManagementContext';
import Login from './pages/Login';
import Dashboard from './pages/Dashboard';
import ManufacturingOrderDetail from './pages/ManufacturingOrderDetail';
import PlannedOrderDetail from './pages/PlannedOrderDetail';
import Inconsistencies from './pages/Inconsistencies';

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

function App() {
  return (
    <AuthProvider>
      <ContextManagementProvider>
        <Router future={{ v7_startTransition: true, v7_relativeSplatPath: true }}>
          <Routes>
            <Route path="/login" element={<Login />} />
          <Route
            path="/"
            element={
              <ProtectedRoute>
                <Dashboard />
              </ProtectedRoute>
            }
          />
          <Route
            path="/manufacturing-orders/:id"
            element={
              <ProtectedRoute>
                <ManufacturingOrderDetail />
              </ProtectedRoute>
            }
          />
          <Route
            path="/planned-orders/:id"
            element={
              <ProtectedRoute>
                <PlannedOrderDetail />
              </ProtectedRoute>
            }
          />
          <Route
            path="/inconsistencies"
            element={
              <ProtectedRoute>
                <Inconsistencies />
              </ProtectedRoute>
            }
          />
          </Routes>
        </Router>
      </ContextManagementProvider>
    </AuthProvider>
  );
}

export default App;
