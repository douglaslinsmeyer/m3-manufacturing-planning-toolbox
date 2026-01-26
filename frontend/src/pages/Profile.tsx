import React, { useState, useEffect } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { useContextManagement } from '../contexts/ContextManagementContext';
import { AppLayout } from '../components/AppLayout';
import { api } from '../services/api';
import type { UserProfileGroup, UserSettings } from '../types';

const Profile: React.FC = () => {
  const { userProfile, environment, refreshProfile } = useAuth();
  const [activeTab, setActiveTab] = useState<'basic' | 'm3' | 'groups' | 'settings'>('basic');
  const [refreshing, setRefreshing] = useState(false);
  const [expandedSections, setExpandedSections] = useState<Set<string>>(new Set());

  // User settings state
  const [userSettings, setUserSettings] = useState<UserSettings | null>(null);
  const [loadingSettings, setLoadingSettings] = useState(false);
  const [saving, setSaving] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);
  const [saveSuccess, setSaveSuccess] = useState<string | null>(null);

  const handleRefresh = async () => {
    setRefreshing(true);
    try {
      await refreshProfile();
    } catch (error) {
      console.error('Failed to refresh profile:', error);
    } finally {
      setRefreshing(false);
    }
  };

  const toggleSection = (sectionType: string) => {
    const newExpanded = new Set(expandedSections);
    if (newExpanded.has(sectionType)) {
      newExpanded.delete(sectionType);
    } else {
      newExpanded.add(sectionType);
    }
    setExpandedSections(newExpanded);
  };

  // Load user settings when switching to My Settings tab
  useEffect(() => {
    if (activeTab === 'settings') {
      loadUserSettings();
    }
  }, [activeTab]);

  const loadUserSettings = async () => {
    setLoadingSettings(true);
    setSaveError(null);
    try {
      const settings = await api.getUserSettings();
      setUserSettings(settings);
    } catch (err: any) {
      setSaveError(err.response?.data || 'Failed to load settings');
    } finally {
      setLoadingSettings(false);
    }
  };

  const handleSaveUserSettings = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!userSettings) return;

    setSaving(true);
    setSaveError(null);
    setSaveSuccess(null);
    try {
      await api.updateUserSettings(userSettings);
      setSaveSuccess('Settings saved successfully');
      setTimeout(() => setSaveSuccess(null), 3000);
    } catch (err: any) {
      setSaveError(err.response?.data || 'Failed to save settings');
    } finally {
      setSaving(false);
    }
  };

  if (!userProfile) {
    return (
      <AppLayout>
        <div className="flex items-center justify-center h-96">
          <div className="flex flex-col items-center gap-4">
            <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary-200 border-t-primary-600" />
            <p className="text-sm text-slate-500">Loading profile...</p>
          </div>
        </div>
      </AppLayout>
    );
  }

  // Group the groups by type
  const groupsByType = userProfile.groups?.reduce((acc, group) => {
    if (!acc[group.type]) {
      acc[group.type] = [];
    }
    acc[group.type].push(group);
    return acc;
  }, {} as Record<string, UserProfileGroup[]>) || {};

  return (
    <AppLayout>
      <div className="px-4 py-6 sm:px-6 lg:px-12 lg:py-10">
        <div className="max-w-5xl mx-auto">
          {/* Header with Refresh Button */}
          <div className="mb-6 flex justify-between items-start">
            <div>
              <h1 className="text-3xl font-bold text-slate-900">Profile</h1>
              <p className="mt-2 text-sm text-slate-600">
                View your user information and role assignments
              </p>
            </div>
            <button
              onClick={handleRefresh}
              disabled={refreshing}
              className="px-6 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed font-medium transition-colors"
            >
              {refreshing ? 'Refreshing...' : 'Refresh Profile'}
            </button>
          </div>

          {/* Tabs */}
          <div className="border-b border-slate-200 mb-6">
            <nav className="-mb-px flex space-x-8">
              <button
                onClick={() => setActiveTab('basic')}
                className={`py-4 px-1 border-b-2 font-medium text-sm transition-colors ${
                  activeTab === 'basic'
                    ? 'border-primary-500 text-primary-600'
                    : 'border-transparent text-slate-500 hover:text-slate-700 hover:border-slate-300'
                }`}
              >
                Basic Info
              </button>
              <button
                onClick={() => setActiveTab('m3')}
                className={`py-4 px-1 border-b-2 font-medium text-sm transition-colors ${
                  activeTab === 'm3'
                    ? 'border-primary-500 text-primary-600'
                    : 'border-transparent text-slate-500 hover:text-slate-700 hover:border-slate-300'
                }`}
              >
                M3 Defaults
              </button>
              <button
                onClick={() => setActiveTab('groups')}
                className={`py-4 px-1 border-b-2 font-medium text-sm transition-colors ${
                  activeTab === 'groups'
                    ? 'border-primary-500 text-primary-600'
                    : 'border-transparent text-slate-500 hover:text-slate-700 hover:border-slate-300'
                }`}
              >
                Groups & Roles
                {userProfile.groups && userProfile.groups.length > 0 && (
                  <span className="ml-2 px-2 py-0.5 text-xs font-semibold bg-slate-100 text-slate-700 rounded">
                    {userProfile.groups.length}
                  </span>
                )}
              </button>
              <button
                onClick={() => setActiveTab('settings')}
                className={`py-4 px-1 border-b-2 font-medium text-sm transition-colors ${
                  activeTab === 'settings'
                    ? 'border-primary-500 text-primary-600'
                    : 'border-transparent text-slate-500 hover:text-slate-700 hover:border-slate-300'
                }`}
              >
                My Settings
              </button>
            </nav>
          </div>

          {/* Content */}
          {activeTab === 'basic' && (
            <BasicInfoTab
              userProfile={userProfile}
              environment={environment}
            />
          )}
          {activeTab === 'm3' && (
            <M3DefaultsTab m3Info={userProfile.m3Info} />
          )}
          {activeTab === 'groups' && (
            <GroupsRolesTab
              groups={userProfile.groups}
              groupsByType={groupsByType}
              expandedSections={expandedSections}
              toggleSection={toggleSection}
            />
          )}
          {activeTab === 'settings' && (
            loadingSettings ? (
              <div className="flex items-center justify-center h-64">
                <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary-200 border-t-primary-600" />
              </div>
            ) : (
              <>
                {/* Messages */}
                {saveError && (
                  <div className="mb-4 bg-error-50 border border-error-200 text-error-800 px-4 py-3 rounded">
                    {saveError}
                  </div>
                )}
                {saveSuccess && (
                  <div className="mb-4 bg-success-50 border border-success-200 text-success-800 px-4 py-3 rounded">
                    {saveSuccess}
                  </div>
                )}
                <MySettingsTab
                  settings={userSettings}
                  onSettingsChange={setUserSettings}
                  onSave={handleSaveUserSettings}
                  saving={saving}
                />
              </>
            )
          )}
        </div>
      </div>
    </AppLayout>
  );
};

