import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { api } from '../services/api';
import { ContextBar } from '../components/ContextBar';
import { ContextSwitcher } from '../components/ContextSwitcher';
import type { SnapshotSummary, SnapshotStatus } from '../types';
import './Dashboard.css';

const Dashboard: React.FC = () => {
  const { environment, userContext, logout, isAuthenticated, loading: authLoading } = useAuth();
  const [summary, setSummary] = useState<SnapshotSummary | null>(null);
  const [snapshotStatus, setSnapshotStatus] = useState<SnapshotStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [contextSwitcherOpen, setContextSwitcherOpen] = useState(false);

  useEffect(() => {
    // Only load data if auth check completed AND user is authenticated
    if (authLoading || !isAuthenticated) {
      return;
    }

    loadDashboardData();
    const interval = setInterval(checkSnapshotStatus, 5000);
    return () => clearInterval(interval);
  }, [isAuthenticated, authLoading]);

  const loadDashboardData = async () => {
    try {
      const [summaryData, statusData] = await Promise.all([
        api.getSnapshotSummary(),
        api.getSnapshotStatus(),
      ]);
      setSummary(summaryData);
      setSnapshotStatus(statusData);
    } catch (error) {
      console.error('Failed to load dashboard data:', error);
    } finally {
      setLoading(false);
    }
  };

  const checkSnapshotStatus = async () => {
    try {
      const status = await api.getSnapshotStatus();
      setSnapshotStatus(status);
      if (status.status === 'completed') {
        loadDashboardData();
      }
    } catch (error) {
      console.error('Failed to check snapshot status:', error);
    }
  };

  const handleRefresh = async () => {
    setRefreshing(true);
    try {
      await api.refreshSnapshot();
      checkSnapshotStatus();
    } catch (error) {
      console.error('Failed to refresh snapshot:', error);
    } finally {
      setRefreshing(false);
    }
  };

  const handleLogout = async () => {
    try {
      await logout();
    } catch (error) {
      console.error('Logout failed:', error);
    }
  };

  const handleSwitchEnvironment = async () => {
    if (window.confirm(`Switch to ${environment === 'TRN' ? 'PRD' : 'TRN'} environment? This will log you out and clear all data.`)) {
      try {
        await logout();
        // Redirect to login will happen automatically via AuthContext
      } catch (error) {
        console.error('Failed to switch environment:', error);
      }
    }
  };

  if (loading) {
    return <div className="loading">Loading dashboard...</div>;
  }

  return (
    <div className="dashboard">
      <header className="dashboard-header">
        <div className="header-left">
          <h1>M3 Manufacturing Planning Tools</h1>
          <button
            onClick={handleSwitchEnvironment}
            className="environment-badge clickable"
            title={`Switch to ${environment === 'TRN' ? 'PRD' : 'TRN'}`}
          >
            {environment}
          </button>
        </div>
        <div className="header-right">
          <ContextBar onOpenSwitcher={() => setContextSwitcherOpen(true)} />
          <button onClick={handleLogout} className="logout-button">
            Logout
          </button>
        </div>
      </header>

      {/* Context Switcher Modal */}
      <ContextSwitcher
        isOpen={contextSwitcherOpen}
        onClose={() => setContextSwitcherOpen(false)}
      />

      <div className="dashboard-content">
        <div className="snapshot-section">
          <div className="section-header">
            <h2>Data Snapshot</h2>
            <button
              onClick={handleRefresh}
              disabled={refreshing || snapshotStatus?.status === 'running'}
              className="refresh-button"
            >
              {refreshing || snapshotStatus?.status === 'running' ? 'Refreshing...' : 'Refresh Data'}
            </button>
          </div>

          {snapshotStatus?.status === 'running' && (
            <div className="snapshot-progress">
              <div className="progress-bar">
                <div
                  className="progress-fill"
                  style={{ width: `${snapshotStatus.progress}%` }}
                />
              </div>
              <p>{snapshotStatus.currentStep || 'Processing...'}</p>
            </div>
          )}

          {summary?.lastRefresh && (
            <p className="last-refresh">
              Last refreshed: {new Date(summary.lastRefresh).toLocaleString()}
            </p>
          )}
        </div>

        <div className="stats-grid">
          <Link to="/production-orders" className="stat-card">
            <h3>Production Orders</h3>
            <div className="stat-number">{summary?.totalProductionOrders || 0}</div>
            <p className="stat-detail">
              {summary?.totalManufacturingOrders || 0} MOs, {summary?.totalPlannedOrders || 0} MOPs
            </p>
          </Link>

          <Link to="/customer-orders" className="stat-card">
            <h3>Customer Orders</h3>
            <div className="stat-number">{summary?.totalCustomerOrders || 0}</div>
          </Link>

          <Link to="/deliveries" className="stat-card">
            <h3>Deliveries</h3>
            <div className="stat-number">{summary?.totalDeliveries || 0}</div>
          </Link>

          <Link to="/inconsistencies" className="stat-card highlight">
            <h3>Inconsistencies</h3>
            <div className="stat-number">{summary?.inconsistenciesCount || 0}</div>
            <p className="stat-detail">Issues requiring attention</p>
          </Link>
        </div>

        <div className="quick-actions">
          <h2>Quick Actions</h2>
          <div className="action-buttons">
            <Link to="/production-orders" className="action-button">
              View Production Timeline
            </Link>
            <Link to="/inconsistencies" className="action-button">
              Review Inconsistencies
            </Link>
            <Link to="/customer-orders" className="action-button">
              Browse Customer Orders
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
