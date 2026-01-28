import React from 'react';
import { X, Trash2, XCircle, Calendar } from 'lucide-react';

export interface BulkAction {
  id: string;
  label: string;
  icon: React.ReactNode;
  variant: 'danger' | 'warning' | 'primary';
  enabled: boolean;
  disabledReason?: string;
}

interface BulkActionToolbarProps {
  selectedCount: number;
  availableActions: BulkAction[];
  onExecute: (actionId: string) => void;
  onClear: () => void;
}

export function BulkActionToolbar({
  selectedCount,
  availableActions,
  onExecute,
  onClear,
}: BulkActionToolbarProps) {
  if (selectedCount === 0) {
    return null;
  }

  return (
    <div className="fixed bottom-6 left-1/2 transform -translate-x-1/2 z-50">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 px-6 py-4">
        <div className="flex items-center gap-6">
          {/* Selection count */}
          <div className="flex items-center gap-3">
            <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
              {selectedCount} {selectedCount === 1 ? 'item' : 'items'} selected
            </span>
            <button
              onClick={onClear}
              className="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
              title="Clear selection"
            >
              <X className="h-4 w-4" />
            </button>
          </div>

          {/* Divider */}
          <div className="h-6 w-px bg-gray-300 dark:bg-gray-600" />

          {/* Action buttons */}
          <div className="flex items-center gap-2">
            {availableActions.map((action) => (
              <button
                key={action.id}
                onClick={() => action.enabled && onExecute(action.id)}
                disabled={!action.enabled}
                title={action.enabled ? action.label : action.disabledReason}
                className={`
                  flex items-center gap-2 px-4 py-2 rounded-md text-sm font-medium
                  transition-colors duration-200
                  ${action.enabled
                    ? action.variant === 'danger'
                      ? 'bg-red-600 hover:bg-red-700 text-white'
                      : action.variant === 'warning'
                      ? 'bg-yellow-600 hover:bg-yellow-700 text-white'
                      : 'bg-blue-600 hover:bg-blue-700 text-white'
                    : 'bg-gray-300 dark:bg-gray-700 text-gray-500 dark:text-gray-500 cursor-not-allowed'
                  }
                `}
              >
                {action.icon}
                <span>{action.label}</span>
              </button>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

export default BulkActionToolbar;
