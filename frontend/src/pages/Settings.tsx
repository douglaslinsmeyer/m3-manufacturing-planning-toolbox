import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { AppLayout } from '../components/AppLayout';
import { useAuth } from '../contexts/AuthContext';
import { api } from '../services/api';
import type { SystemSettingsGrouped } from '../types';

const Settings: React.FC = () => {
  const navigate = useNavigate();
  const { userProfile } = useAuth();
  const [activeTab, setActiveTab] = useState<'data-refresh' | 'detectors' | 'integration'>('data-refresh');
  const [systemSettings, setSystemSettings] = useState<SystemSettingsGrouped | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

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

  // Load system settings on mount
  useEffect(() => {
    if (isAdmin) {
      loadSystemSettings();
    }
  }, [isAdmin]);

  const loadSystemSettings = async () => {
    setLoading(true);
    setError(null);
    try {
      const settings = await api.getSystemSettings();
      setSystemSettings(settings);
    } catch (err: any) {
      setError(err.response?.data || 'Failed to load system settings');
    } finally {
      setLoading(false);
    }
  };

  const handleSaveSystemSettings = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!systemSettings) return;

    setSaving(true);
    setError(null);
    setSuccessMessage(null);
    try {
      // Convert systemSettings back to flat map
      const flatSettings: Record<string, string> = {};
      Object.values(systemSettings.categories).flat().forEach(setting => {
        flatSettings[setting.key] = setting.value;
      });

      await api.updateSystemSettings(flatSettings);
      setSuccessMessage('System settings saved successfully');
      setTimeout(() => setSuccessMessage(null), 3000);
    } catch (err: any) {
      setError(err.response?.data || 'Failed to save system settings');
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
                onClick={() => setActiveTab('integration')}
                className={`py-4 px-1 border-b-2 font-medium text-sm transition-colors ${
                  activeTab === 'integration'
                    ? 'border-primary-500 text-primary-600'
                    : 'border-transparent text-slate-500 hover:text-slate-700 hover:border-slate-300'
                }`}
              >
                Integration & Performance
              </button>
            </nav>
          </div>

          {/* Messages */}
          {error && (
            <div className="mb-4 bg-error-50 border border-error-200 text-error-800 px-4 py-3 rounded">
              {error}
            </div>
          )}
          {successMessage && (
            <div className="mb-4 bg-success-50 border border-success-200 text-success-800 px-4 py-3 rounded">
              {successMessage}
            </div>
          )}

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
  activeTab: 'data-refresh' | 'detectors' | 'integration';
}

const SystemSettingsForm: React.FC<SystemSettingsFormProps> = ({
  settings,
  onSettingsChange,
  onSave,
  saving,
  activeTab,
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
      case 'integration':
        return ['integration', 'performance'];
      default:
        return [];
    }
  };

  const categoriesToShow = getCategoriesToShow();

  const categoryLabels: Record<string, string> = {
    data_refresh: 'Data Refresh Settings',
    detection: 'Detector Configuration',
    integration: 'M3 Integration Settings',
    performance: 'Performance & Caching',
    security: 'Security Settings',
  };

  return (
    <form onSubmit={onSave} className="bg-white shadow rounded-lg">
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

      {/* Actions */}
      <div className="px-6 py-4 bg-slate-50 flex justify-end gap-3">
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
