import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { api, IssueSummary } from '../services/api';
import { AppLayout } from '../components/AppLayout';
import { useSnapshotProgress } from '../hooks/useSnapshotProgress';
import { useContextManagement } from '../hooks/useContextManagement';
import { IssueBreakdownHierarchy } from '../components/IssueBreakdownHierarchy';
import PhaseProgressBar from '../components/PhaseProgressBar';
import DetectorProgressBar from '../components/DetectorProgressBar';
import { DetectorTrigger } from '../components/DetectorTrigger';
import { ToastContainer } from '../components/Toast';
import { useToast } from '../hooks/useToast';
import type { SnapshotSummary, SnapshotStatus } from '../types';

function ArrowPathIcon({ className }: { className?: string }) {
  return (
    <svg className={className} fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0l3.181 3.183a8.25 8.25 0 0013.803-3.7M4.031 9.865a8.25 8.25 0 0113.803-3.7l3.181 3.182m0-4.991v4.99" />
    </svg>
  );
}

function ExclamationTriangleIcon({ className }: { className?: string }) {
  return (
    <svg className={className} fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z" />
    </svg>
  );
}

function CheckCircleIcon({ className }: { className?: string }) {
  return (
    <svg className={className} fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
    </svg>
  );
}

function ArrowTrendingUpIcon({ className }: { className?: string }) {
  return (
    <svg className={className} fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" d="M2.25 18L9 11.25l4.306 4.307a11.95 11.95 0 015.814-5.519l2.74-1.22m0 0l-5.94-2.28m5.94 2.28l-2.28 5.941" />
    </svg>
  );
}

function BuildingIcon({ className }: { className?: string }) {
  return (
    <svg className={className} fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" d="M3.75 21h16.5M4.5 3h15M5.25 3v18m13.5-18v18M9 6.75h1.5m-1.5 3h1.5m-1.5 3h1.5m3-6H15m-1.5 3H15m-1.5 3H15M9 21v-3.375c0-.621.504-1.125 1.125-1.125h3.75c.621 0 1.125.504 1.125 1.125V21" />
    </svg>
  );
}

function CubeIcon({ className }: { className?: string }) {
  return (
    <svg className={className} fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" d="M21 7.5l-9-5.25L3 7.5m18 0l-9 5.25m9-5.25v9l-9 5.25M3 7.5l9 5.25M3 7.5v9l9 5.25m0-9v9" />
    </svg>
  );
}

function ShoppingCartIcon({ className }: { className?: string }) {
  return (
    <svg className={className} fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" d="M2.25 3h1.386c.51 0 .955.343 1.087.835l.383 1.437M7.5 14.25a3 3 0 00-3 3h15.75m-12.75-3h11.218c1.121-2.3 2.1-4.684 2.924-7.138a60.114 60.114 0 00-16.536-1.84M7.5 14.25L5.106 5.272M6 20.25a.75.75 0 11-1.5 0 .75.75 0 011.5 0zm12.75 0a.75.75 0 11-1.5 0 .75.75 0 011.5 0z" />
    </svg>
  );
}

const stats = [
  { name: 'Current Facility', key: 'currentFacility', href: '#', icon: BuildingIcon, color: 'info' },
  { name: 'Production Orders', key: 'totalProductionOrders', href: '#', icon: CubeIcon, color: 'primary' },
  { name: 'CO Lines', key: 'totalCustomerOrderLines', href: '#', icon: ShoppingCartIcon, color: 'success' },
  { name: 'Inconsistencies', key: 'inconsistenciesCount', href: '/inconsistencies', icon: ExclamationTriangleIcon, color: 'warning' },
];

