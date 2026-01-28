import React from 'react';
import { PhaseProgress } from '../types';

interface PhaseProgressBarProps {
  phase: PhaseProgress;
  label: string;
}

const PhaseProgressBar: React.FC<PhaseProgressBarProps> = ({ phase, label }) => {
  const getOperationProgress = (operation: string | undefined, recordCount?: number): number => {
    if (!operation) return 5; // Just started, no operation yet

    // Querying phase: 0-30%
    if (operation.toLowerCase().includes('querying') || operation.toLowerCase().includes('loading page')) {
      // Check for pagination progress "Loading page X/Y"
      const pageMatch = operation.match(/page (\d+)\/(\d+)/i);
      if (pageMatch) {
        const current = parseInt(pageMatch[1]);
        const total = parseInt(pageMatch[2]);
        if (total > 0) {
          return Math.floor((current / total) * 30);
        }
      }
      return 15; // Mid-point of query phase
    }

    // Processing phase: 30-50%
    if (operation.toLowerCase().includes('processing')) {
      return 40;
    }

    // Inserting phase: 50-95%
    if (operation.toLowerCase().includes('inserting')) {
      // Try to extract total count from operation string
      // Example: "Inserting 15234 customer order lines into database..."
      const countMatch = operation.match(/inserting (\d+)/i);
      if (countMatch && recordCount) {
        const totalRecords = parseInt(countMatch[1]);
        if (totalRecords > 0) {
          // Calculate progress within 50-95% range (45% span)
          const insertProgress = Math.min((recordCount / totalRecords) * 45, 45);
          return 50 + Math.floor(insertProgress);
        }
      }
      return 70; // Mid-point of insert phase if no count available
    }

    // Unknown operation - conservative estimate
    return 10;
  };

  const getStatusIcon = () => {
    switch (phase.status) {
      case 'completed':
        return <span className="text-green-600 font-bold">✓</span>;
      case 'running':
        return (
          <div className="animate-spin h-4 w-4 border-2 border-primary-600 border-t-transparent rounded-full" />
        );
      case 'failed':
        return <span className="text-red-600 font-bold">✗</span>;
      default:
        return <span className="text-slate-400">○</span>;
    }
  };

  const getStatusColor = () => {
    switch (phase.status) {
      case 'completed': return 'bg-green-600';
      case 'running': return 'bg-primary-600';
      case 'failed': return 'bg-red-600';
      default: return 'bg-slate-300';
    }
  };

  const getProgressWidth = (): string => {
    if (phase.status === 'completed') return '100%';
    if (phase.status === 'failed') return '0%';
    if (phase.status === 'pending') return '0%';

    // Running phase - calculate based on operation
    const progress = getOperationProgress(phase.currentOperation, phase.recordCount);

    // Cap at 95% until actually completed (prevents confusing "100% but still running")
    return `${Math.min(progress, 95)}%`;
  };

  return (
    <div className="flex items-center gap-3 py-2">
      <div className="flex-shrink-0 w-5 h-5 flex items-center justify-center">
        {getStatusIcon()}
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex justify-between items-center mb-1">
          <span className={`text-sm font-medium ${
            phase.status === 'running' ? 'text-slate-900' : 'text-slate-600'
          }`}>
            {label}
          </span>
          {phase.recordCount !== undefined && phase.recordCount > 0 && (
            <span className="text-xs text-slate-500">
              {phase.recordCount.toLocaleString()} records
            </span>
          )}
        </div>

        {/* Visual operation stage indicator */}
        {phase.status === 'running' && (
          <div className="flex items-center gap-2 mb-1 text-xs">
            <div className={`flex items-center gap-1 ${
              phase.currentOperation?.toLowerCase().includes('querying') || phase.currentOperation?.toLowerCase().includes('loading page')
                ? 'text-primary-600 font-semibold'
                : 'text-slate-400'
            }`}>
              <div className="h-1.5 w-1.5 rounded-full bg-current" />
              Query
            </div>
            <div className="h-px w-3 bg-slate-300" />
            <div className={`flex items-center gap-1 ${
              phase.currentOperation?.toLowerCase().includes('processing')
                ? 'text-primary-600 font-semibold'
                : 'text-slate-400'
            }`}>
              <div className="h-1.5 w-1.5 rounded-full bg-current" />
              Process
            </div>
            <div className="h-px w-3 bg-slate-300" />
            <div className={`flex items-center gap-1 ${
              phase.currentOperation?.toLowerCase().includes('inserting')
                ? 'text-primary-600 font-semibold'
                : 'text-slate-400'
            }`}>
              <div className="h-1.5 w-1.5 rounded-full bg-current" />
              Insert
            </div>
          </div>
        )}

        {/* Show current operation when running */}
        {phase.status === 'running' && phase.currentOperation && (
          <p className="text-xs text-slate-500 pl-6 italic mb-1">
            {phase.currentOperation}
          </p>
        )}

        <div className="h-2 w-full rounded-full bg-slate-100">
          <div
            className={`h-2 rounded-full transition-all duration-500 ${getStatusColor()}`}
            style={{ width: getProgressWidth() }}
          />
        </div>
      </div>
    </div>
  );
};

export default PhaseProgressBar;
