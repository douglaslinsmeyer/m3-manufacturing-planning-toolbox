import React, { useEffect, useState } from 'react';
import { AppLayout } from '../components/AppLayout';
import { buildM3BookmarkURL, M3Config } from '../utils/m3Links';

interface Issue {
  id: number;
  detectorType: string;
  facility: string;
  warehouse?: string;
  issueKey: string;
  productionOrderNumber?: string;
  productionOrderType?: string;
  coNumber?: string;
  coLine?: string;
  coSuffix?: string;
  detectedAt: string;
  issueData: Record<string, any>;
}

interface IssueSummary {
  total: number;
  by_detector: Record<string, number>;
  by_facility: Record<string, number>;
  by_warehouse: Record<string, number>;
}

function ExclamationTriangleIcon({ className }: { className?: string }) {
  return (
    <svg className={className} fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z" />
    </svg>
  );
}

const DETECTOR_LABELS: Record<string, string> = {
  'unlinked_production_orders': 'Unlinked Production Orders',
  'start_date_mismatch': 'Start Date Mismatches',
  'production_timing': 'Production Timing Issues',
};

const Inconsistencies: React.FC = () => {
  const [summary, setSummary] = useState<IssueSummary | null>(null);
  const [issues, setIssues] = useState<Issue[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedDetector, setSelectedDetector] = useState<string>('');
  const [selectedWarehouse, setSelectedWarehouse] = useState<string>('');
  const [m3Config, setM3Config] = useState<M3Config | null>(null);

  // Initialize filters from URL on mount and fetch data
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const detector = params.get('detector');
    const warehouse = params.get('warehouse');

    if (detector) setSelectedDetector(detector);
    if (warehouse) setSelectedWarehouse(warehouse);

    // Fetch config and summary once on mount
    fetchM3Config();
    fetchSummary();
  }, []);

  // Fetch issues when filters change
  useEffect(() => {
    fetchIssues();
  }, [selectedDetector, selectedWarehouse]);

  const fetchM3Config = async () => {
    try {
      const response = await fetch('/api/m3-config', {
        credentials: 'include',
      });
      const data = await response.json();
      setM3Config(data);
    } catch (error) {
      console.error('Failed to fetch M3 config:', error);
    }
  };

  const fetchSummary = async () => {
    try {
      const response = await fetch('/api/issues/summary', {
        credentials: 'include',
      });
      const data = await response.json();
      setSummary(data);
    } catch (error) {
      console.error('Failed to fetch issue summary:', error);
    }
  };

  const fetchIssues = async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams();
      if (selectedDetector) params.append('detector_type', selectedDetector);
      if (selectedWarehouse) params.append('warehouse', selectedWarehouse);
      params.append('limit', '100');

      const response = await fetch(`/api/issues?${params}`, {
        credentials: 'include',
      });
      const data = await response.json();
      setIssues(data);
    } catch (error) {
      console.error('Failed to fetch issues:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <AppLayout>
      <div className="px-4 py-6 sm:px-6 lg:px-12 lg:py-10">
        {/* Page Header */}
        <div className="mb-6 lg:mb-10">
          <div className="flex items-center gap-3">
            <div className="rounded-lg bg-warning-100 p-2">
              <ExclamationTriangleIcon className="h-6 w-6 text-warning-600" />
            </div>
            <div>
              <h1 className="text-2xl font-semibold text-slate-900">Planning Issues</h1>
              <p className="mt-1 text-sm text-slate-500">
                Detected data quality and planning problems
              </p>
            </div>
          </div>
        </div>

        {/* Summary Cards */}
        {summary && (
          <div className="mb-6 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <div className="rounded-xl bg-white p-6 shadow-sm ring-1 ring-slate-200">
              <div className="text-sm font-medium text-slate-500">Total Issues</div>
              <div className="mt-2 text-3xl font-semibold text-slate-900">{summary.total}</div>
            </div>

            {Object.entries(summary.by_detector).slice(0, 3).map(([detector, count]) => (
              <div key={detector} className="rounded-xl bg-white p-6 shadow-sm ring-1 ring-slate-200">
                <div className="text-sm font-medium text-slate-500">
                  {DETECTOR_LABELS[detector] || detector}
                </div>
                <div className="mt-2 text-3xl font-semibold text-slate-900">{count}</div>
              </div>
            ))}
          </div>
        )}

        {/* Filters */}
        <div className="mb-6 rounded-xl bg-white p-6 shadow-sm ring-1 ring-slate-200">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div>
              <label className="block text-sm font-medium text-slate-700 mb-2">
                Issue Type
              </label>
              <select
                value={selectedDetector}
                onChange={(e) => setSelectedDetector(e.target.value)}
                className="block w-full rounded-lg border-slate-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
              >
                <option value="">All Types</option>
                {summary && Object.keys(summary.by_detector).map((detector) => (
                  <option key={detector} value={detector}>
                    {DETECTOR_LABELS[detector] || detector} ({summary.by_detector[detector]})
                  </option>
                ))}
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-slate-700 mb-2">
                Warehouse
              </label>
              <select
                value={selectedWarehouse}
                onChange={(e) => setSelectedWarehouse(e.target.value)}
                className="block w-full rounded-lg border-slate-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
              >
                <option value="">All Warehouses</option>
                {summary && summary.by_warehouse && Object.keys(summary.by_warehouse).map((warehouse) => (
                  <option key={warehouse} value={warehouse}>
                    {warehouse} ({summary.by_warehouse[warehouse]})
                  </option>
                ))}
              </select>
            </div>
          </div>
        </div>

        {/* Issues List */}
        <div className="rounded-xl bg-white shadow-sm ring-1 ring-slate-200">
          {loading ? (
            <div className="p-12 text-center">
              <div className="h-8 w-8 mx-auto animate-spin rounded-full border-4 border-primary-200 border-t-primary-600"></div>
              <div className="mt-4 text-slate-500">Loading issues...</div>
            </div>
          ) : issues.length === 0 ? (
            <div className="p-12 text-center">
              <ExclamationTriangleIcon className="mx-auto h-12 w-12 text-slate-300" />
              <h3 className="mt-4 text-lg font-medium text-slate-900">No Issues Found</h3>
              <p className="mt-2 text-sm text-slate-500">
                No planning issues detected in the current snapshot.
              </p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-slate-200">
                <thead className="bg-slate-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-slate-500">
                      Issue Type
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-slate-500">
                      Affected Orders
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-slate-500">
                      Facility
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-slate-500">
                      Details
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-slate-500">
                      Detected
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-200 bg-white">
                  {issues.map((issue) => (
                    <tr key={issue.id} className="hover:bg-slate-50">
                      <td className="whitespace-nowrap px-6 py-4 text-sm font-medium text-slate-900">
                        {DETECTOR_LABELS[issue.detectorType] || issue.detectorType}
                      </td>
                      <td className="px-6 py-4 text-sm text-slate-500">
                        {issue.productionOrderNumber && (
                          <div>
                            {m3Config ? (
                              <a
                                href={buildM3BookmarkURL(
                                  m3Config,
                                  issue.productionOrderType as 'MO' | 'MOP',
                                  issue.productionOrderNumber,
                                  issue.issueData.company || '100',
                                  issue.facility,
                                  issue.issueData.product_number
                                )}
                                target="_blank"
                                rel="noopener noreferrer"
                                className="font-medium text-primary-600 hover:text-primary-700 hover:underline"
                              >
                                {issue.productionOrderNumber}
                              </a>
                            ) : (
                              <span className="font-medium">{issue.productionOrderNumber}</span>
                            )}
                            <span className="ml-2 text-xs text-slate-400">
                              ({issue.productionOrderType})
                            </span>
                          </div>
                        )}
                        {issue.coNumber && (
                          <div className="text-xs text-slate-400">
                            CO: {issue.coNumber}-{issue.coLine}
                          </div>
                        )}
                      </td>
                      <td className="whitespace-nowrap px-6 py-4 text-sm text-slate-500">
                        {issue.facility}
                        {issue.warehouse && <div className="text-xs text-slate-400">{issue.warehouse}</div>}
                      </td>
                      <td className="px-6 py-4 text-sm text-slate-500">
                        <IssueDetailsCell issueData={issue.issueData} detectorType={issue.detectorType} />
                      </td>
                      <td className="whitespace-nowrap px-6 py-4 text-sm text-slate-500">
                        {new Date(issue.detectedAt).toLocaleDateString()}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>
    </AppLayout>
  );
};

// Helper component to display issue-specific details
const IssueDetailsCell: React.FC<{ issueData: Record<string, any>; detectorType: string }> = ({ issueData, detectorType }) => {
  if (detectorType === 'unlinked_production_orders') {
    return (
      <div className="text-xs">
        <div>Item: {issueData.item_number}</div>
      </div>
    );
  }

  if (detectorType === 'start_date_mismatch') {
    return (
      <div className="text-xs">
        {issueData.dates && issueData.dates.length > 0 && (
          <div>Dates: {issueData.dates.join(', ')}</div>
        )}
      </div>
    );
  }

  if (detectorType === 'production_timing') {
    return (
      <div className="text-xs">
        <div className="font-medium text-warning-600">
          {issueData.timing_issue === 'too_early' ? 'Starts Too Early' : 'Starts Too Late'}
        </div>
        <div>Start: {issueData.start_date}</div>
        <div>Delivery: {issueData.delivery_date}</div>
        <div>Difference: {issueData.days_difference} days</div>
      </div>
    );
  }

  return <div className="text-xs">{JSON.stringify(issueData)}</div>;
};

export default Inconsistencies;
