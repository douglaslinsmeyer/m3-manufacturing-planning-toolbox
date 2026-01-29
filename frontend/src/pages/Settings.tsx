import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { AppLayout } from '../components/AppLayout';
import { useAuth } from '../contexts/AuthContext';
import { api } from '../services/api';
import type { SystemSettingsGrouped, CacheStatus, RefreshResult } from '../types';
import DetectorSection from '../components/DetectorSection';
import { ToastContainer } from '../components/Toast';
import { useToast } from '../hooks/useToast';

const Settings: React.FC = () => {
  const navigate = useNavigate();
  const { userProfile } = useAuth();
  const toast = useToast();
  const [activeTab, setActiveTab] = useState<'data-refresh' | 'detectors' | 'anomalies'>('data-refresh');
  const [systemSettings, setSystemSettings] = useState<SystemSettingsGrouped | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  // Context cache management state
  const [cacheStatus, setCacheStatus] = useState<CacheStatus[] | null>(null);
  const [refreshingCache, setRefreshingCache] = useState(false);

  // Check if user is admin
  const isAdmin = userProfile?.groups?.some(
    g => g.type === 'Security Role' && g.display === 'Infor-SystemAdministrator'
  ) || false;

  // Redirect non-admin users to profile
  useEffect(() => {
    if (userProfile && !isAdmin) {
      navigate('/profile');
    }
  }, [isAdmin, userProfile, navigate]);

  // Load system settings and cache status on mount
  useEffect(() => {
    if (isAdmin) {
      loadSystemSettings();
      if (activeTab === 'data-refresh') {
        loadCacheStatus();
      }
    }
  }, [isAdmin, activeTab]);

  const loadSystemSettings = async () => {
    setLoading(true);
    try {
      const settings = await api.getSystemSettings();
      setSystemSettings(settings);
    } catch (err: any) {
      toast.error(err.response?.data || 'Failed to load system settings');
    } finally {
      setLoading(false);
    }
  };

  const loadCacheStatus = async () => {
    try {
      const status = await api.getContextCacheStatus();
      setCacheStatus(status);
    } catch (err: any) {
      console.error('Failed to load cache status:', err);
      toast.error('Failed to load cache status');
    }
  };

  const handleRefreshContextCache = async () => {
    setRefreshingCache(true);
    try {
      const result = await api.refreshContextCache('all');
      toast.success(`Cache refreshed: ${result.companiesRefreshed} companies, ${result.divisionsRefreshed} divisions, ${result.warehousesRefreshed} warehouses (${result.durationMs}ms)`);
      await loadCacheStatus(); // Reload status after refresh
    } catch (err: any) {
      toast.error(err.response?.data?.message || 'Failed to refresh cache');
    } finally {
      setRefreshingCache(false);
    }
  };

  const handleSaveSystemSettings = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!systemSettings) return;

    setSaving(true);
    try {
      // Convert systemSettings back to flat map
      const flatSettings: Record<string, string> = {};
      Object.values(systemSettings.categories).flat().forEach(setting => {
        flatSettings[setting.key] = setting.value;
      });

      await api.updateSystemSettings(flatSettings);
      toast.success('System settings saved successfully');
    } catch (err: any) {
      toast.error(err.response?.data || 'Failed to save system settings');
    } finally {
      setSaving(false);
    }
  };

  // Don't render if not admin (redirect will happen)
  if (!isAdmin) {
    return null;
  }

  return (
    <AppLayout>
      <ToastContainer toasts={toast.toasts} onClose={toast.removeToast} />
      <div className="px-4 py-6 sm:px-6 lg:px-12 lg:py-10">
        <div className="max-w-5xl mx-auto">
          {/* Header */}
          <div className="mb-6">
            <h1 className="text-3xl font-bold text-slate-900">System Settings</h1>
            <p className="mt-2 text-sm text-slate-600">
              Manage system-wide configuration (Admin Only)
            </p>
          </div>

          {/* Tabs */}
          <div className="border-b border-slate-200 mb-6">
            <nav className="-mb-px flex space-x-8">
              <button
                onClick={() => setActiveTab('data-refresh')}
                className={`py-4 px-1 border-b-2 font-medium text-sm transition-colors ${
                  activeTab === 'data-refresh'
                    ? 'border-primary-500 text-primary-600'
                    : 'border-transparent text-slate-500 hover:text-slate-700 hover:border-slate-300'
                }`}
              >
                Data Refresh
              </button>
              <button
                onClick={() => setActiveTab('detectors')}
                className={`py-4 px-1 border-b-2 font-medium text-sm transition-colors ${
                  activeTab === 'detectors'
                    ? 'border-primary-500 text-primary-600'
                    : 'border-transparent text-slate-500 hover:text-slate-700 hover:border-slate-300'
                }`}
              >
                Detectors
              </button>
              <button
                onClick={() => setActiveTab('anomalies')}
                className={`py-4 px-1 border-b-2 font-medium text-sm transition-colors ${
                  activeTab === 'anomalies'
                    ? 'border-primary-500 text-primary-600'
                    : 'border-transparent text-slate-500 hover:text-slate-700 hover:border-slate-300'
                }`}
              >
                Anomalies
              </button>
            </nav>
          </div>

          {/* Content */}
          {loading ? (
            <div className="flex items-center justify-center h-64">
              <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary-200 border-t-primary-600" />
            </div>
          ) : (
            <SystemSettingsForm
              settings={systemSettings}
              onSettingsChange={setSystemSettings}
              onSave={handleSaveSystemSettings}
              saving={saving}
              activeTab={activeTab}
              cacheStatus={cacheStatus}
              refreshingCache={refreshingCache}
              onRefreshCache={handleRefreshContextCache}
            />
          )}
        </div>
      </div>
    </AppLayout>
  );
};

