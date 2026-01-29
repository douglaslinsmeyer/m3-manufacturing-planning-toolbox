import React from 'react';
import { PreviewData } from '../../pages/BulkOperations';
import { Trash2, XCircle, Calendar } from 'lucide-react';

interface OperationPanelProps {
  previewData: PreviewData | null;
  selectedOperation: string;
  onOperationChange: (operation: string) => void;
  operationParams: Record<string, any>;
  onParamsChange: (params: Record<string, any>) => void;
  onExecute: () => void;
}

export const OperationPanel: React.FC<OperationPanelProps> = ({
  previewData,
  selectedOperation,
  onOperationChange,
  operationParams,
  onParamsChange,
  onExecute,
}) => {
  const canExecute =
    previewData &&
    previewData.total_count > 0 &&
    selectedOperation &&
    (selectedOperation !== 'reschedule' || operationParams.new_date);

  const getOperationIcon = (op: string) => {
    switch (op) {
      case 'delete':
        return <Trash2 className="h-4 w-4" />;
      case 'close':
        return <XCircle className="h-4 w-4" />;
      case 'reschedule':
        return <Calendar className="h-4 w-4" />;
      default:
        return null;
    }
  };

  const getOperationDescription = (op: string) => {
    switch (op) {
      case 'delete':
        return 'Permanently delete MOPs and MOs from M3';
      case 'close':
        return 'Close manufacturing orders (status > 22)';
      case 'reschedule':
        return 'Reschedule production orders to a new date';
      default:
        return '';
    }
  };

  return (
    <div className="bg-white rounded-lg shadow p-6">
      <h2 className="text-lg font-semibold mb-4 text-gray-900">Operation</h2>

      {/* Operation Selection */}
      <div className="mb-6">
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Select Action
        </label>
        <div className="space-y-2">
          {[
            { value: 'delete', label: 'Delete Production Orders', color: 'red' },
            { value: 'close', label: 'Close Manufacturing Orders', color: 'yellow' },
            { value: 'reschedule', label: 'Reschedule Orders', color: 'blue' },
          ].map((option) => (
            <button
              key={option.value}
              onClick={() => onOperationChange(option.value)}
              disabled={!previewData}
              className={`w-full text-left px-4 py-3 rounded-lg border-2 transition-all flex items-center gap-3 ${
                selectedOperation === option.value
                  ? option.color === 'red'
                    ? 'border-red-500 bg-red-50 text-red-900'
                    : option.color === 'yellow'
                    ? 'border-yellow-500 bg-yellow-50 text-yellow-900'
                    : 'border-blue-500 bg-blue-50 text-blue-900'
                  : 'border-gray-200 bg-white text-gray-700 hover:border-gray-300'
              } ${!previewData ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}`}
            >
              <div
                className={`flex-shrink-0 ${
                  selectedOperation === option.value
                    ? option.color === 'red'
                      ? 'text-red-600'
                      : option.color === 'yellow'
                      ? 'text-yellow-600'
                      : 'text-blue-600'
                    : 'text-gray-400'
                }`}
              >
                {getOperationIcon(option.value)}
              </div>
              <div className="flex-1">
                <div className="font-medium text-sm">{option.label}</div>
                <div className="text-xs text-gray-600 mt-0.5">
                  {getOperationDescription(option.value)}
                </div>
              </div>
              {selectedOperation === option.value && (
                <div
                  className={`flex-shrink-0 h-2 w-2 rounded-full ${
                    option.color === 'red'
                      ? 'bg-red-500'
                      : option.color === 'yellow'
                      ? 'bg-yellow-500'
                      : 'bg-blue-500'
                  }`}
                />
              )}
            </button>
          ))}
        </div>
      </div>

      {/* Operation-Specific Parameters */}
      {selectedOperation === 'reschedule' && (
        <div className="mb-6 p-4 bg-blue-50 rounded-lg border border-blue-200">
          <label className="block text-sm font-medium text-blue-900 mb-2">
            New Date (YYYYMMDD)
          </label>
          <input
            type="text"
            value={operationParams.new_date || ''}
            onChange={(e) =>
              onParamsChange({ ...operationParams, new_date: e.target.value })
            }
            placeholder="20260315"
            maxLength={8}
            className="w-full px-3 py-2 border border-blue-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
          />
          <p className="mt-2 text-xs text-blue-700">
            Enter the new start date in YYYYMMDD format
          </p>
        </div>
      )}

      {/* Warning Box */}
      {canExecute && previewData && (
        <div className="mb-6 p-4 bg-yellow-50 border-l-4 border-yellow-400 rounded">
          <div className="flex items-start">
            <svg
              className="h-5 w-5 text-yellow-400 mt-0.5 mr-3"
              fill="currentColor"
              viewBox="0 0 20 20"
            >
              <path
                fillRule="evenodd"
                d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z"
                clipRule="evenodd"
              />
            </svg>
            <div>
              <p className="text-sm font-medium text-yellow-800">
                Warning: Irreversible Operation
              </p>
              <p className="mt-1 text-sm text-yellow-700">
                This will affect{' '}
                <strong className="font-bold">
                  {previewData.total_count.toLocaleString()}
                </strong>{' '}
                production orders. This action cannot be undone.
              </p>
            </div>
          </div>
        </div>
      )}

      {/* Execute Button */}
      <button
        onClick={onExecute}
        disabled={!canExecute}
        className={`w-full px-4 py-3 rounded-md font-medium transition-all ${
          canExecute
            ? 'bg-red-600 text-white hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2 shadow-lg hover:shadow-xl'
            : 'bg-gray-200 text-gray-400 cursor-not-allowed'
        }`}
      >
        {canExecute ? (
          <span className="flex items-center justify-center">
            <svg
              className="w-5 h-5 mr-2"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M5 13l4 4L19 7"
              />
            </svg>
            Execute Bulk Operation
          </span>
        ) : (
          'Execute Bulk Operation'
        )}
      </button>

      {/* Help Text */}
      <div className="mt-6 pt-6 border-t border-gray-200">
        <h4 className="text-xs font-semibold text-gray-700 mb-2">What happens next:</h4>
        <ul className="text-xs text-gray-600 space-y-2">
          <li className="flex items-start">
            <span className="inline-block w-1.5 h-1.5 bg-gray-400 rounded-full mt-1.5 mr-2"></span>
            <span>Orders are processed in batches for optimal performance</span>
          </li>
          <li className="flex items-start">
            <span className="inline-block w-1.5 h-1.5 bg-gray-400 rounded-full mt-1.5 mr-2"></span>
            <span>Progress tracking modal shows real-time updates</span>
          </li>
          <li className="flex items-start">
            <span className="inline-block w-1.5 h-1.5 bg-gray-400 rounded-full mt-1.5 mr-2"></span>
            <span>Operation can be cancelled mid-execution if needed</span>
          </li>
          <li className="flex items-start">
            <span className="inline-block w-1.5 h-1.5 bg-gray-400 rounded-full mt-1.5 mr-2"></span>
            <span>All actions are logged in the audit trail</span>
          </li>
        </ul>
      </div>
    </div>
  );
};