// Basic Info Tab Component
interface BasicInfoTabProps {
  userProfile: any;
  environment?: 'TRN' | 'PRD';
}

const BasicInfoTab: React.FC<BasicInfoTabProps> = ({ userProfile, environment }) => (
  <div className="bg-white shadow rounded-lg">
    <div className="px-6 py-5">
      <h3 className="text-lg font-semibold text-slate-900 mb-4">Basic Information</h3>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <ProfileField label="Display Name" value={userProfile.displayName} />
        <ProfileField label="Username" value={userProfile.userName} />
        <ProfileField label="Email" value={userProfile.email} />
        <ProfileField label="Title" value={userProfile.title} />
        <ProfileField label="Department" value={userProfile.department} />
        <ProfileField label="User ID" value={userProfile.id} />
        <ProfileField label="Environment" value={environment} />
      </div>
    </div>
  </div>
);

// M3 Defaults Tab Component
interface M3DefaultsTabProps {
  m3Info?: any;
}

const M3DefaultsTab: React.FC<M3DefaultsTabProps> = ({ m3Info }) => {
  if (!m3Info) {
    return (
      <div className="bg-white shadow rounded-lg">
        <div className="px-6 py-12 text-center">
          <p className="text-slate-500">No M3 information available</p>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white shadow rounded-lg">
      <div className="px-6 py-5">
        <h3 className="text-lg font-semibold text-slate-900 mb-4">M3 Defaults & Preferences</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <ProfileField label="M3 User ID" value={m3Info.userId} />
          <ProfileField label="Full Name" value={m3Info.fullName} />
          <ProfileField label="Default Company" value={m3Info.defaultCompany} />
          <ProfileField label="Default Division" value={m3Info.defaultDivision} />
          <ProfileField label="Default Facility" value={m3Info.defaultFacility} />
          <ProfileField label="Default Warehouse" value={m3Info.defaultWarehouse} />
          <ProfileField label="Language" value={m3Info.languageCode} />
          <ProfileField label="Date Format" value={m3Info.dateFormat} />
          <ProfileField label="Time Zone" value={m3Info.timeZone} />
        </div>
      </div>
    </div>
  );
};

// Groups & Roles Tab Component
interface GroupsRolesTabProps {
  groups?: UserProfileGroup[];
  groupsByType: Record<string, UserProfileGroup[]>;
  expandedSections: Set<string>;
  toggleSection: (type: string) => void;
}

const GroupsRolesTab: React.FC<GroupsRolesTabProps> = ({
  groups,
  groupsByType,
  expandedSections,
  toggleSection,
}) => {
  if (!groups || groups.length === 0) {
    return (
      <div className="bg-white shadow rounded-lg">
        <div className="px-6 py-12 text-center">
          <p className="text-slate-500">No groups or roles assigned</p>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white shadow rounded-lg">
      <div className="px-6 py-5">
        <h3 className="text-lg font-semibold text-slate-900 mb-4">
          Groups & Roles ({groups.length})
        </h3>
        <div className="space-y-3">
          {Object.entries(groupsByType).map(([type, groupList]) => (
            <div key={type} className="border border-slate-200 rounded">
              <button
                onClick={() => toggleSection(type)}
                className="w-full px-4 py-3 flex justify-between items-center bg-slate-50 hover:bg-slate-100 transition-colors"
              >
                <span className="font-medium text-slate-900">
                  {type} ({groupList.length})
                </span>
                <svg
                  className={`w-5 h-5 text-slate-500 transform transition-transform ${
                    expandedSections.has(type) ? 'rotate-180' : ''
                  }`}
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                </svg>
              </button>
              {expandedSections.has(type) && (
                <div className="px-4 py-3 bg-white">
                  <ul className="space-y-2">
                    {groupList.map((group) => (
                      <li key={group.value} className="text-sm text-slate-700">
                        {group.display}
                      </li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

// My Settings Tab Component
interface MySettingsTabProps {
  settings: UserSettings | null;
  onSettingsChange: (settings: UserSettings | null) => void;
  onSave: (e: React.FormEvent) => void;
  saving: boolean;
}

const MySettingsTab: React.FC<MySettingsTabProps> = ({
  settings,
  onSettingsChange,
  onSave,
  saving,
}) => {
  const {
    companies,
    divisions,
    facilities,
    warehouses,
    loadCompanies,
    loadDivisions,
    loadFacilities,
    loadWarehouses,
  } = useContextManagement();

  const [loadingDivisions, setLoadingDivisions] = useState(false);
  const [loadingWarehouses, setLoadingWarehouses] = useState(false);

  if (!settings) return null;

  const updateSetting = (key: keyof UserSettings, value: any) => {
    onSettingsChange({ ...settings, [key]: value });
  };

  // Load initial data on mount
  useEffect(() => {
    loadCompanies();
    loadFacilities();

    if (settings?.defaultCompany) {
      loadDivisions(settings.defaultCompany);
      loadWarehouses(
        settings.defaultCompany,
        settings.defaultDivision,
        settings.defaultFacility
      );
    }
  }, []);

  // Handle company change with cascading logic
  const handleCompanyChange = async (newCompany: string) => {
    updateSetting('defaultCompany', newCompany);

    // Clear dependent fields
    updateSetting('defaultDivision', '');
    updateSetting('defaultWarehouse', '');

    // Reload divisions and warehouses for new company
    if (newCompany) {
      setLoadingDivisions(true);
      setLoadingWarehouses(true);
      try {
        await loadDivisions(newCompany);
        await loadWarehouses(newCompany);
      } finally {
        setLoadingDivisions(false);
        setLoadingWarehouses(false);
      }
    }
  };

  // Handle division change - reload warehouses with filter
  const handleDivisionChange = async (newDivision: string) => {
    updateSetting('defaultDivision', newDivision);

    if (settings.defaultCompany) {
      setLoadingWarehouses(true);
      try {
        await loadWarehouses(
          settings.defaultCompany,
          newDivision,
          settings.defaultFacility
        );
      } finally {
        setLoadingWarehouses(false);
      }
    }
  };

  // Handle facility change - reload warehouses with filter
  const handleFacilityChange = async (newFacility: string) => {
    updateSetting('defaultFacility', newFacility);

    if (settings.defaultCompany) {
      setLoadingWarehouses(true);
      try {
        await loadWarehouses(
          settings.defaultCompany,
          settings.defaultDivision,
          newFacility
        );
      } finally {
        setLoadingWarehouses(false);
      }
    }
  };

  return (
    <form onSubmit={onSave} className="bg-white shadow rounded-lg">
      {/* Default Context */}
      <div className="px-6 py-5 border-b border-slate-200">
        <h3 className="text-lg font-semibold text-slate-900 mb-4">Default Context</h3>
        <p className="text-sm text-slate-600 mb-4">
          Set your preferred default warehouse and facility. These will be used when you log in.
        </p>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {/* Company Dropdown */}
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">
              Default Company
            </label>
            <select
              value={settings.defaultCompany || ''}
              onChange={(e) => handleCompanyChange(e.target.value)}
              className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent bg-white"
            >
              <option value="">-- Select Company --</option>
              {companies.map(c => (
                <option key={c.companyNumber} value={c.companyNumber}>
                  {c.companyNumber} - {c.companyName}
                </option>
              ))}
            </select>
          </div>

          {/* Division Dropdown */}
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">
              Default Division
            </label>
            <select
              value={settings.defaultDivision || ''}
              onChange={(e) => handleDivisionChange(e.target.value)}
              disabled={!settings.defaultCompany || loadingDivisions}
              className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent bg-white disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <option value="">-- Select Division --</option>
              {divisions.map(d => (
                <option key={d.division} value={d.division}>
                  {d.division} - {d.divisionName}
                </option>
              ))}
            </select>
            {!settings.defaultCompany && (
              <p className="mt-1 text-xs text-slate-500">Select a company first</p>
            )}
          </div>

          {/* Facility Dropdown */}
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">
              Default Facility
            </label>
            <select
              value={settings.defaultFacility || ''}
              onChange={(e) => handleFacilityChange(e.target.value)}
              className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent bg-white"
            >
              <option value="">-- Select Facility --</option>
              {facilities.map(f => (
                <option key={f.facility} value={f.facility}>
                  {f.facility} - {f.facilityName}
                </option>
              ))}
            </select>
          </div>

          {/* Warehouse Dropdown */}
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">
              Default Warehouse
            </label>
            <select
              value={settings.defaultWarehouse || ''}
              onChange={(e) => updateSetting('defaultWarehouse', e.target.value)}
              disabled={!settings.defaultCompany || loadingWarehouses}
              className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent bg-white disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <option value="">-- Select Warehouse --</option>
              {warehouses.map(w => (
                <option key={w.warehouse} value={w.warehouse}>
                  {w.warehouse} - {w.warehouseName}
                </option>
              ))}
            </select>
            {!settings.defaultCompany && (
              <p className="mt-1 text-xs text-slate-500">Select a company first</p>
            )}
          </div>
        </div>
      </div>

      {/* Display Preferences */}
      <div className="px-6 py-5 border-b border-slate-200">
        <h3 className="text-lg font-semibold text-slate-900 mb-4">Display Preferences</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">
              Items per Page
            </label>
            <input
              type="number"
              min="1"
              max="200"
              value={settings.itemsPerPage}
              onChange={(e) => updateSetting('itemsPerPage', parseInt(e.target.value))}
              className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">Theme</label>
            <select
              value={settings.theme}
              onChange={(e) => updateSetting('theme', e.target.value)}
              className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
            >
              <option value="light">Light</option>
              <option value="dark">Dark</option>
              <option value="auto">Auto (System)</option>
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">
              Date Format
            </label>
            <select
              value={settings.dateFormat}
              onChange={(e) => updateSetting('dateFormat', e.target.value)}
              className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
            >
              <option value="YYYY-MM-DD">YYYY-MM-DD (2026-01-25)</option>
              <option value="MM/DD/YYYY">MM/DD/YYYY (01/25/2026)</option>
              <option value="DD/MM/YYYY">DD/MM/YYYY (25/01/2026)</option>
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">
              Time Format
            </label>
            <select
              value={settings.timeFormat}
              onChange={(e) => updateSetting('timeFormat', e.target.value)}
              className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
            >
              <option value="12h">12-hour (3:30 PM)</option>
              <option value="24h">24-hour (15:30)</option>
            </select>
          </div>
        </div>
      </div>

      {/* Notification Preferences */}
      <div className="px-6 py-5 border-b border-slate-200">
        <h3 className="text-lg font-semibold text-slate-900 mb-4">Notifications</h3>
        <div className="space-y-3">
          <label className="flex items-center gap-3">
            <input
              type="checkbox"
              checked={settings.enableNotifications}
              onChange={(e) => updateSetting('enableNotifications', e.target.checked)}
              className="h-4 w-4 text-primary-600 border-slate-300 rounded focus:ring-primary-500"
            />
            <span className="text-sm text-slate-700">Enable notifications</span>
          </label>
          <label className="flex items-center gap-3">
            <input
              type="checkbox"
              checked={settings.notificationSound}
              onChange={(e) => updateSetting('notificationSound', e.target.checked)}
              disabled={!settings.enableNotifications}
              className="h-4 w-4 text-primary-600 border-slate-300 rounded focus:ring-primary-500 disabled:opacity-50"
            />
            <span className="text-sm text-slate-700">Notification sound</span>
          </label>
        </div>
      </div>

      {/* Actions */}
      <div className="px-6 py-4 bg-slate-50 flex justify-end gap-3">
        <button
          type="submit"
          disabled={saving}
          className="px-6 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed font-medium transition-colors"
        >
          {saving ? 'Saving...' : 'Save Changes'}
        </button>
      </div>
    </form>
  );
};

// Profile Field Component (unchanged)
const ProfileField: React.FC<{ label: string; value?: string }> = ({ label, value }) => (
  <div>
    <dt className="text-sm font-medium text-slate-500">{label}</dt>
    <dd className="mt-1 text-sm text-slate-900">{value || 'N/A'}</dd>
  </div>
);

export default Profile;
