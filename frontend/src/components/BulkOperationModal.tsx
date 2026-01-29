import React from 'react';
import { CheckCircle, XCircle, Loader2, X } from 'lucide-react';

export interface BulkOperationResult {
  issue_id: number;
  production_order: string;
  status: 'success' | 'error' | 'pending';
  message?: string;
  error?: string;
  is_duplicate?: boolean;         // NEW
  primary_issue_id?: number;      // NEW
}

interface BulkOperationModalProps {
  isOpen: boolean;
  title: string;
  total: number;
  completed: number;
  results: BulkOperationResult[];
  onClose: () => void;
  isProcessing: boolean;
}

export function BulkOperationModal({
  isOpen,
  title,
  total,
  completed,
  results,
  onClose,
  isProcessing,
}: BulkOperationModalProps) {
  if (!isOpen) {
    return null;
  }

  const successCount = results.filter((r) => r.status === 'success').length;
  const failureCount = results.filter((r) => r.status === 'error').length;
  const progressPercent = total > 0 ? (completed / total) * 100 : 0;

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      <div className="flex min-h-screen items-center justify-center p-4">
        {/* Backdrop */}
        <div
          className="fixed inset-0 bg-black bg-opacity-50 transition-opacity"
          onClick={!isProcessing ? onClose : undefined}
        />

        {/* Modal */}
        <div className="relative bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-2xl w-full max-h-[80vh] flex flex-col">
          {/* Header */}
          <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-700">
            <h2 className="text-lg font-semibold text-gray-900 dark:text-white">{title}</h2>
            {!isProcessing && (
              <button
                onClick={onClose}
                className="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
              >
                <X className="h-5 w-5" />
              </button>
            )}
          </div>

          {/* Progress bar */}
          <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
            <div className="mb-2 flex justify-between text-sm text-gray-600 dark:text-gray-400">
              <span>
                {completed} of {total} {completed === 1 ? 'item' : 'items'} processed
              </span>
              <span>{Math.round(progressPercent)}%</span>
            </div>
            <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2.5">
              <div
                className="bg-blue-600 h-2.5 rounded-full transition-all duration-300"
                style={{ width: `${progressPercent}%` }}
              />
            </div>
          </div>

          {/* Results list */}
          <div className="flex-1 overflow-y-auto px-6 py-4">
            <div className="space-y-2">
              {results.map((result) => (
                <div
                  key={result.issue_id}
                  className={`
                    flex items-start gap-3 p-3 rounded-md
                    ${result.status === 'success'
                      ? 'bg-green-50 dark:bg-green-900/20'
                      : result.status === 'error'
                      ? 'bg-red-50 dark:bg-red-900/20'
                      : 'bg-gray-50 dark:bg-gray-700/50'
                    }
                  `}
                >
                  {/* Status icon */}
                  <div className="flex-shrink-0 mt-0.5">
                    {result.status === 'success' && (
                      <CheckCircle className="h-5 w-5 text-green-600 dark:text-green-400" />
                    )}
                    {result.status === 'error' && (
                      <XCircle className="h-5 w-5 text-red-600 dark:text-red-400" />
                    )}
                    {result.status === 'pending' && (
                      <Loader2 className="h-5 w-5 text-gray-400 animate-spin" />
                    )}
                  </div>

                  {/* Content */}
                  <div className="flex-1 min-w-0">
                    <div className="flex items-baseline gap-2 flex-wrap">
                      <span className="font-medium text-gray-900 dark:text-white">
                        {result.production_order}
                      </span>
                      <span className="text-xs text-gray-500 dark:text-gray-400">
                        (Issue #{result.issue_id})
                      </span>

                      {/* Duplicate indicator badge */}
                      {result.is_duplicate && result.primary_issue_id && (
                        <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-100 dark:bg-blue-900/30 text-blue-800 dark:text-blue-300">
                          Same order as #{result.primary_issue_id}
                        </span>
                      )}
                    </div>
                    {result.message && (
                      <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                        {result.message}
                      </p>
                    )}
                    {result.error && (
                      <p className="text-sm text-red-600 dark:text-red-400 mt-1">
                        {result.error}
                      </p>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Footer */}
          <div className="px-6 py-4 border-t border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-900/50">
            <div className="flex items-center justify-between">
              <div className="text-sm text-gray-600 dark:text-gray-400">
                {isProcessing ? (
                  <div className="flex items-center gap-2">
                    <Loader2 className="h-4 w-4 animate-spin" />
                    <span>Processing...</span>
                  </div>
                ) : (
                  <span>
                    <span className="text-green-600 dark:text-green-400 font-medium">
                      {successCount} successful
                    </span>
                    {failureCount > 0 && (
                      <>
                        ,{' '}
                        <span className="text-red-600 dark:text-red-400 font-medium">
                          {failureCount} failed
                        </span>
                      </>
                    )}
                    {' '}out of {total} total
                  </span>
                )}
              </div>
              {!isProcessing && (
                <button
                  onClick={onClose}
                  className="px-4 py-2 bg-gray-600 hover:bg-gray-700 text-white rounded-md text-sm font-medium transition-colors"
                >
                  Close
                </button>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export default BulkOperationModal;
