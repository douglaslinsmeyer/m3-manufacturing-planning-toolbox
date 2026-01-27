import React from 'react';
import type { SystemSetting } from '../types';
import HierarchicalThresholdInput from './HierarchicalThresholdInput';
import FilterInput from './FilterInput';

interface DetectorSectionProps {
  detectorName: string;
  detectorLabel: string;
  detectorDescription: string;
  settings: SystemSetting[];
  onSettingsChange: (updated: SystemSetting[]) => void;
}

const DetectorSection: React.FC<DetectorSectionProps> = ({
  detectorName,
  detectorLabel,
  detectorDescription,
  settings,
  onSettingsChange,
}) => {
  // Filter settings for this specific detector
  const detectorSettings = settings.filter(s =>
    s.key.startsWith(`detector_${detectorName}_`)
  );

  // Separate enabled toggle from other settings
  const enabledSetting = detectorSettings.find(s => s.key.endsWith('_enabled'));

  // Identify hierarchical threshold settings (have "hierarchical": true in constraints)
  const hierarchicalSettings = detectorSettings.filter(s => {
    if (s.type !== 'json') return false;
    return s.constraints?.hierarchical === true;
  });

  // Identify filter settings (non-hierarchical, non-enabled)
  const filterSettings = detectorSettings.filter(s => {
    if (s.key.endsWith('_enabled')) return false;
    if (s.constraints?.hierarchical === true) return false;
    return true;
  });

  const updateSetting = (key: string, value: string) => {
    const updated = settings.map(s =>
      s.key === key ? { ...s, value } : s
    );
    onSettingsChange(updated);
  };

  return (
    <div className="bg-white shadow rounded-lg overflow-hidden">
      {/* Header with Toggle */}
      <div className="px-6 py-4 bg-slate-50 border-b border-slate-200">
        <div className="flex items-center justify-between">
          <div className="flex-1">
            <h3 className="text-lg font-semibold text-slate-900">{detectorLabel}</h3>
            <p className="text-sm text-slate-600 mt-1">{detectorDescription}</p>
          </div>
          {enabledSetting && (
            <div className="flex items-center gap-3">
              <span className="text-sm text-slate-700">Detector Status:</span>
              <label className="relative inline-flex items-center cursor-pointer">
                <input
                  type="checkbox"
                  checked={enabledSetting.value === 'true'}
                  onChange={(e) => updateSetting(enabledSetting.key, e.target.checked ? 'true' : 'false')}
                  className="sr-only peer"
                />
                <div className="w-11 h-6 bg-slate-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-slate-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary-600"></div>
                <span className="ml-3 text-sm font-medium text-slate-900">
                  {enabledSetting.value === 'true' ? 'Enabled' : 'Disabled'}
                </span>
              </label>
            </div>
          )}
        </div>
      </div>

      {/* Configuration Body */}
      <div className="px-6 py-5 space-y-6">
        {/* Hierarchical Thresholds Section */}
        {hierarchicalSettings.length > 0 && (
          <div>
            <h4 className="text-sm font-semibold text-slate-700 mb-3 uppercase tracking-wide">
              Hierarchical Thresholds
            </h4>
            <p className="text-xs text-slate-500 mb-3">
              Configure global defaults and scope-specific overrides (warehouse, facility, MO type)
            </p>
            <div className="space-y-4">
              {hierarchicalSettings.map(setting => (
                <HierarchicalThresholdInput
                  key={setting.key}
                  setting={setting}
                  onChange={(value) => updateSetting(setting.key, value)}
                />
              ))}
            </div>
          </div>
        )}

        {/* Global Filters Section */}
        {filterSettings.length > 0 && (
          <div>
            <h4 className="text-sm font-semibold text-slate-700 mb-3 uppercase tracking-wide">
              Global Filters
            </h4>
            <p className="text-xs text-slate-500 mb-3">
              Apply these filters system-wide across all contexts
            </p>
            <div className="space-y-4">
              {filterSettings.map(setting => (
                <FilterInput
                  key={setting.key}
                  setting={setting}
                  onChange={(value) => updateSetting(setting.key, value)}
                />
              ))}
            </div>
          </div>
        )}

        {/* Empty State */}
        {hierarchicalSettings.length === 0 && filterSettings.length === 0 && (
          <div className="text-center py-8 text-slate-500">
            No configurable parameters for this detector
          </div>
        )}
      </div>
    </div>
  );
};

export default DetectorSection;
