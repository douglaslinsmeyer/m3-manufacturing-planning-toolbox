import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './contexts/AuthContext';
import { useAuth } from './contexts/AuthContext';
import Login from './pages/Login';
import Dashboard from './pages/Dashboard';
import ProductionOrders from './pages/ProductionOrders';
import ManufacturingOrderDetail from './pages/ManufacturingOrderDetail';
import PlannedOrderDetail from './pages/PlannedOrderDetail';
import CustomerOrders from './pages/CustomerOrders';
import Deliveries from './pages/Deliveries';
import Inconsistencies from './pages/Inconsistencies';
import './App.css';

// Protected route wrapper
const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isAuthenticated, loading } = useAuth();

  // Don't render anything until auth check completes
  if (loading) {
    return (
      <div style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        height: '100vh',
        fontSize: '18px'
      }}>
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
      <Router>
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
            path="/production-orders"
            element={
              <ProtectedRoute>
                <ProductionOrders />
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
            path="/customer-orders"
            element={
              <ProtectedRoute>
                <CustomerOrders />
              </ProtectedRoute>
            }
          />
          <Route
            path="/deliveries"
            element={
              <ProtectedRoute>
                <Deliveries />
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
    </AuthProvider>
  );
}

export default App;
