import React, { useState } from 'react';
import type { SystemSetting } from '../types';

interface HierarchicalThresholdInputProps {
  setting: SystemSetting;
  onChange: (value: string) => void;
}

const HierarchicalThresholdInput: React.FC<HierarchicalThresholdInputProps> = ({
  setting,
  onChange,
}) => {
  const [expanded, setExpanded] = useState(false);

  // Parse current JSON value
  let parsedValue: { global: number; overrides: any[] };
  try {
    parsedValue = JSON.parse(setting.value);
  } catch (e) {
    console.error('Failed to parse hierarchical threshold JSON:', e);
    parsedValue = { global: 0, overrides: [] };
  }

  const global = parsedValue.global || 0;
  const overrides = parsedValue.overrides || [];

  const updateGlobal = (newGlobal: number) => {
    const updated = { ...parsedValue, global: newGlobal };
    onChange(JSON.stringify(updated));
  };

  const addOverride = () => {
    const newOverride = { value: global };
    const updated = { ...parsedValue, overrides: [...overrides, newOverride] };
    onChange(JSON.stringify(updated));
    setExpanded(true);
  };

  const updateOverride = (index: number, field: string, value: any) => {
    const updatedOverrides = [...overrides];
    if (value === '' || value === undefined) {
      delete updatedOverrides[index][field];
    } else {
      updatedOverrides[index][field] = value;
    }
    const updated = { ...parsedValue, overrides: updatedOverrides };
    onChange(JSON.stringify(updated));
  };

  const deleteOverride = (index: number) => {
    const updatedOverrides = overrides.filter((_: any, i: number) => i !== index);
    const updated = { ...parsedValue, overrides: updatedOverrides };
    onChange(JSON.stringify(updated));
  };

  const getScopeLabel = (override: any): string => {
    const parts = [];
    if (override.warehouse) parts.push(`WH: ${override.warehouse}`);
    if (override.facility) parts.push(`Fac: ${override.facility}`);
    if (override.moType) parts.push(`MO: ${override.moType}`);
    return parts.length > 0 ? parts.join(' + ') : 'No scope';
  };

  const getSpecificityScore = (override: any): number => {
    let score = 0;
    if (override.warehouse) score += 4;
    if (override.facility) score += 2;
    if (override.moType) score += 1;
    return score;
  };

  return (
    <div className="border border-slate-200 rounded-lg">
      {/* Global Default */}
      <div className="px-4 py-3 bg-slate-50 flex items-center justify-between">
        <div className="flex-1">
          <label className="text-sm font-medium text-slate-900">
            {setting.description || setting.key}
          </label>
          <p className="text-xs text-slate-500 mt-0.5">Global default value (applies when no override matches)</p>
        </div>
        <div className="flex items-center gap-3">
          <input
            type="number"
            value={global}
            onChange={(e) => updateGlobal(parseFloat(e.target.value) || 0)}
            min={setting.constraints?.min || 0}
            max={setting.constraints?.max}
            className="w-24 px-3 py-2 border border-slate-300 rounded-lg text-right font-mono focus:ring-2 focus:ring-primary-500"
          />
          <span className="text-sm text-slate-600 min-w-[60px]">{setting.constraints?.unit || 'units'}</span>
        </div>
      </div>

      {/* Overrides Section */}
      {overrides.length > 0 && (
        <div className="border-t border-slate-200">
          <button
            type="button"
            onClick={() => setExpanded(!expanded)}
            className="w-full px-4 py-2 flex items-center justify-between text-sm font-medium text-slate-700 hover:bg-slate-50 transition-colors"
          >
            <span>
              {overrides.length} scope override{overrides.length !== 1 ? 's' : ''}
            </span>
            <span className="text-slate-400">{expanded ? '▼' : '▶'}</span>
          </button>

          {expanded && (
            <div className="px-4 py-3 space-y-3 bg-slate-25">
              {overrides
                .map((override: any, index: number) => ({ override, index, score: getSpecificityScore(override) }))
                .sort((a, b) => b.score - a.score) // Sort by specificity (most specific first)
                .map(({ override, index, score }) => (
                  <div key={index} className="flex items-center gap-3 bg-white p-3 rounded-lg border border-slate-200 shadow-sm">
                    <div className="flex-1 grid grid-cols-3 gap-2">
                      <input
                        type="text"
                        placeholder="Warehouse"
                        value={override.warehouse || ''}
                        onChange={(e) => updateOverride(index, 'warehouse', e.target.value)}
                        className="px-2 py-1.5 text-sm border border-slate-300 rounded focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                      />
                      <input
                        type="text"
                        placeholder="Facility"
                        value={override.facility || ''}
                        onChange={(e) => updateOverride(index, 'facility', e.target.value)}
                        className="px-2 py-1.5 text-sm border border-slate-300 rounded focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                      />
                      <input
                        type="text"
                        placeholder="MO Type"
                        value={override.moType || ''}
                        onChange={(e) => updateOverride(index, 'moType', e.target.value)}
                        className="px-2 py-1.5 text-sm border border-slate-300 rounded focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                      />
                    </div>
                    <div className="flex items-center gap-2">
                      <input
                        type="number"
                        value={override.value || 0}
                        onChange={(e) => updateOverride(index, 'value', parseFloat(e.target.value) || 0)}
                        min={setting.constraints?.min || 0}
                        max={setting.constraints?.max}
                        className="w-20 px-2 py-1.5 text-sm border border-slate-300 rounded text-right font-mono focus:ring-2 focus:ring-primary-500"
                      />
                      <span className="px-2 py-1 text-xs font-semibold rounded bg-slate-100 text-slate-600">
                        P{7 - Math.min(score, 7)}
                      </span>
                    </div>
                    <button
                      type="button"
                      onClick={() => deleteOverride(index)}
                      className="px-2 py-1 text-xs text-error-600 hover:text-error-800 font-medium hover:bg-error-50 rounded transition-colors"
                    >
                      Delete
                    </button>
                  </div>
                ))}
            </div>
          )}
        </div>
      )}

      {/* Add Override Button */}
      <div className="px-4 py-2 border-t border-slate-200">
        <button
          type="button"
          onClick={addOverride}
          className="text-sm text-primary-600 hover:text-primary-800 font-medium transition-colors"
        >
          + Add Scope Override
        </button>
      </div>
    </div>
  );
};

export default HierarchicalThresholdInput;
