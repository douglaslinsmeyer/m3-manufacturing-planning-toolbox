import React from 'react';
import { PreviewData } from '../../pages/BulkOperations';

interface PreviewPanelProps {
  previewData: PreviewData | null;
  isLoading: boolean;
}

export const PreviewPanel: React.FC<PreviewPanelProps> = ({
  previewData,
  isLoading,
}) => {
  if (isLoading) {
    return (
      <div className="bg-white rounded-lg shadow p-6">
        <div className="flex items-center justify-center h-96">
          <div className="text-center">
            <svg
              className="animate-spin h-10 w-10 text-blue-600 mx-auto mb-4"
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
            <p className="text-gray-600">Loading preview...</p>
          </div>
        </div>
      </div>
    );
  }

  if (!previewData) {
    return (
      <div className="bg-white rounded-lg shadow p-6">
        <div className="flex items-center justify-center h-96 text-gray-400">
          <div className="text-center">
            <svg
              className="mx-auto h-12 w-12 text-gray-400 mb-4"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              strokeWidth={1}
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01"
              />
            </svg>
            <p className="text-lg font-medium mb-1">No preview available</p>
            <p className="text-sm">Set filters and click Preview</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-lg shadow p-6">
      <h2 className="text-lg font-semibold mb-4 text-gray-900">Preview</h2>

      {/* Total Count */}
      <div className="mb-6 p-6 bg-gradient-to-br from-blue-50 to-blue-100 rounded-lg border border-blue-200">
        <p className="text-4xl font-bold text-blue-900 mb-1">
          {previewData.total_count.toLocaleString()}
        </p>
        <p className="text-sm text-blue-700 font-medium">Matching Issues</p>
      </div>

      {/* Summary Statistics */}
      <div className="space-y-6">
        {/* By Order Type */}
        {Object.keys(previewData.summary.by_order_type).length > 0 && (
          <div>
            <h3 className="text-sm font-medium text-gray-700 mb-2">By Order Type</h3>
            <div className="space-y-2">
              {Object.entries(previewData.summary.by_order_type).map(([type, count]) => (
                <div
                  key={type}
                  className="flex justify-between items-center py-2 px-3 bg-gray-50 rounded-md"
                >
                  <span className="text-sm text-gray-700 font-medium">{type}</span>
                  <span className="text-sm font-semibold text-gray-900">{count.toLocaleString()}</span>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* By Facility */}
        {Object.keys(previewData.summary.by_facility).length > 0 && (
          <div>
            <h3 className="text-sm font-medium text-gray-700 mb-2">By Facility</h3>
            <div className="space-y-2">
              {Object.entries(previewData.summary.by_facility).map(([fac, count]) => (
                <div
                  key={fac}
                  className="flex justify-between items-center py-2 px-3 bg-gray-50 rounded-md"
                >
                  <span className="text-sm text-gray-700 font-medium">{fac}</span>
                  <span className="text-sm font-semibold text-gray-900">{count.toLocaleString()}</span>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* By Detector Type */}
        {Object.keys(previewData.summary.by_detector).length > 0 && (
          <div>
            <h3 className="text-sm font-medium text-gray-700 mb-2">By Issue Type</h3>
            <div className="space-y-2">
              {Object.entries(previewData.summary.by_detector).map(([det, count]) => (
                <div
                  key={det}
                  className="flex justify-between items-center py-2 px-3 bg-gray-50 rounded-md"
                >
                  <span className="text-sm text-gray-700 font-medium truncate">{det}</span>
                  <span className="text-sm font-semibold text-gray-900">{count.toLocaleString()}</span>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>

      {/* Sample Issues */}
      {previewData.sample_issues && previewData.sample_issues.length > 0 && (
        <div className="mt-6 pt-6 border-t border-gray-200">
          <h3 className="text-sm font-medium text-gray-700 mb-3">Sample Issues</h3>
          <div className="space-y-2 max-h-64 overflow-y-auto">
            {previewData.sample_issues.slice(0, 10).map((issue) => (
              <div
                key={issue.id}
                className="p-3 bg-gray-50 rounded-md text-xs border border-gray-200 hover:border-gray-300 transition-colors"
              >
                <div className="flex items-center justify-between mb-1">
                  <span className="font-semibold text-gray-900">
                    {issue.production_order_number}
                  </span>
                  <span className="text-xs px-2 py-0.5 bg-blue-100 text-blue-800 rounded-full font-medium">
                    {issue.production_order_type}
                  </span>
                </div>
                <div className="text-gray-600">
                  {issue.detector_type.replace(/_/g, ' ')}
                </div>
                <div className="text-gray-500 mt-1">
                  {issue.facility} {issue.warehouse && `• ${issue.warehouse}`}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};
