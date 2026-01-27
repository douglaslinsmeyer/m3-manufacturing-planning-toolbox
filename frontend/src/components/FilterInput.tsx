import React from 'react';
import type { SystemSetting } from '../types';

interface FilterInputProps {
  setting: SystemSetting;
  onChange: (value: string) => void;
}

const FilterInput: React.FC<FilterInputProps> = ({ setting, onChange }) => {
  // Handle JSON array type (status codes, facilities)
  if (setting.type === 'json') {
    let arrayValue: string[];
    try {
      arrayValue = JSON.parse(setting.value) as string[];
    } catch (e) {
      console.error('Failed to parse JSON array:', e);
      arrayValue = [];
    }

    const displayValue = arrayValue.join(', ');

    return (
      <div>
        <label className="block text-sm font-medium text-slate-700 mb-1">
          {setting.description || setting.key}
        </label>
        <input
          type="text"
          value={displayValue}
          onChange={(e) => {
            const array = e.target.value
              .split(',')
              .map(s => s.trim())
              .filter(s => s.length > 0);
            onChange(JSON.stringify(array));
          }}
          className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
          placeholder="10, 20, 90"
        />
        <p className="text-xs text-slate-500 mt-1">
          Enter comma-separated values (e.g., status codes: "10, 20, 90")
        </p>
      </div>
    );
  }

  // Handle integer type
  if (setting.type === 'integer') {
    return (
      <div>
        <label className="block text-sm font-medium text-slate-700 mb-1">
          {setting.description || setting.key}
          {setting.constraints?.unit && (
            <span className="ml-1 text-slate-500">({setting.constraints.unit})</span>
          )}
        </label>
        <input
          type="number"
          value={setting.value}
          onChange={(e) => onChange(e.target.value)}
          min={setting.constraints?.min}
          max={setting.constraints?.max}
          step="1"
          className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
        />
        {setting.constraints && (setting.constraints.min !== undefined || setting.constraints.max !== undefined) && (
          <p className="text-xs text-slate-500 mt-1">
            Range: {setting.constraints.min || '−∞'} to {setting.constraints.max || '∞'}
            {setting.constraints.unit && ` ${setting.constraints.unit}`}
          </p>
        )}
      </div>
    );
  }

  // Handle float type
  if (setting.type === 'float') {
    return (
      <div>
        <label className="block text-sm font-medium text-slate-700 mb-1">
          {setting.description || setting.key}
          {setting.constraints?.unit && (
            <span className="ml-1 text-slate-500">({setting.constraints.unit})</span>
          )}
        </label>
        <input
          type="number"
          value={setting.value}
          onChange={(e) => onChange(e.target.value)}
          min={setting.constraints?.min}
          max={setting.constraints?.max}
          step="0.01"
          className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
        />
        {setting.constraints && (setting.constraints.min !== undefined || setting.constraints.max !== undefined) && (
          <p className="text-xs text-slate-500 mt-1">
            Range: {setting.constraints.min || '−∞'} to {setting.constraints.max || '∞'}
            {setting.constraints.unit && ` ${setting.constraints.unit}`}
          </p>
        )}
      </div>
    );
  }

  // Skip boolean (rendered as toggle in DetectorSection header)
  return null;
};

export default FilterInput;
