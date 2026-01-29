import React, { useEffect, useState } from 'react';
import { api } from '../../services/api';
import { useContextManagement } from '../../contexts/ContextManagementContext';
import { IssueCriteria } from '../../pages/BulkOperations';

interface FilterBuilderProps {
  filters: IssueCriteria;
  onChange: (filters: IssueCriteria) => void;
  onPreview: () => void;
  isLoading: boolean;
}

interface Detector {
  name: string;
  label: string;
  description: string;
}

export const FilterBuilder: React.FC<FilterBuilderProps> = ({
  filters,
  onChange,
  onPreview,
  isLoading,
}) => {
  const { effectiveContext, warehouses, loadWarehouses } = useContextManagement();
  const [detectors, setDetectors] = useState<Detector[]>([]);

  // Load detectors and warehouses on mount
  useEffect(() => {
    loadDetectors();

    // Load warehouses if we have effective context
    if (effectiveContext) {
      loadWarehouses(
        effectiveContext.company,
        effectiveContext.division,
        effectiveContext.facility
      );
    }
  }, [effectiveContext, loadWarehouses]);

  const loadDetectors = async () => {
    try {
      const response = await api.get('/detection/detectors');
      // Backend returns array directly, filter to only show enabled detectors
      const enabledDetectors = (response.data || []).filter((d: any) => d.enabled);
      setDetectors(enabledDetectors);
    } catch (error) {
      console.error('Failed to load detectors:', error);
    }
  };

  return (
    <div className="bg-white rounded-lg shadow p-6">
      <h2 className="text-lg font-semibold mb-4 text-gray-900">Filter Criteria</h2>

      {/* Detector Type Dropdown */}
      <div className="mb-4">
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Issue Type
        </label>
        <select
          value={filters.detector_type}
          onChange={(e) => onChange({ ...filters, detector_type: e.target.value })}
          className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
        >
          <option value="">All Types</option>
          {detectors.map((detector) => (
            <option key={detector.name} value={detector.name}>
              {detector.label}
            </option>
          ))}
        </select>
      </div>

      {/* Warehouse Dropdown */}
      <div className="mb-4">
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Warehouse
        </label>
        <select
          value={filters.warehouse}
          onChange={(e) => onChange({ ...filters, warehouse: e.target.value })}
          className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
        >
          <option value="">All Warehouses</option>
          {warehouses.map((wh) => (
            <option key={wh.warehouse} value={wh.warehouse}>
              {wh.warehouse} - {wh.warehouseName}
            </option>
          ))}
        </select>
      </div>

      {/* Product Number Text Input */}
      <div className="mb-4">
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Product Number (PRNO)
        </label>
        <input
          type="text"
          value={filters.product_number}
          onChange={(e) => onChange({ ...filters, product_number: e.target.value })}
          placeholder="e.g., G440D"
          className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
        />
        <p className="mt-1 text-xs text-gray-500">
          Filter by specific product number
        </p>
      </div>

      {/* Include Ignored Checkbox */}
      <div className="mb-6">
        <label className="flex items-center">
          <input
            type="checkbox"
            checked={filters.include_ignored}
            onChange={(e) => onChange({ ...filters, include_ignored: e.target.checked })}
            className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
          />
          <span className="ml-2 text-sm text-gray-700">Include ignored issues</span>
        </label>
      </div>

      {/* Preview Button */}
      <button
        onClick={onPreview}
        disabled={isLoading}
        className={`w-full px-4 py-2 rounded-md font-medium transition-colors ${
          isLoading
            ? 'bg-gray-300 text-gray-500 cursor-not-allowed'
            : 'bg-blue-600 text-white hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2'
        }`}
      >
        {isLoading ? (
          <span className="flex items-center justify-center">
            <svg
              className="animate-spin -ml-1 mr-2 h-4 w-4 text-white"
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
            >
              <circle
                className="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                strokeWidth="4"
              ></circle>
              <path
                className="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              ></path>
            </svg>
            Loading...
          </span>
        ) : (
          'Preview Matching Issues'
        )}
      </button>

      {/* Help Text */}
      <div className="mt-4 p-3 bg-blue-50 rounded-md">
        <p className="text-xs text-blue-800">
          <strong>Tip:</strong> Set filters above and click "Preview" to see how many
          issues match your criteria before executing any operations.
        </p>
      </div>
    </div>
  );
};