const Dashboard: React.FC = () => {
  const { isAuthenticated, loading: authLoading } = useAuth();
  const { effectiveContext, loadEffectiveContext } = useContextManagement();
  const toast = useToast();
  const [summary, setSummary] = useState<SnapshotSummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [currentJobId, setCurrentJobId] = useState<string | null>(null);
  const [recovering, setRecovering] = useState(false);
  const [issueSummary, setIssueSummary] = useState<IssueSummary | null>(null);
  const [scopeInfoExpanded, setScopeInfoExpanded] = useState(false);

  // Use SSE hook for real-time progress updates
  const { status: sseStatus, isConnected, error: sseError } = useSnapshotProgress(currentJobId);

  // Combine SSE status with fallback to API polling for initial state
  const [fallbackStatus, setFallbackStatus] = useState<SnapshotStatus | null>(null);
  const snapshotStatus = sseStatus || fallbackStatus;

  // Check for in-progress refresh on page load
  useEffect(() => {
    if (authLoading || !isAuthenticated) {
      return;
    }

    const checkForActiveJob = async () => {
      try {
        setRecovering(true);
        const { jobId } = await api.getActiveJob();

        if (jobId) {
          console.log('Reconnecting to in-progress refresh:', jobId);
          setCurrentJobId(jobId);
        }
      } catch (error) {
        console.error('Failed to check for active job:', error);
        // Non-fatal error - continue normal page load
      } finally {
        setRecovering(false);
      }
    };

    checkForActiveJob();
  }, [isAuthenticated, authLoading]);

  // Load effective context on mount
  useEffect(() => {
    if (authLoading || !isAuthenticated) {
      return;
    }

    loadEffectiveContext();
  }, [isAuthenticated, authLoading, loadEffectiveContext]);

  useEffect(() => {
    if (authLoading || !isAuthenticated) {
      return;
    }

    loadDashboardData();
  }, [isAuthenticated, authLoading]);

  // Check if refresh is complete and reload data
  useEffect(() => {
    if (snapshotStatus?.status === 'completed') {
      loadDashboardData();
      // Clear job ID after completion
      setTimeout(() => setCurrentJobId(null), 1000);
    }
  }, [snapshotStatus?.status]);

  const loadDashboardData = async () => {
    try {
      const [summaryData, statusData, issueData] = await Promise.all([
        api.getSnapshotSummary(),
        api.getSnapshotStatus(),
        api.getIssueSummary(),
      ]);
      setSummary(summaryData);
      setFallbackStatus(statusData);
      setIssueSummary(issueData);
    } catch (error) {
      console.error('Failed to load dashboard data:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleRefresh = async () => {
    setRefreshing(true);
    try {
      const response = await api.refreshSnapshot();
      console.log('Refresh response:', response);
      // Extract job ID from response to start SSE connection
      if (response.jobId) {
        console.log('Setting currentJobId:', response.jobId);
        setCurrentJobId(response.jobId);
      } else {
        console.error('No jobId in response:', response);
      }
    } catch (error) {
      console.error('Failed to refresh snapshot:', error);
    } finally {
      setRefreshing(false);
    }
  };

  const handleCancelRefresh = async () => {
    if (!currentJobId) return;

    try {
      await api.cancelRefresh(currentJobId);

      console.log('Refresh cancelled successfully');
      toast.success('Data refresh cancelled successfully');
      setCurrentJobId(null);

      // Reload dashboard data to reflect cancellation
      await loadDashboardData();
    } catch (error) {
      console.error('Failed to cancel refresh:', error);
      toast.error('Failed to cancel data refresh');
    }
  };

  const getStatValue = (key: string): string | number => {
    if (!summary) return 0;

    // Special handling for facility context tile
    if (key === 'currentFacility') {
      return effectiveContext?.facility || '---';
    }

    // Use real count from issue summary for inconsistencies
    if (key === 'inconsistenciesCount' && issueSummary) {
      return issueSummary.total;
    }

    return (summary as any)[key] || 0;
  };

  const getColorClasses = (color: string) => {
    const colors: Record<string, { bg: string; icon: string; text: string }> = {
      primary: { bg: 'bg-primary-50', icon: 'text-primary-600', text: 'text-primary-600' },
      info: { bg: 'bg-info-50', icon: 'text-info-600', text: 'text-info-600' },
      success: { bg: 'bg-success-50', icon: 'text-success-600', text: 'text-success-600' },
      warning: { bg: 'bg-warning-50', icon: 'text-warning-600', text: 'text-warning-600' },
    };
    return colors[color] || colors.primary;
  };

  if (loading) {
    return (
      <AppLayout>
        <div className="flex items-center justify-center h-96">
          <div className="flex flex-col items-center gap-4">
            <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary-200 border-t-primary-600" />
            <p className="text-sm text-slate-500">Loading dashboard...</p>
          </div>
        </div>
      </AppLayout>
    );
  }

  return (
    <AppLayout>
      <div className="px-4 py-6 sm:px-6 lg:px-12 lg:py-10">
        {/* Page Header */}
        <div className="mb-6 lg:mb-10">
          <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <h1 className="text-xl font-semibold text-slate-900 sm:text-2xl">Dashboard</h1>
              <p className="mt-1 text-sm text-slate-500">
                Overview of your manufacturing planning data
              </p>
            </div>
            <div className="flex items-center gap-2">
              <DetectorTrigger
                environment={effectiveContext?.environment || 'PRD'}
                disabled={refreshing || snapshotStatus?.status === 'running' || !summary}
                onTrigger={(jobId) => {
                  console.log('Detection triggered:', jobId);
                  // TODO: Could add toast notification or tracking here
                }}
              />
              <button
                onClick={handleRefresh}
                disabled={refreshing || snapshotStatus?.status === 'running'}
                className="inline-flex items-center gap-2 rounded-md bg-primary-600 px-3 py-1.5 text-sm font-medium text-white shadow-sm transition-all hover:bg-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <ArrowPathIcon className={`h-4 w-4 ${refreshing || snapshotStatus?.status === 'running' ? 'animate-spin' : ''}`} />
                {refreshing || snapshotStatus?.status === 'running' ? 'Refreshing...' : 'Refresh Data'}
              </button>
            </div>
          </div>
        </div>

        {/* Reconnection indicator */}
        {recovering && (
          <div className="mb-6 lg:mb-10 rounded-lg bg-blue-50 p-4 shadow-sm ring-1 ring-blue-200">
            <div className="flex items-center gap-3">
              <div className="h-4 w-4 animate-spin rounded-full border-2 border-blue-200 border-t-blue-600" />
              <span className="text-sm font-medium text-blue-900">
                Checking for in-progress refresh...
              </span>
            </div>
          </div>
        )}

        {/* Refresh Progress */}
        {snapshotStatus?.status === 'running' && (
          <div className="mb-6 lg:mb-10 rounded-lg bg-white p-6 shadow-sm ring-1 ring-slate-200">
            {/* Header with percentage and cancel button */}
            <div className="flex items-center justify-between mb-3">
              <span className="text-lg font-semibold text-slate-900">Refreshing data...</span>
              <div className="flex items-center gap-4">
                <span className="text-2xl font-bold text-primary-600">{snapshotStatus.progress}%</span>
                <button
                  onClick={handleCancelRefresh}
                  className="px-4 py-2 text-sm font-medium text-white bg-red-600 hover:bg-red-700 rounded-md shadow-sm transition-colors"
                  title="Cancel refresh"
                >
                  Cancel
                </button>
              </div>
            </div>

            {/* Progress bar */}
            <div className="h-3 w-full rounded-full bg-slate-100 mb-4">
              <div
                className="h-3 rounded-full bg-primary-600 transition-all duration-500 ease-out"
                style={{ width: `${snapshotStatus.progress}%` }}
              />
            </div>

            {/* Step indicator */}
            {snapshotStatus.totalSteps && snapshotStatus.completedSteps !== undefined && (
              <div className="flex items-center gap-2 mb-3">
                <span className="text-sm font-medium text-slate-600">
                  Step {snapshotStatus.completedSteps + 1} of {snapshotStatus.totalSteps}
                </span>
                {/* Visual step dots */}
                <div className="flex gap-1">
                  {Array.from({ length: snapshotStatus.totalSteps }).map((_, i) => (
                    <div
                      key={i}
                      className={`h-1.5 w-1.5 rounded-full transition-colors ${
                        i < snapshotStatus.completedSteps! ? 'bg-primary-600' : 'bg-slate-300'
                      }`}
                    />
                  ))}
                </div>
              </div>
            )}

            {/* Current operation */}
            {snapshotStatus.currentStep && (
              <p className="text-base font-medium text-slate-700 mb-3">
                {snapshotStatus.currentStep}
              </p>
            )}

            {/* Parallel Phases Display */}
            {snapshotStatus.parallelPhases && snapshotStatus.parallelPhases.length > 0 ? (
              <div className="space-y-3 mb-4">
                <h4 className="text-sm font-medium text-slate-700 mb-3">
                  Loading Data in Parallel
                </h4>
                {snapshotStatus.parallelPhases.map((phase) => (
                  <PhaseProgressBar
                    key={phase.phase}
                    phase={phase}
                    label={
                      phase.phase === 'mops' ? 'Planned Manufacturing Orders' :
                      phase.phase === 'mos' ? 'Manufacturing Orders' :
                      'Customer Order Lines'
                    }
                  />
                ))}
              </div>
            ) : (
              /* Fallback to legacy metrics display */
              <div className="flex flex-wrap gap-4 text-sm text-slate-600 mb-4">
                {snapshotStatus.coLinesProcessed !== undefined && snapshotStatus.coLinesProcessed > 0 && (
                  <div className="flex items-center gap-1">
                    <span className="font-medium">CO Lines:</span>
                    <span>{snapshotStatus.coLinesProcessed.toLocaleString()}</span>
                  </div>
                )}
                {snapshotStatus.mosProcessed !== undefined && snapshotStatus.mosProcessed > 0 && (
                  <div className="flex items-center gap-1">
                    <span className="font-medium">MOs:</span>
                    <span>{snapshotStatus.mosProcessed.toLocaleString()}</span>
                  </div>
                )}
                {snapshotStatus.mopsProcessed !== undefined && snapshotStatus.mopsProcessed > 0 && (
                  <div className="flex items-center gap-1">
                    <span className="font-medium">MOPs:</span>
                    <span>{snapshotStatus.mopsProcessed.toLocaleString()}</span>
                  </div>
                )}
              </div>
            )}

            {/* Parallel Detectors Display */}
            {snapshotStatus.parallelDetectors && snapshotStatus.parallelDetectors.length > 0 && (
              <div className="space-y-3 mb-4 pt-4 border-t border-slate-200">
                <h4 className="text-sm font-medium text-slate-700 mb-3">
                  Running Issue Detectors
                </h4>
                {snapshotStatus.parallelDetectors.map((detector) => (
                  <DetectorProgressBar
                    key={detector.detectorName}
                    detector={detector}
                  />
                ))}
              </div>
            )}

            {/* Performance Metrics */}
            {(snapshotStatus.recordsPerSecond || snapshotStatus.estimatedTimeRemaining) && (
              <div className="pt-3 border-t border-slate-200 flex flex-wrap gap-4 text-sm text-slate-600 mb-3">
                {snapshotStatus.recordsPerSecond && snapshotStatus.recordsPerSecond > 0 && (
                  <div className="flex items-center gap-1">
                    <span className="font-medium">Rate:</span>
                    <span>~{Math.round(snapshotStatus.recordsPerSecond)}/sec</span>
                  </div>
                )}
                {snapshotStatus.estimatedTimeRemaining && snapshotStatus.estimatedTimeRemaining > 0 && (
                  <div className="flex items-center gap-1">
                    <span className="font-medium">ETA:</span>
                    <span>~{Math.ceil(snapshotStatus.estimatedTimeRemaining)}s</span>
                  </div>
                )}
              </div>
            )}

            {/* Error notice if SSE unavailable */}
            {sseError && (
              <div className="mt-3 text-xs text-red-600 flex items-center gap-1">
                <ExclamationTriangleIcon className="h-3 w-3" />
                {sseError}
              </div>
            )}

            {/* Connection status indicator */}
            {!sseError && (
              <div className="mt-3 text-xs text-slate-500 flex items-center gap-1">
                <div className={`h-2 w-2 rounded-full ${isConnected ? 'bg-green-500' : 'bg-slate-400'}`} />
                {isConnected ? 'Live updates' : 'Connecting...'}
              </div>
            )}
          </div>
        )}

        {/* Refresh Cancelled Message */}
        {snapshotStatus?.status === 'cancelled' && (
          <div className="mb-6 lg:mb-10 rounded-lg bg-yellow-50 p-6 shadow-sm ring-1 ring-yellow-200">
            <div className="flex items-start gap-3">
              <ExclamationTriangleIcon className="h-6 w-6 text-yellow-600 flex-shrink-0 mt-0.5" />
              <div className="flex-1">
                <h3 className="text-lg font-semibold text-yellow-900 mb-2">Data Refresh Cancelled</h3>
                <p className="text-sm text-yellow-700">
                  The data refresh was cancelled and did not complete.
                </p>
              </div>
            </div>
          </div>
        )}

        {/* Refresh Failure Message */}
        {snapshotStatus?.status === 'failed' && (
          <div className="mb-6 lg:mb-10 rounded-lg bg-red-50 p-6 shadow-sm ring-1 ring-red-200">
            <div className="flex items-start gap-3">
              <ExclamationTriangleIcon className="h-6 w-6 text-red-600 flex-shrink-0 mt-0.5" />
              <div className="flex-1">
                <h3 className="text-lg font-semibold text-red-900 mb-2">Data Refresh Failed</h3>
                <p className="text-sm text-red-700 mb-3">
                  The data refresh encountered an error and could not complete.
                </p>
                {snapshotStatus.error && (
                  <div className="bg-white rounded p-3 text-sm text-slate-700 font-mono whitespace-pre-wrap break-words">
                    {snapshotStatus.error}
                  </div>
                )}
              </div>
            </div>
          </div>
        )}

        {/* Refresh Success Message */}
        {snapshotStatus?.status === 'completed' && currentJobId && (
          <div className="mb-6 lg:mb-10 rounded-lg bg-green-50 p-6 shadow-sm ring-1 ring-green-200">
            <div className="flex items-start gap-3">
              <CheckCircleIcon className="h-6 w-6 text-green-600 flex-shrink-0 mt-0.5" />
              <div className="flex-1">
                <h3 className="text-lg font-semibold text-green-900 mb-1">Data Refresh Completed</h3>
                <p className="text-sm text-green-700">
                  Successfully refreshed {(snapshotStatus.coLinesProcessed || 0).toLocaleString()} CO lines,{' '}
                  {(snapshotStatus.mosProcessed || 0).toLocaleString()} MOs, and{' '}
                  {(snapshotStatus.mopsProcessed || 0).toLocaleString()} MOPs.
                </p>
              </div>
            </div>
          </div>
        )}

        {/* Stats Grid */}
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4 mb-6 lg:gap-8 lg:mb-10">
          {stats.map((stat) => {
            const colors = getColorClasses(stat.color);
            const value = getStatValue(stat.key);
            const isWarning = stat.key === 'inconsistenciesCount' && typeof value === 'number' && value > 0;

            return (
              <Link
                key={stat.name}
                to={stat.href}
                onClick={(e) => stat.href === '#' && e.preventDefault()}
                className={`relative overflow-hidden rounded-lg bg-white p-4 sm:p-5 lg:p-8 shadow-sm transition-all duration-200 hover:shadow-md no-underline ${
                  isWarning ? 'ring-1 ring-warning-400 hover:ring-warning-500' : 'ring-1 ring-slate-200 hover:ring-primary-400'
                } ${stat.href === '#' ? 'cursor-default' : ''}`}
              >
                <div className="flex items-start justify-between gap-3">
                  <div className="min-w-0 flex-1">
                    <p className="text-sm font-medium text-slate-600">{stat.name}</p>
                    <p className={`mt-1 text-2xl sm:text-3xl font-bold tracking-tight ${isWarning ? 'text-warning-600' : 'text-slate-900'}`}>
                      {typeof value === 'number' ? value.toLocaleString() : value}
                    </p>

                    {/* Breakdown displays */}
                    {stat.key === 'currentFacility' && effectiveContext && (
                      <p className="mt-1 text-xs text-slate-500 font-mono">
                        {effectiveContext.company}/{effectiveContext.division}/{effectiveContext.facility}/{effectiveContext.warehouse}
                      </p>
                    )}

                    {stat.key === 'totalProductionOrders' && summary && (
                      <p className="mt-1 text-xs text-slate-500">
                        {summary.totalManufacturingOrders.toLocaleString()} MOs, {summary.totalPlannedOrders.toLocaleString()} MOPs
                      </p>
                    )}

                    {stat.key === 'totalCustomerOrderLines' && (
                      <p className="mt-1 text-xs text-slate-500">
                        Reserved lines only
                      </p>
                    )}
                  </div>
                  <div className={`shrink-0 rounded-lg p-2 sm:p-2.5 ${colors.bg}`}>
                    <stat.icon className={`h-5 w-5 sm:h-6 sm:w-6 ${colors.icon}`} />
                  </div>
                </div>
                {isWarning && (
                  <div className="absolute bottom-0 left-0 right-0 h-1 bg-warning-400" />
                )}
              </Link>
            );
          })}
        </div>

        {/* Data Scope Disclosure */}
        {summary && effectiveContext && (
          <div className="mb-6 lg:mb-10 rounded-xl bg-blue-50 shadow-sm ring-1 ring-blue-200">
            <div
              className="px-6 py-4 cursor-pointer"
              onClick={() => setScopeInfoExpanded(!scopeInfoExpanded)}
            >
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div className="rounded-md bg-blue-100 p-2">
                    <svg className="h-5 w-5 text-blue-600" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" d="M11.25 11.25l.041-.02a.75.75 0 011.063.852l-.708 2.836a.75.75 0 001.063.853l.041-.021M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9-3.75h.008v.008H12V8.25z" />
                    </svg>
                  </div>
                  <div>
                    <h3 className="text-sm font-semibold text-blue-900">Data Scope Information</h3>
                    <p className="text-xs text-blue-700 mt-0.5">
                      All data is scoped to facility: <span className="font-mono font-bold">{effectiveContext.facility}</span>
                    </p>
                  </div>
                </div>
                <svg
                  className={`h-5 w-5 text-blue-600 transition-transform ${scopeInfoExpanded ? 'rotate-180' : ''}`}
                  fill="none"
                  viewBox="0 0 24 24"
                  strokeWidth={2}
                  stroke="currentColor"
                >
                  <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 8.25l-7.5 7.5-7.5-7.5" />
                </svg>
              </div>
            </div>

            {scopeInfoExpanded && (
              <div className="px-6 pb-4 border-t border-blue-200 pt-4">
                <h4 className="text-xs font-semibold text-blue-900 mb-2">Data Filters Applied:</h4>
                <dl className="space-y-2 text-xs text-blue-800">
                  <div className="flex gap-2">
                    <dt className="font-semibold min-w-[140px]">Manufacturing Orders:</dt>
                    <dd>Only preliminary/released (WHST ≤ 20)</dd>
                  </div>
                  <div className="flex gap-2">
                    <dt className="font-semibold min-w-[140px]">Planned Orders:</dt>
                    <dd>Only firmed planned orders (PSTS = 20)</dd>
                  </div>
                  <div className="flex gap-2">
                    <dt className="font-semibold min-w-[140px]">Customer Order Lines:</dt>
                    <dd>Only reserved lines (ORST 20-29)</dd>
                  </div>
                </dl>
                <div className="mt-3 pt-3 border-t border-blue-200">
                  <p className="text-xs text-blue-700">
                    <strong>Note:</strong> This ensures you're viewing only the active planning data that requires attention.
                    Completed orders, delivered items, and preliminary quotations are excluded.
                  </p>
                </div>
              </div>
            )}
          </div>
        )}

        {/* Issue Breakdown - Facility > Warehouse > Detector */}
        {issueSummary && issueSummary.total > 0 && (
          <div className="mb-6 lg:mb-10 rounded-xl bg-white shadow-sm ring-1 ring-slate-200">
            <div className="px-6 py-4 border-b border-slate-200">
              <div className="flex items-center gap-3">
                <ExclamationTriangleIcon className="h-5 w-5 text-warning-600" />
                <h2 className="text-base font-semibold text-slate-900">
                  Issue Breakdown
                </h2>
                <span className="text-sm text-slate-500">
                  ({issueSummary.total} {issueSummary.total === 1 ? 'issue' : 'issues'})
                </span>
              </div>
            </div>
            <div className="p-6">
              <IssueBreakdownHierarchy summary={issueSummary} />
            </div>
          </div>
        )}

        {/* Secondary Content */}
        <div className="grid grid-cols-1 gap-4 lg:grid-cols-2 lg:gap-8">
          {/* Last Refresh Info */}
          <div className="rounded-lg bg-white p-4 sm:p-5 lg:p-8 shadow-sm ring-1 ring-slate-200">
            <div className="flex items-center gap-2 mb-3">
              <div className="rounded-md bg-slate-100 p-1.5">
                <ArrowTrendingUpIcon className="h-4 w-4 text-slate-600" />
              </div>
              <h2 className="text-base font-semibold text-slate-900">Data Status</h2>
            </div>
            <dl className="space-y-2">
              <div className="flex justify-between">
                <dt className="text-sm text-slate-500">Last refreshed</dt>
                <dd className="text-sm font-medium text-slate-900">
                  {summary?.lastRefresh
                    ? new Date(summary.lastRefresh).toLocaleString()
                    : 'Never'}
                </dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-sm text-slate-500">Status</dt>
                <dd className="text-sm font-medium">
                  <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
                    snapshotStatus?.status === 'running'
                      ? 'bg-primary-100 text-primary-700'
                      : 'bg-success-100 text-success-700'
                  }`}>
                    {snapshotStatus?.status === 'running' ? 'Refreshing' : 'Ready'}
                  </span>
                </dd>
              </div>
            </dl>
          </div>

          {/* Quick Actions */}
          <div className="rounded-lg bg-white p-4 sm:p-5 lg:p-8 shadow-sm ring-1 ring-slate-200">
            <h2 className="text-base font-semibold text-slate-900 mb-3">Quick Actions</h2>
            <div className="grid grid-cols-1 gap-2">
              <Link
                to="/inconsistencies"
                className="flex items-center gap-2 rounded-md border border-slate-200 p-2.5 transition-colors hover:bg-slate-50 no-underline"
              >
                <ExclamationTriangleIcon className="h-4 w-4 text-slate-400" />
                <span className="text-sm font-medium text-slate-700">Review Inconsistencies</span>
              </Link>
            </div>
          </div>
        </div>
      </div>
      <ToastContainer toasts={toast.toasts} onClose={toast.removeToast} />
    </AppLayout>
  );
};

export default Dashboard;
