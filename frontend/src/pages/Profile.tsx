import React, { useState } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { AppLayout } from '../components/AppLayout';
import type { UserProfileGroup } from '../types';

const Profile: React.FC = () => {
  const { userProfile, environment, refreshProfile } = useAuth();
  const [refreshing, setRefreshing] = useState(false);
  const [expandedSections, setExpandedSections] = useState<Set<string>>(new Set());

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
      <div className="max-w-4xl mx-auto">
        <div className="bg-white shadow rounded-lg overflow-hidden">
          {/* Header */}
          <div className="px-6 py-5 border-b border-gray-200">
            <div className="flex justify-between items-center">
              <h1 className="text-2xl font-bold text-gray-900">User Profile</h1>
              <button
                onClick={handleRefresh}
                disabled={refreshing}
                className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {refreshing ? 'Refreshing...' : 'Refresh'}
              </button>
            </div>
          </div>

          {/* Basic Information */}
          <div className="px-6 py-5">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Basic Information</h2>
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

          {/* M3 Defaults & Preferences */}
          {userProfile.m3Info && (
            <div className="px-6 py-5 border-t border-gray-200">
              <h2 className="text-lg font-semibold text-gray-900 mb-4">M3 Defaults & Preferences</h2>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <ProfileField label="M3 User ID" value={userProfile.m3Info.userId} />
                <ProfileField label="Full Name" value={userProfile.m3Info.fullName} />
                <ProfileField label="Default Company" value={userProfile.m3Info.defaultCompany} />
                <ProfileField label="Default Division" value={userProfile.m3Info.defaultDivision} />
                <ProfileField label="Default Facility" value={userProfile.m3Info.defaultFacility} />
                <ProfileField label="Default Warehouse" value={userProfile.m3Info.defaultWarehouse} />
                <ProfileField label="Language" value={userProfile.m3Info.languageCode} />
                <ProfileField label="Date Format" value={userProfile.m3Info.dateFormat} />
                <ProfileField label="Time Zone" value={userProfile.m3Info.timeZone} />
              </div>
            </div>
          )}

          {/* Groups Section */}
          {userProfile.groups && userProfile.groups.length > 0 && (
            <div className="px-6 py-5 border-t border-gray-200">
              <h2 className="text-lg font-semibold text-gray-900 mb-4">
                Groups & Roles ({userProfile.groups.length})
              </h2>
              <div className="space-y-3">
                {Object.entries(groupsByType).map(([type, groups]) => (
                  <div key={type} className="border border-gray-200 rounded">
                    <button
                      onClick={() => toggleSection(type)}
                      className="w-full px-4 py-3 flex justify-between items-center bg-gray-50 hover:bg-gray-100"
                    >
                      <span className="font-medium text-gray-900">
                        {type} ({groups.length})
                      </span>
                      <svg
                        className={`w-5 h-5 text-gray-500 transform transition-transform ${
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
                          {groups.map((group) => (
                            <li key={group.value} className="text-sm text-gray-700">
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
          )}
        </div>
      </div>
      </div>
    </AppLayout>
  );
};

const ProfileField: React.FC<{ label: string; value?: string }> = ({ label, value }) => (
  <div>
    <dt className="text-sm font-medium text-gray-500">{label}</dt>
    <dd className="mt-1 text-sm text-gray-900">{value || 'N/A'}</dd>
  </div>
);

export default Profile;
