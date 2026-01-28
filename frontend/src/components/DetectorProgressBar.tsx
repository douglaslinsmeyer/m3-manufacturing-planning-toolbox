import React from 'react';
import { DetectorProgress } from '../types';

interface DetectorProgressBarProps {
  detector: DetectorProgress;
}

const DetectorProgressBar: React.FC<DetectorProgressBarProps> = ({ detector }) => {
  const getStatusIcon = () => {
    switch (detector.status) {
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
    switch (detector.status) {
      case 'completed': return 'bg-green-600';
      case 'running': return 'bg-primary-600';
      case 'failed': return 'bg-red-600';
      default: return 'bg-slate-300';
    }
  };

  const getProgressWidth = () => {
    if (detector.status === 'completed') return '100%';
    if (detector.status === 'running') return '50%';
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
            detector.status === 'running' ? 'text-slate-900' : 'text-slate-600'
          }`}>
            {detector.displayLabel}
          </span>
          <div className="flex items-center gap-2 text-xs text-slate-500">
            {detector.issuesFound !== undefined && detector.issuesFound > 0 && (
              <span className="font-medium text-warning-600">
                {detector.issuesFound} {detector.issuesFound === 1 ? 'issue' : 'issues'}
              </span>
            )}
            {detector.durationMs !== undefined && detector.durationMs > 0 && (
              <span>
                {detector.durationMs < 1000
                  ? `${detector.durationMs}ms`
                  : `${(detector.durationMs / 1000).toFixed(1)}s`}
              </span>
            )}
          </div>
        </div>
        <div className="h-2 w-full rounded-full bg-slate-100">
          <div
            className={`h-2 rounded-full transition-all duration-500 ${getStatusColor()}`}
            style={{ width: getProgressWidth() }}
          />
        </div>
        {detector.error && (
          <p className="mt-1 text-xs text-red-600">{detector.error}</p>
        )}
      </div>
    </div>
  );
};

export default DetectorProgressBar;
