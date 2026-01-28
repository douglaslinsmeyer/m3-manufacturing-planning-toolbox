import React, { useEffect, useState } from 'react';
import api from '../services/api';
import type { Anomaly, AnomalySummary } from '../types';
import { AppLayout } from '../components/AppLayout';
import { ToastContainer } from '../components/Toast';
import { useToast } from '../hooks/useToast';
import {
  AlertTriangle,
  AlertCircle,
  Info,
  RefreshCw,
  ChevronDown,
  ChevronUp,
  Check,
  CheckCircle2,
} from 'lucide-react';

const Anomalies: React.FC = () => {
  const [anomalies, setAnomalies] = useState<Anomaly[]>([]);
  const [summary, setSummary] = useState<AnomalySummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [severityFilter, setSeverityFilter] = useState<string>('');
  const [expandedAnomaly, setExpandedAnomaly] = useState<number | null>(null);
  const [acknowledging, setAcknowledging] = useState<number | null>(null);
  const [resolving, setResolving] = useState<number | null>(null);
  const [currentPage, setCurrentPage] = useState<number>(1);
  const [pageSize, setPageSize] = useState<number>(50);
  const [totalCount, setTotalCount] = useState<number>(0);
  const [totalPages, setTotalPages] = useState<number>(0);
  const toast = useToast();

  // Reset to page 1 when filter changes
  useEffect(() => {
    if (currentPage !== 1) {
      setCurrentPage(1);
    }
  }, [severityFilter]);

  // Fetch when pagination or filter changes
  useEffect(() => {
    loadAnomalies();
  }, [severityFilter, currentPage, pageSize]);

  const loadAnomalies = async () => {
    try {
      setLoading(true);
      const [anomaliesData, summaryData] = await Promise.all([
        api.listAnomalies({
          severity: severityFilter || undefined,
          page: currentPage,
          pageSize: pageSize,
        }),
        api.getAnomalySummary(),
      ]);
      setAnomalies(anomaliesData.data);
      setTotalCount(anomaliesData.pagination.totalCount);
      setTotalPages(anomaliesData.pagination.totalPages);
      setSummary(summaryData);
      setError(null);
    } catch (err: any) {
      console.error('Failed to load anomalies:', err);
      setError(err.response?.data?.error || err.message || 'Failed to load anomalies');
    } finally {
      setLoading(false);
    }
  };

  const getSeverityIcon = (severity: string) => {
    switch (severity) {
      case 'critical':
        return <AlertTriangle className="w-6 h-6 text-red-600" />;
      case 'warning':
        return <AlertCircle className="w-6 h-6 text-yellow-600" />;
      case 'info':
        return <Info className="w-6 h-6 text-blue-600" />;
      default:
        return <Info className="w-6 h-6 text-gray-600" />;
    }
  };

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical':
        return 'border-red-500 bg-red-50';
      case 'warning':
        return 'border-yellow-500 bg-yellow-50';
      case 'info':
        return 'border-blue-500 bg-blue-50';
      default:
        return 'border-gray-500 bg-gray-50';
    }
  };

  const getSeverityBadge = (severity: string) => {
    switch (severity) {
      case 'critical':
        return 'bg-red-100 text-red-800';
      case 'warning':
        return 'bg-yellow-100 text-yellow-800';
      case 'info':
        return 'bg-blue-100 text-blue-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const getDetectorDisplayName = (detectorType: string) => {
    if (!detectorType) return 'Unknown';
    return detectorType
      .replace('anomaly_', '')
      .split('_')
      .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
      .join(' ');
  };

  const toggleAnomaly = (id: number) => {
    setExpandedAnomaly(expandedAnomaly === id ? null : id);
  };

  const handleAcknowledge = async (id: number, e: React.MouseEvent) => {
    e.stopPropagation();
    try {
      setAcknowledging(id);
      await api.acknowledgeAnomaly(id);
      await loadAnomalies(); // Reload to show updated status
      toast.success('Anomaly acknowledged successfully');
    } catch (err: any) {
      console.error('Failed to acknowledge anomaly:', err);
      toast.error('Failed to acknowledge anomaly. Please try again.');
    } finally {
      setAcknowledging(null);
    }
  };

  const handleResolve = async (id: number, e: React.MouseEvent) => {
    e.stopPropagation();
    try {
      setResolving(id);
      await api.resolveAnomaly(id);
      await loadAnomalies(); // Reload to show updated status
      toast.success('Anomaly resolved successfully');
    } catch (err: any) {
      console.error('Failed to resolve anomaly:', err);
      toast.error('Failed to resolve anomaly. Please try again.');
    } finally {
      setResolving(null);
    }
  };

  if (loading && !anomalies.length) {
    return (
      <AppLayout>
        <div className="flex items-center justify-center h-96">
          <div className="flex flex-col items-center gap-4">
            <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary-200 border-t-primary-600" />
            <p className="text-sm text-slate-500">Loading anomalies...</p>
          </div>
        </div>
      </AppLayout>
    );
  }

  if (error) {
    return (
      <AppLayout>
        <div className="flex items-center justify-center h-96">
          <div className="flex flex-col items-center gap-4">
            <AlertTriangle className="h-12 w-12 text-red-400" />
            <div className="text-center">
              <h3 className="text-lg font-medium text-slate-900 mb-2">Error Loading Anomalies</h3>
              <p className="text-sm text-red-600">{error}</p>
            </div>
          </div>
        </div>
      </AppLayout>
    );
  }

  const criticalCount = summary?.by_severity?.critical || 0;
  const warningCount = summary?.by_severity?.warning || 0;
  const infoCount = summary?.by_severity?.info || 0;

  return (
    <AppLayout>
      <ToastContainer toasts={toast.toasts} onClose={toast.removeToast} />
      <div className="px-4 py-6 sm:px-6 lg:px-12 lg:py-10">
      {/* Page Header */}
      <div className="mb-6 lg:mb-10">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div className="flex items-center gap-3">
            <div className="rounded-lg bg-red-100 p-2">
              <AlertTriangle className="h-6 w-6 text-red-600" />
            </div>
            <div>
              <h1 className="text-2xl font-semibold text-slate-900">Anomaly Detection</h1>
              <p className="mt-1 text-sm text-slate-500">
                Statistical anomalies and unusual patterns in production data
              </p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={loadAnomalies}
              disabled={loading}
              className="inline-flex items-center gap-2 rounded-md bg-primary-600 px-3 py-1.5 text-sm font-medium text-white shadow-sm transition-all hover:bg-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
              {loading ? 'Refreshing...' : 'Refresh'}
            </button>
          </div>
        </div>
      </div>

        {/* Summary Stats */}
        {summary && (
          <div className="mb-6 grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4 lg:gap-8">
            <div className="rounded-xl bg-white p-6 shadow-sm ring-1 ring-slate-200">
              <div className="text-sm font-medium text-slate-500">Total Anomalies</div>
              <div className="mt-2 text-3xl font-semibold text-slate-900">{summary.total}</div>
            </div>
            <div className="rounded-xl bg-red-50 p-6 shadow-sm ring-1 ring-red-200">
              <div className="flex items-center gap-2 mb-1">
                <AlertTriangle className="w-4 h-4 text-red-600" />
                <div className="text-sm font-medium text-red-600">Critical</div>
              </div>
              <div className="mt-2 text-3xl font-semibold text-red-700">{criticalCount}</div>
            </div>
            <div className="rounded-xl bg-yellow-50 p-6 shadow-sm ring-1 ring-yellow-200">
              <div className="flex items-center gap-2 mb-1">
                <AlertCircle className="w-4 h-4 text-yellow-600" />
                <div className="text-sm font-medium text-yellow-600">Warning</div>
              </div>
              <div className="mt-2 text-3xl font-semibold text-yellow-700">{warningCount}</div>
            </div>
            <div className="rounded-xl bg-blue-50 p-6 shadow-sm ring-1 ring-blue-200">
              <div className="flex items-center gap-2 mb-1">
                <Info className="w-4 h-4 text-blue-600" />
                <div className="text-sm font-medium text-blue-600">Info</div>
              </div>
              <div className="mt-2 text-3xl font-semibold text-blue-700">{infoCount}</div>
            </div>
          </div>
        )}

        {/* Filters */}
        <div className="mb-6 rounded-xl bg-white p-6 shadow-sm ring-1 ring-slate-200">
          <div className="flex items-center gap-4">
            <label className="block text-sm font-medium text-slate-700">Filter by Severity:</label>
            <select
              value={severityFilter}
              onChange={(e) => setSeverityFilter(e.target.value)}
              className="block rounded-lg border-slate-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 text-sm"
            >
              <option value="">All Severities</option>
              <option value="critical">Critical</option>
              <option value="warning">Warning</option>
              <option value="info">Info</option>
            </select>
          </div>
        </div>

      {/* Anomalies List */}
      <div className="rounded-xl bg-white shadow-sm ring-1 ring-slate-200">
        {anomalies.length === 0 ? (
          <div className="p-12 text-center">
            <Info className="mx-auto h-12 w-12 text-slate-300" />
            <h3 className="mt-4 text-lg font-medium text-slate-900">No Anomalies Detected</h3>
            <p className="mt-2 text-sm text-slate-500">
              No statistical anomalies found in the current production data.
            </p>
          </div>
        ) : (
          <div className="p-6 space-y-4">
          {anomalies.map((anomaly) => (
            <div
              key={anomaly.id}
              className={`border-l-4 rounded-xl shadow-sm ring-1 ring-slate-200 overflow-hidden ${getSeverityColor(
                anomaly.severity
              )}`}
            >
              <div
                className="bg-white p-4 cursor-pointer hover:bg-gray-50 transition-colors"
                onClick={() => toggleAnomaly(anomaly.id)}
              >
                <div className="flex items-start justify-between">
                  <div className="flex items-start gap-3 flex-1">
                    <div className="flex-shrink-0 mt-1">
                      {getSeverityIcon(anomaly.severity)}
                    </div>
                    <div className="flex-1">
                      <div className="flex items-center gap-2 mb-2">
                        <span
                          className={`px-2 py-1 text-xs font-semibold rounded ${getSeverityBadge(
                            anomaly.severity
                          )}`}
                        >
                          {anomaly.severity.toUpperCase()}
                        </span>
                        <span className="text-sm text-gray-600">
                          {getDetectorDisplayName(anomaly.detectorType)}
                        </span>
                        {anomaly.status && anomaly.status !== 'active' && (
                          <span className="px-2 py-1 text-xs font-medium rounded bg-gray-100 text-gray-700">
                            {anomaly.status.toUpperCase()}
                          </span>
                        )}
                      </div>
                      <div className="text-sm text-gray-900 mb-2">
                        <strong className="font-semibold">
                          {anomaly.entityType === 'product'
                            ? `Product ${anomaly.entityId}`
                            : anomaly.entityType === 'warehouse'
                            ? `Warehouse ${anomaly.entityId}`
                            : 'System-wide'}
                        </strong>
                        {anomaly.warehouse && ` in warehouse ${anomaly.warehouse}`}
                      </div>
                      <div className="flex items-center gap-4 text-sm text-gray-600">
                        <span>Affected: {anomaly.affectedCount?.toLocaleString() || 0} records</span>
                        <span>
                          Threshold: {anomaly.thresholdValue || 'N/A'} | Actual:{' '}
                          {anomaly.actualValue?.toFixed(2) || 'N/A'}
                        </span>
                      </div>
                    </div>
                  </div>
                  <button className="ml-4">
                    {expandedAnomaly === anomaly.id ? (
                      <ChevronUp className="w-5 h-5 text-gray-500" />
                    ) : (
                      <ChevronDown className="w-5 h-5 text-gray-500" />
                    )}
                  </button>
                </div>
              </div>

              {/* Expanded Details */}
              {expandedAnomaly === anomaly.id && (
                <div className="bg-white border-t border-gray-200 p-4">
                  <h4 className="text-sm font-semibold text-gray-700 mb-3">Anomaly Details</h4>
                  <div className="grid grid-cols-2 gap-4 text-sm mb-4">
                    {Object.entries(anomaly.metrics || {}).map(([key, value]) => (
                      <div key={key} className="flex">
                        <span className="font-medium text-gray-700 w-40">
                          {key.replace(/_/g, ' ').replace(/\b\w/g, (l) => l.toUpperCase())}:
                        </span>
                        <span className="text-gray-900">
                          {typeof value === 'number'
                            ? value.toLocaleString()
                            : typeof value === 'object'
                            ? JSON.stringify(value)
                            : String(value)}
                        </span>
                      </div>
                    ))}
                  </div>

                  {/* Action Buttons */}
                  <div className="flex items-center gap-2 pt-4 border-t border-gray-200">
                    {anomaly.status === 'active' && (
                      <>
                        <button
                          onClick={(e) => handleAcknowledge(anomaly.id, e)}
                          disabled={acknowledging === anomaly.id}
                          className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white text-sm rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                          {acknowledging === anomaly.id ? (
                            <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                          ) : (
                            <Check className="w-4 h-4" />
                          )}
                          Acknowledge
                        </button>
                        <button
                          onClick={(e) => handleResolve(anomaly.id, e)}
                          disabled={resolving === anomaly.id}
                          className="flex items-center gap-2 px-4 py-2 bg-green-600 text-white text-sm rounded-lg hover:bg-green-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                          {resolving === anomaly.id ? (
                            <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                          ) : (
                            <CheckCircle2 className="w-4 h-4" />
                          )}
                          Resolve
                        </button>
                      </>
                    )}
                    {anomaly.status === 'acknowledged' && (
                      <button
                        onClick={(e) => handleResolve(anomaly.id, e)}
                        disabled={resolving === anomaly.id}
                        className="flex items-center gap-2 px-4 py-2 bg-green-600 text-white text-sm rounded-lg hover:bg-green-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                      >
                        {resolving === anomaly.id ? (
                          <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                        ) : (
                          <CheckCircle2 className="w-4 h-4" />
                        )}
                        Mark as Resolved
                      </button>
                    )}
                  </div>

                  <div className="mt-4 pt-4 border-t border-gray-200 text-xs text-gray-500">
                    Detected: {new Date(anomaly.detectedAt).toLocaleString()}
                    {anomaly.acknowledgedAt && (
                      <> • Acknowledged: {new Date(anomaly.acknowledgedAt).toLocaleString()}{anomaly.acknowledgedBy && ` by ${anomaly.acknowledgedBy}`}</>
                    )}
                    {anomaly.resolvedAt && (
                      <> • Resolved: {new Date(anomaly.resolvedAt).toLocaleString()}{anomaly.resolvedBy && ` by ${anomaly.resolvedBy}`}</>
                    )}
                  </div>
                </div>
              )}
            </div>
          ))}
          </div>
        )}

        {/* Pagination Controls */}
        {!loading && anomalies.length > 0 && (
          <div className="px-6 py-4 border-t border-slate-200 bg-slate-50">
            <div className="flex items-center justify-between">
              {/* Left: Page size selector */}
              <div className="flex items-center gap-2">
                <span className="text-sm text-slate-700">Show</span>
                <select
                  value={pageSize}
                  onChange={(e) => {
                    setPageSize(Number(e.target.value));
                    setCurrentPage(1);
                  }}
                  className="rounded-md border-slate-300 text-sm focus:border-primary-500 focus:ring-primary-500"
                >
                  <option value={25}>25</option>
                  <option value={50}>50</option>
                  <option value={100}>100</option>
                  <option value={200}>200</option>
                </select>
                <span className="text-sm text-slate-700">per page</span>
              </div>

              {/* Center: Page info */}
              <div className="text-sm text-slate-700">
                Showing {Math.min((currentPage - 1) * pageSize + 1, totalCount).toLocaleString()} to {Math.min(currentPage * pageSize, totalCount).toLocaleString()} of {totalCount.toLocaleString()} anomalies
              </div>

              {/* Right: Page navigation */}
              <div className="flex items-center gap-2">
                <button
                  onClick={() => setCurrentPage(1)}
                  disabled={currentPage === 1}
                  className="px-3 py-1.5 text-sm border border-slate-300 rounded-md disabled:opacity-50 disabled:cursor-not-allowed hover:bg-white hover:shadow-sm"
                >
                  First
                </button>
                <button
                  onClick={() => setCurrentPage(currentPage - 1)}
                  disabled={currentPage === 1}
                  className="px-3 py-1.5 text-sm border border-slate-300 rounded-md disabled:opacity-50 disabled:cursor-not-allowed hover:bg-white hover:shadow-sm"
                >
                  Previous
                </button>
                <span className="px-3 py-1.5 text-sm text-slate-700">
                  Page {currentPage} of {totalPages}
                </span>
                <button
                  onClick={() => setCurrentPage(currentPage + 1)}
                  disabled={currentPage === totalPages}
                  className="px-3 py-1.5 text-sm border border-slate-300 rounded-md disabled:opacity-50 disabled:cursor-not-allowed hover:bg-white hover:shadow-sm"
                >
                  Next
                </button>
                <button
                  onClick={() => setCurrentPage(totalPages)}
                  disabled={currentPage === totalPages}
                  className="px-3 py-1.5 text-sm border border-slate-300 rounded-md disabled:opacity-50 disabled:cursor-not-allowed hover:bg-white hover:shadow-sm"
                >
                  Last
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
      </div>
    </AppLayout>
  );
};

export default Anomalies;