// System Settings Form Component
interface SystemSettingsFormProps {
  settings: SystemSettingsGrouped | null;
  onSettingsChange: (settings: SystemSettingsGrouped | null) => void;
  onSave: (e: React.FormEvent) => void;
  saving: boolean;
  activeTab: 'data-refresh' | 'detectors' | 'anomalies';
  cacheStatus: CacheStatus[] | null;
  refreshingCache: boolean;
  onRefreshCache: () => Promise<void>;
}

const SystemSettingsForm: React.FC<SystemSettingsFormProps> = ({
  settings,
  onSettingsChange,
  onSave,
  saving,
  activeTab,
  cacheStatus,
  refreshingCache,
  onRefreshCache,
}) => {
  if (!settings) return null;

  const updateSettingValue = (category: string, key: string, value: string) => {
    const updated = { ...settings };
    const categorySettings = [...updated.categories[category]];
    const settingIndex = categorySettings.findIndex(s => s.key === key);
    if (settingIndex !== -1) {
      categorySettings[settingIndex] = {
        ...categorySettings[settingIndex],
        value,
      };
      updated.categories[category] = categorySettings;
      onSettingsChange(updated);
    }
  };

  // Map tab to categories to display
  const getCategoriesToShow = (): string[] => {
    switch (activeTab) {
      case 'data-refresh':
        return ['data_refresh'];
      case 'detectors':
        return ['detection'];
      default:
        return [];
    }
  };

  const categoriesToShow = getCategoriesToShow();

  const categoryLabels: Record<string, string> = {
    data_refresh: 'Data Refresh Performance',
    detection: 'Detector Configuration',
  };

  // Helper function to format relative time
  const formatAge = (timestamp: string): string => {
    if (!timestamp) return 'Never';
    const date = new Date(timestamp);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
    const diffDays = Math.floor(diffHours / 24);

    if (diffDays > 0) return `${diffDays} day${diffDays > 1 ? 's' : ''} ago`;
    if (diffHours > 0) return `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`;
    return 'Just now';
  };

  const formatDateTime = (timestamp: string): string => {
    if (!timestamp) return 'Never';
    return new Date(timestamp).toLocaleString();
  };

  // Special handling for data-refresh tab - show cache management UI first
  if (activeTab === 'data-refresh') {
    return (
      <form onSubmit={onSave}>
        <div className="space-y-6">
          {/* Context Cache Section */}
          <div className="bg-white shadow rounded-lg">
            <div className="px-6 py-5 border-b border-slate-200">
              <h3 className="text-lg font-semibold text-slate-900 mb-2">Context Cache</h3>
              <p className="text-sm text-slate-600 mb-4">
                M3 organizational and reference data is cached for performance.
                Refresh manually when M3 data changes.
              </p>

              {/* Cache Status Table */}
              {cacheStatus && cacheStatus.length > 0 ? (
                <div className="bg-slate-50 rounded-lg p-4 mb-4">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b border-slate-300">
                        <th className="text-left py-2 font-semibold">Resource</th>
                        <th className="text-right py-2 font-semibold">Records</th>
                        <th className="text-right py-2 font-semibold">Last Refresh</th>
                        <th className="text-right py-2 font-semibold">Age</th>
                      </tr>
                    </thead>
                    <tbody>
                      {cacheStatus.map(status => (
                        <tr key={status.resourceType} className="border-b border-slate-200 last:border-0">
                          <td className="py-2">
                            {status.resourceType}
                            {status.resourceType === 'User Profiles' && (
                              <span className="ml-2 text-xs text-slate-500">
                                (all users)
                              </span>
                            )}
                          </td>
                          <td className="text-right">{status.recordCount}</td>
                          <td className="text-right text-xs">{formatDateTime(status.lastRefresh)}</td>
                          <td className="text-right">
                            <span className={status.isStale ? 'text-amber-600 font-medium' : 'text-slate-600'}>
                              {formatAge(status.lastRefresh)}
                            </span>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              ) : (
                <div className="bg-slate-50 rounded-lg p-4 mb-4 text-center text-sm text-slate-500">
                  Loading cache status...
                </div>
              )}

              {/* Refresh Button */}
              <button
                type="button"
                onClick={onRefreshCache}
                disabled={refreshingCache}
                className="px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed font-medium transition-colors"
              >
                {refreshingCache ? 'Refreshing...' : 'Refresh All Context Cache'}
              </button>
            </div>
          </div>

          {/* Performance Settings */}
          {settings.categories['data_refresh'] && (
            <div className="bg-white shadow rounded-lg">
              <div className="px-6 py-5">
                <h3 className="text-lg font-semibold text-slate-900 mb-4">Performance Settings</h3>
                <div className="space-y-4">
                  {settings.categories['data_refresh'].map((setting) => (
                    <div key={setting.key}>
                      <label className="block text-sm font-medium text-slate-700 mb-1">
                        {setting.description || setting.key}
                        {setting.constraints && setting.constraints.unit && (
                          <span className="ml-1 text-slate-500">({setting.constraints.unit})</span>
                        )}
                      </label>
                      {setting.type === 'boolean' ? (
                        <select
                          value={setting.value}
                          onChange={(e) => updateSettingValue('data_refresh', setting.key, e.target.value)}
                          className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                        >
                          <option value="true">Enabled</option>
                          <option value="false">Disabled</option>
                        </select>
                      ) : (
                        <input
                          type={setting.type === 'integer' || setting.type === 'float' ? 'number' : 'text'}
                          value={setting.value}
                          onChange={(e) => updateSettingValue('data_refresh', setting.key, e.target.value)}
                          min={setting.constraints?.min}
                          max={setting.constraints?.max}
                          step={setting.type === 'float' ? '0.01' : undefined}
                          className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                        />
                      )}
                      <p className="mt-1 text-xs text-slate-500">
                        Key: {setting.key}
                        {setting.constraints && (setting.constraints.min !== undefined || setting.constraints.max !== undefined) && (
                          <span className="ml-2">
                            Range: {setting.constraints.min || '−∞'} to {setting.constraints.max || '∞'}
                          </span>
                        )}
                      </p>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Save Button for Performance Settings */}
        {settings.categories['data_refresh'] && settings.categories['data_refresh'].length > 0 && (
          <div className="mt-6 flex justify-end">
            <button
              type="submit"
              disabled={saving}
              className="px-6 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed font-medium transition-colors"
            >
              {saving ? 'Saving...' : 'Save Performance Settings'}
            </button>
          </div>
        )}
      </form>
    );
  }

  // Special handling for detectors tab - use custom detector sections
  if (activeTab === 'detectors' && settings.categories['detection']) {
    return (
      <form onSubmit={onSave}>
        <div className="space-y-6">
          {/* Unlinked Production Orders Detector Section */}
          <DetectorSection
            detectorName="unlinked_production_orders"
            detectorLabel="Unlinked Production Orders"
            detectorDescription="Detects production orders without customer order linkage"
            settings={settings.categories['detection']}
            onSettingsChange={(updated) => {
              const newSettings = { ...settings };
              newSettings.categories['detection'] = updated;
              onSettingsChange(newSettings);
            }}
          />

          {/* Joint Delivery Date Mismatch Detector Section */}
          <DetectorSection
            detectorName="joint_delivery_date_mismatch"
            detectorLabel="Joint Delivery Date Mismatch"
            detectorDescription="Detects production orders within same joint delivery group with mismatched delivery dates"
            settings={settings.categories['detection']}
            onSettingsChange={(updated) => {
              const newSettings = { ...settings };
              newSettings.categories['detection'] = updated;
              onSettingsChange(newSettings);
            }}
          />

          {/* DLIX Date Mismatch Detector Section */}
          <DetectorSection
            detectorName="dlix_date_mismatch"
            detectorLabel="Delivery Date Mismatches"
            detectorDescription="Detects production orders within same delivery (DLIX) with misaligned start dates"
            settings={settings.categories['detection']}
            onSettingsChange={(updated) => {
              const newSettings = { ...settings };
              newSettings.categories['detection'] = updated;
              onSettingsChange(newSettings);
            }}
          />
        </div>

        {/* Save Button */}
        <div className="mt-6 flex justify-end">
          <button
            type="submit"
            disabled={saving}
            className="px-6 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed font-medium transition-colors"
          >
            {saving ? 'Saving...' : 'Save Detector Settings'}
          </button>
        </div>
      </form>
    );
  }

  // Special handling for anomalies tab - use custom detector sections
  if (activeTab === 'anomalies' && settings.categories['anomaly_detection']) {
    return (
      <form onSubmit={onSave}>
        <div className="space-y-6">
          <div className="mb-6">
            <h2 className="text-xl font-semibold text-slate-900 mb-2">
              Anomaly Detection Settings
            </h2>
            <p className="text-sm text-slate-600">
              Configure thresholds for statistical anomaly detection across aggregate data patterns. Anomaly detectors analyze patterns in the data to identify potential planning issues or data quality problems.
            </p>
          </div>

          {/* Unlinked Concentration Detector */}
          <DetectorSection
            detectorName="unlinked_concentration"
            detectorLabel="Unlinked Concentration"
            detectorDescription="Detects when a single product accounts for an excessive percentage of unlinked MOPs, indicating potential runaway planning for that product."
            settings={settings.categories['anomaly_detection']}
            onSettingsChange={(updated) => {
              const newSettings = { ...settings };
              newSettings.categories['anomaly_detection'] = updated;
              onSettingsChange(newSettings);
            }}
            keyPrefix="anomaly"
          />

          {/* Date Clustering Detector */}
          <DetectorSection
            detectorName="date_clustering"
            detectorLabel="Date Clustering"
            detectorDescription="Detects when too many MOPs for a product are scheduled on the same date, indicating bulk planning issues or misconfiguration."
            settings={settings.categories['anomaly_detection']}
            onSettingsChange={(updated) => {
              const newSettings = { ...settings };
              newSettings.categories['anomaly_detection'] = updated;
              onSettingsChange(newSettings);
            }}
            keyPrefix="anomaly"
          />

          {/* MOP-to-Demand Ratio Detector */}
          <DetectorSection
            detectorName="mop_demand_ratio"
            detectorLabel="MOP-to-Demand Ratio"
            detectorDescription="Detects excessive unlinked MOPs relative to actual customer order demand, indicating over-planning."
            settings={settings.categories['anomaly_detection']}
            onSettingsChange={(updated) => {
              const newSettings = { ...settings };
              newSettings.categories['anomaly_detection'] = updated;
              onSettingsChange(newSettings);
            }}
            keyPrefix="anomaly"
          />

          {/* Absolute Volume Detector */}
          <DetectorSection
            detectorName="absolute_volume"
            detectorLabel="Absolute Volume"
            detectorDescription="Detects products with excessive absolute count of unlinked MOPs, regardless of percentages or ratios."
            settings={settings.categories['anomaly_detection']}
            onSettingsChange={(updated) => {
              const newSettings = { ...settings };
              newSettings.categories['anomaly_detection'] = updated;
              onSettingsChange(newSettings);
            }}
            keyPrefix="anomaly"
          />
        </div>

        {/* Save Button */}
        <div className="mt-6 flex justify-end">
          <button
            type="submit"
            disabled={saving}
            className="px-6 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed font-medium transition-colors"
          >
            {saving ? 'Saving...' : 'Save Anomaly Settings'}
          </button>
        </div>
      </form>
    );
  }

  // Standard category rendering for other tabs
  return (
    <form onSubmit={onSave}>
      <div className="bg-white shadow rounded-lg">
        {categoriesToShow.map((categoryKey, idx) => {
          const categorySettings = settings.categories[categoryKey];

          if (!categorySettings || categorySettings.length === 0) {
            return (
              <div key={categoryKey} className="px-6 py-12 text-center">
                <p className="text-slate-500">No settings available in this category</p>
              </div>
            );
          }

          return (
            <div
              key={categoryKey}
              className={`px-6 py-5 ${idx < categoriesToShow.length - 1 ? 'border-b border-slate-200' : ''}`}
            >
              <h3 className="text-lg font-semibold text-slate-900 mb-4">
                {categoryLabels[categoryKey] || categoryKey}
              </h3>
              <div className="space-y-4">
                {categorySettings.map((setting) => (
                  <div key={setting.key}>
                    <label className="block text-sm font-medium text-slate-700 mb-1">
                      {setting.description || setting.key}
                      {setting.constraints && setting.constraints.unit && (
                        <span className="ml-1 text-slate-500">({setting.constraints.unit})</span>
                      )}
                    </label>
                    {setting.type === 'boolean' ? (
                      <select
                        value={setting.value}
                        onChange={(e) => updateSettingValue(categoryKey, setting.key, e.target.value)}
                        className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                      >
                        <option value="true">Enabled</option>
                        <option value="false">Disabled</option>
                      </select>
                    ) : (
                      <input
                        type={setting.type === 'integer' || setting.type === 'float' ? 'number' : 'text'}
                        value={setting.value}
                        onChange={(e) => updateSettingValue(categoryKey, setting.key, e.target.value)}
                        min={setting.constraints?.min}
                        max={setting.constraints?.max}
                        step={setting.type === 'float' ? '0.01' : undefined}
                        className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                      />
                    )}
                    <p className="mt-1 text-xs text-slate-500">
                      Key: {setting.key}
                      {setting.constraints && (setting.constraints.min !== undefined || setting.constraints.max !== undefined) && (
                        <span className="ml-2">
                          Range: {setting.constraints.min || '−∞'} to {setting.constraints.max || '∞'}
                        </span>
                      )}
                    </p>
                  </div>
                ))}
              </div>
            </div>
          );
        })}
      </div>

      {/* Save Button */}
      <div className="mt-6 flex justify-end">
        <button
          type="submit"
          disabled={saving}
          className="px-6 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed font-medium transition-colors"
        >
          {saving ? 'Saving...' : 'Save System Settings'}
        </button>
      </div>
    </form>
  );
};

export default Settings;
