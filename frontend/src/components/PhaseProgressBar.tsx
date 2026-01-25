import React from 'react';
import { PhaseProgress } from '../types';

interface PhaseProgressBarProps {
  phase: PhaseProgress;
  label: string;
}

const PhaseProgressBar: React.FC<PhaseProgressBarProps> = ({ phase, label }) => {
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

  const getProgressWidth = () => {
    if (phase.status === 'completed') return '100%';
    if (phase.status === 'running') return '50%';
    return '0%';
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
