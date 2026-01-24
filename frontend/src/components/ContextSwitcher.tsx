import React, { useEffect, useState } from 'react';
import { useContextManagement } from '../contexts/ContextManagementContext';
import type { M3Warehouse } from '../types';

interface ContextSwitcherProps {
  isOpen: boolean;
  onClose: () => void;
}

function XMarkIcon({ className }: { className?: string }) {
  return (
    <svg className={className} fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
    </svg>
  );
}

function CheckIcon({ className }: { className?: string }) {
  return (
    <svg className={className} fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" d="M4.5 12.75l6 6 9-13.5" />
    </svg>
  );
}

function BuildingOfficeIcon({ className }: { className?: string }) {
  return (
    <svg className={className} fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" d="M3.75 21h16.5M4.5 3h15M5.25 3v18m13.5-18v18M9 6.75h1.5m-1.5 3h1.5m-1.5 3h1.5m3-6H15m-1.5 3H15m-1.5 3H15M9 21v-3.375c0-.621.504-1.125 1.125-1.125h3.75c.621 0 1.125.504 1.125 1.125V21" />
    </svg>
  );
}

export const ContextSwitcher: React.FC<ContextSwitcherProps> = ({ isOpen, onClose }) => {
  const {
    effectiveContext,
    warehouses,
    loading,
    error,
    loadWarehouses,
    setTemporaryOverride,
    clearTemporaryOverrides,
  } = useContextManagement();

  const [selectedWarehouse, setSelectedWarehouse] = useState<string>('');

  useEffect(() => {
    if (isOpen && effectiveContext) {
      loadWarehouses(
        effectiveContext.company,
        effectiveContext.division,
        effectiveContext.facility
      );
      setSelectedWarehouse(effectiveContext.warehouse);
    }
  }, [isOpen, effectiveContext, loadWarehouses]);

  const handleWarehouseChange = async (warehouse: string) => {
    try {
      if (warehouse === effectiveContext?.userDefaults.warehouse) {
        await clearTemporaryOverrides();
      } else {
        await setTemporaryOverride({ warehouse });
      }
      onClose();
    } catch (err) {
      console.error('Failed to change warehouse:', err);
    }
  };

  const handleReset = async () => {
    try {
      await clearTemporaryOverrides();
      onClose();
    } catch (err) {
      console.error('Failed to reset context:', err);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-[100] overflow-y-auto">
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-slate-900/60 backdrop-blur-sm transition-opacity"
        onClick={onClose}
      />

      {/* Modal */}
      <div className="flex min-h-full items-center justify-center p-4">
        <div
          className="relative w-full max-w-lg transform overflow-hidden rounded-xl bg-white shadow-2xl transition-all"
          onClick={(e) => e.stopPropagation()}
        >
          {/* Header */}
          <div className="flex items-center justify-between border-b border-slate-200 px-6 py-4">
            <div className="flex items-center gap-3">
              <div className="rounded-lg bg-primary-100 p-2">
                <BuildingOfficeIcon className="h-5 w-5 text-primary-600" />
              </div>
              <div>
                <h2 className="text-lg font-semibold text-slate-900">Switch Context</h2>
                <p className="text-sm text-slate-500">Change your active warehouse</p>
              </div>
            </div>
            <button
              onClick={onClose}
              className="rounded-lg p-2 text-slate-400 transition-colors hover:bg-slate-100 hover:text-slate-600"
            >
              <XMarkIcon className="h-5 w-5" />
            </button>
          </div>

          {/* Current Context */}
          <div className="border-b border-slate-200 bg-slate-50 px-6 py-4">
            <label className="text-xs font-medium uppercase tracking-wide text-slate-500">
              Current Context
            </label>
            <div className="mt-2 flex items-center gap-2">
              <code className="flex-1 rounded-lg bg-white px-3 py-2 font-mono text-sm text-slate-700 ring-1 ring-slate-200">
                {effectiveContext
                  ? `${effectiveContext.company} / ${effectiveContext.division} / ${effectiveContext.facility} / ${effectiveContext.warehouse}`
                  : 'Loading...'}
              </code>
              {effectiveContext?.hasTemporaryOverrides && (
                <span className="rounded-full bg-warning-100 px-2.5 py-1 text-xs font-medium text-warning-700">
                  Override
                </span>
              )}
            </div>
          </div>

          {/* Warehouse List */}
          <div className="px-6 py-4">
            <h3 className="mb-3 text-sm font-medium text-slate-700">Select Warehouse</h3>

            {loading && (
              <div className="flex items-center justify-center py-8">
                <div className="h-6 w-6 animate-spin rounded-full border-2 border-primary-200 border-t-primary-600" />
              </div>
            )}

            {error && (
              <div className="rounded-lg bg-error-50 px-4 py-3 text-sm text-error-600">
                {error}
              </div>
            )}

            {!loading && !error && (
              <div className="max-h-64 space-y-2 overflow-y-auto">
                {/* Default option */}
                <button
                  onClick={() => handleWarehouseChange(effectiveContext?.userDefaults.warehouse || '')}
                  className={`flex w-full items-center justify-between rounded-lg border-2 px-4 py-3 text-left transition-all ${
                    !effectiveContext?.hasTemporaryOverrides
                      ? 'border-primary-500 bg-primary-50'
                      : 'border-slate-200 hover:border-slate-300 hover:bg-slate-50'
                  }`}
                >
                  <div>
                    <div className="font-medium text-slate-900">
                      Use Default ({effectiveContext?.userDefaults.warehouse})
                    </div>
                    <div className="text-sm text-slate-500">Your configured default warehouse</div>
                  </div>
                  {!effectiveContext?.hasTemporaryOverrides && (
                    <div className="flex h-6 w-6 items-center justify-center rounded-full bg-primary-500">
                      <CheckIcon className="h-4 w-4 text-white" />
                    </div>
                  )}
                </button>

                {/* Available warehouses */}
                {warehouses.map((wh: M3Warehouse) => {
                  const isSelected =
                    effectiveContext?.hasTemporaryOverrides &&
                    effectiveContext.warehouse === wh.warehouse;

                  return (
                    <button
                      key={wh.warehouse}
                      onClick={() => handleWarehouseChange(wh.warehouse)}
                      className={`flex w-full items-center justify-between rounded-lg border-2 px-4 py-3 text-left transition-all ${
                        isSelected
                          ? 'border-primary-500 bg-primary-50'
                          : 'border-slate-200 hover:border-slate-300 hover:bg-slate-50'
                      }`}
                    >
                      <div>
                        <div className="font-medium text-slate-900">
                          {wh.warehouse} - {wh.warehouseName}
                        </div>
                        {wh.division && wh.facility && (
                          <div className="text-sm text-slate-500">
                            {wh.division} / {wh.facility}
                          </div>
                        )}
                      </div>
                      {isSelected && (
                        <div className="flex h-6 w-6 items-center justify-center rounded-full bg-primary-500">
                          <CheckIcon className="h-4 w-4 text-white" />
                        </div>
                      )}
                    </button>
                  );
                })}
              </div>
            )}
          </div>

          {/* Footer */}
          <div className="flex items-center justify-end gap-3 border-t border-slate-200 bg-slate-50 px-6 py-4">
            {effectiveContext?.hasTemporaryOverrides && (
              <button
                onClick={handleReset}
                className="rounded-lg px-4 py-2 text-sm font-medium text-slate-700 transition-colors hover:bg-slate-200"
              >
                Reset to Default
              </button>
            )}
            <button
              onClick={onClose}
              className="rounded-lg bg-primary-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-500"
            >
              Done
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};
