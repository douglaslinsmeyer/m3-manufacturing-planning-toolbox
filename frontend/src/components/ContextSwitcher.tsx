import React, { useEffect, useState } from 'react';
import { useContextManagement } from '../hooks/useContextManagement';
import type { M3Warehouse } from '../types';
import './ContextSwitcher.css';

interface ContextSwitcherProps {
  isOpen: boolean;
  onClose: () => void;
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
      // Load warehouses filtered by current company/division/facility
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
        // Switching back to default, clear override
        await clearTemporaryOverrides();
      } else {
        // Set temporary override
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
    <div className="context-switcher-overlay" onClick={onClose}>
      <div className="context-switcher-modal" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h2>Context Switch</h2>
          <button className="close-button" onClick={onClose}>×</button>
        </div>

        <div className="modal-content">
          {/* Current Context Display */}
          <div className="current-context-card">
            <label>Current Context</label>
            <div className="context-display">
              {effectiveContext && (
                <>
                  <span>{effectiveContext.company} • {effectiveContext.division} • {effectiveContext.facility} • {effectiveContext.warehouse}</span>
                  {effectiveContext.hasTemporaryOverrides && (
                    <span className="override-indicator">Temporary override active</span>
                  )}
                </>
              )}
            </div>
          </div>

          {/* Warehouse Switcher */}
          <div className="warehouse-section">
            <h3>Quick Warehouse Switch</h3>
            {loading && <div className="loading-spinner">Loading warehouses...</div>}
            {error && <div className="error-message">{error}</div>}
            {!loading && !error && (
              <div className="warehouse-list">
                {/* Default option */}
                <div
                  className={`warehouse-option ${
                    !effectiveContext?.hasTemporaryOverrides ? 'selected' : ''
                  }`}
                  onClick={() => handleWarehouseChange(effectiveContext?.userDefaults.warehouse || '')}
                >
                  <span>Use default ({effectiveContext?.userDefaults.warehouse})</span>
                  {!effectiveContext?.hasTemporaryOverrides && <span className="checkmark">✓</span>}
                </div>

                {/* Available warehouses */}
                {warehouses.map((wh: M3Warehouse) => (
                  <div
                    key={wh.warehouse}
                    className={`warehouse-option ${
                      effectiveContext?.hasTemporaryOverrides &&
                      effectiveContext.warehouse === wh.warehouse
                        ? 'selected'
                        : ''
                    }`}
                    onClick={() => handleWarehouseChange(wh.warehouse)}
                  >
                    <div>
                      <div className="warehouse-name">{wh.warehouse} - {wh.warehouseName}</div>
                      {wh.division && wh.facility && (
                        <div className="warehouse-detail">{wh.division} • {wh.facility}</div>
                      )}
                    </div>
                    {effectiveContext?.hasTemporaryOverrides &&
                      effectiveContext.warehouse === wh.warehouse && (
                        <span className="checkmark">✓</span>
                      )}
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Action Buttons */}
          <div className="modal-actions">
            {effectiveContext?.hasTemporaryOverrides && (
              <button className="reset-button" onClick={handleReset}>
                Reset to Default
              </button>
            )}
            <button className="done-button" onClick={onClose}>
              Done
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};
