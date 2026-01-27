import React, { useEffect, useState } from 'react';
import { AppLayout } from '../components/AppLayout';
import { buildM3BookmarkURL, M3Config } from '../utils/m3Links';
import { api } from '../services/api';
import { ConfirmModal } from '../components/ConfirmModal';
import { JointDeliveryDetailModal } from '../components/JointDeliveryDetailModal';
import { ToastContainer } from '../components/Toast';
import { useToast } from '../hooks/useToast';

interface Issue {
  id: number;
  detectorType: string;
  facility: string;
  warehouse?: string;
  issueKey: string;
  productionOrderNumber?: string;
  productionOrderType?: string;
  moTypeDescription?: string;
  coNumber?: string;
  coLine?: string;
  coSuffix?: string;
  detectedAt: string;
  issueData: Record<string, any>;
  isIgnored?: boolean;
}

interface Detector {
  name: string;
  label: string;
  description: string;
  enabled: boolean;
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

function InformationCircleIcon({ className }: { className?: string }) {
  return (
    <svg className={className} fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" d="M11.25 11.25l.041-.02a.75.75 0 011.063.852l-.708 2.836a.75.75 0 001.063.853l.041-.021M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9-3.75h.008v.008H12V8.25z" />
    </svg>
  );
}

// Format M3 date (YYYYMMDD integer) to readable format
function formatM3Date(dateStr: string | number): string {
  if (!dateStr) return '';
  const str = dateStr.toString();
  if (str.length !== 8) return str;
  const year = str.substring(0, 4);
  const month = str.substring(4, 6);
  const day = str.substring(6, 8);
  return `${year}-${month}-${day}`;
}

// Format M3 date as relative time (e.g., "in 3 days", "2 days ago")
function formatM3DateRelative(dateStr: string | number): { relative: string; absolute: string } {
  if (!dateStr) return { relative: '', absolute: '' };

  const str = dateStr.toString();
  if (str.length !== 8) return { relative: str, absolute: str };

  const year = parseInt(str.substring(0, 4));
  const month = parseInt(str.substring(4, 6)) - 1; // JS months are 0-indexed
  const day = parseInt(str.substring(6, 8));

  const date = new Date(year, month, day);
  const now = new Date();
  now.setHours(0, 0, 0, 0); // Reset to start of day for fair comparison

  const diffTime = date.getTime() - now.getTime();
  const diffDays = Math.round(diffTime / (1000 * 60 * 60 * 24));

  let relative = '';
  if (diffDays === 0) {
    relative = 'today';
  } else if (diffDays === 1) {
    relative = 'tomorrow';
  } else if (diffDays === -1) {
    relative = 'yesterday';
  } else if (diffDays > 0) {
    relative = `in ${diffDays} days`;
  } else {
    relative = `${Math.abs(diffDays)} days ago`;
  }

  const absolute = date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  });

  return { relative, absolute };
}

const Inconsistencies: React.FC = () => {
  const [summary, setSummary] = useState<IssueSummary | null>(null);
  const [issues, setIssues] = useState<Issue[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedDetector, setSelectedDetector] = useState<string>('');
  const [selectedWarehouse, setSelectedWarehouse] = useState<string>('');
  const [showIgnored, setShowIgnored] = useState<boolean>(false);
  const [m3Config, setM3Config] = useState<M3Config | null>(null);
  const [detectorLabels, setDetectorLabels] = useState<Record<string, string>>({});
  const [deleteModalOpen, setDeleteModalOpen] = useState(false);
  const [issueToDelete, setIssueToDelete] = useState<Issue | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);
  const [closeMOModalOpen, setCloseMOModalOpen] = useState(false);
  const [issueToClose, setIssueToClose] = useState<Issue | null>(null);
  const [isClosing, setIsClosing] = useState(false);
  const [alignModalOpen, setAlignModalOpen] = useState(false);
  const [issueToAlign, setIssueToAlign] = useState<Issue | null>(null);
  const [isAligning, setIsAligning] = useState(false);
  const [detailModalOpen, setDetailModalOpen] = useState(false);
  const [selectedIssueForDetail, setSelectedIssueForDetail] = useState<Issue | null>(null);
  const [isInitialized, setIsInitialized] = useState(false);
  const [currentPage, setCurrentPage] = useState<number>(1);
  const [pageSize, setPageSize] = useState<number>(50);
  const [totalCount, setTotalCount] = useState<number>(0);
  const [totalPages, setTotalPages] = useState<number>(0);
  const toast = useToast();

  // Initialize filters from URL on mount and fetch data
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const detector = params.get('detector');
    const warehouse = params.get('warehouse');
    const page = params.get('page');
    const pageSize = params.get('pageSize');

    if (detector) setSelectedDetector(detector);
    if (warehouse) setSelectedWarehouse(warehouse);
    if (page) {
      const parsedPage = parseInt(page, 10);
      if (!isNaN(parsedPage) && parsedPage >= 1) {
        setCurrentPage(parsedPage);
      }
    }
    if (pageSize) {
      const parsedSize = parseInt(pageSize, 10);
      if ([25, 50, 100, 200].includes(parsedSize)) {
        setPageSize(parsedSize);
      }
    }

    // Fetch config and summary once on mount
    fetchM3Config();
    fetchSummary();

    // Mark as initialized to allow fetching
    setIsInitialized(true);
  }, []);

  // Fetch detector metadata for labels
  useEffect(() => {
    const fetchDetectors = async () => {
      try {
        // Use TRN as default - detector labels are environment-agnostic
        const response = await fetch('/api/detection/detectors?environment=TRN');
        const detectors: Detector[] = await response.json();
        const labels: Record<string, string> = {};
        detectors.forEach(d => labels[d.name] = d.label);
        setDetectorLabels(labels);
      } catch (err) {
        console.error('Failed to load detector labels:', err);
      }
    };
    fetchDetectors();
  }, []);

  // Reset to page 1 when filters change
  useEffect(() => {
    if (isInitialized && currentPage !== 1) {
      setCurrentPage(1);
    }
  }, [selectedDetector, selectedWarehouse, showIgnored]);

  // Fetch issues when filters or pagination changes (only after initialization)
  useEffect(() => {
    if (isInitialized) {
      fetchIssues();
    }
  }, [selectedDetector, selectedWarehouse, showIgnored, currentPage, pageSize, isInitialized]);

  // Sync URL with filter and pagination state
  useEffect(() => {
    const params = new URLSearchParams();
    if (selectedDetector) params.set('detector', selectedDetector);
    if (selectedWarehouse) params.set('warehouse', selectedWarehouse);
    if (currentPage > 1) params.set('page', currentPage.toString());
    if (pageSize !== 50) params.set('pageSize', pageSize.toString());

    const newUrl = params.toString() ? `?${params.toString()}` : window.location.pathname;
    window.history.replaceState(null, '', newUrl);
  }, [selectedDetector, selectedWarehouse, currentPage, pageSize]);

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
      const data = await api.getIssueSummary(showIgnored);
      setSummary(data);
    } catch (error) {
      console.error('Failed to fetch issue summary:', error);
    }
  };

  const fetchIssues = async () => {
    setLoading(true);
    try {
      const result = await api.listInconsistencies({
        type: selectedDetector || undefined,
        warehouse: selectedWarehouse || undefined,
        includeIgnored: showIgnored,
        page: currentPage,
        pageSize: pageSize,
      });
      setIssues(result.data);
      setTotalCount(result.pagination.totalCount);
      setTotalPages(result.pagination.totalPages);
    } catch (error) {
      console.error('Failed to fetch issues:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleIgnore = async (issueId: number) => {
    try {
      await api.ignoreIssue(issueId);
      // Refresh issue list and summary
      await Promise.all([fetchIssues(), fetchSummary()]);
      toast.success('Issue ignored successfully');
    } catch (error) {
      console.error('Failed to ignore issue:', error);
      toast.error('Failed to ignore issue. Please try again.');
    }
  };

  const handleUnignore = async (issueId: number) => {
    try {
      await api.unignoreIssue(issueId);
      // Refresh issue list and summary
      await Promise.all([fetchIssues(), fetchSummary()]);
      toast.success('Issue unignored successfully');
    } catch (error) {
      console.error('Failed to unignore issue:', error);
      toast.error('Failed to unignore issue. Please try again.');
    }
  };

  const handleDeleteMOPClick = (issue: Issue) => {
    setIssueToDelete(issue);
    setDeleteModalOpen(true);
  };

  const handleDeleteMOPConfirm = async () => {
    if (!issueToDelete) return;

    setIsDeleting(true);
    try {
      await api.deletePlannedMO(issueToDelete.id);
      const mopNumber = issueToDelete.productionOrderNumber;
      setDeleteModalOpen(false);
      setIssueToDelete(null);
      // Refresh issue list and summary
      await Promise.all([fetchIssues(), fetchSummary()]);
      toast.success(`MOP ${mopNumber} deleted successfully from M3`);
    } catch (error) {
      console.error('Failed to delete MOP:', error);
      toast.error('Failed to delete MOP from M3. Please try again.');
    } finally {
      setIsDeleting(false);
    }
  };

  const handleDeleteMOPCancel = () => {
    setDeleteModalOpen(false);
    setIssueToDelete(null);
  };

  // Helper functions to determine if MO can be deleted or closed
  const canDeleteMO = (issue: Issue): boolean => {
    if (issue.productionOrderType !== 'MO') return false;
    const status = issue.issueData?.status;
    if (!status) return false;
    const statusNum = parseInt(status, 10);
    return !isNaN(statusNum) && statusNum <= 22;
  };

  const canCloseMO = (issue: Issue): boolean => {
    if (issue.productionOrderType !== 'MO') return false;
    const status = issue.issueData?.status;
    if (!status) return false;
    const statusNum = parseInt(status, 10);
    return !isNaN(statusNum) && statusNum > 22;
  };

  // Delete MO handlers
  const handleDeleteMOClick = (issue: Issue) => {
    setIssueToDelete(issue);
    setDeleteModalOpen(true);
  };

  const handleDeleteMOConfirm = async () => {
    if (!issueToDelete) return;

    setIsDeleting(true);
    try {
      const moNumber = issueToDelete.productionOrderNumber;
      await api.deleteMO(issueToDelete.id);

      setDeleteModalOpen(false);
      setIssueToDelete(null);
      await Promise.all([fetchIssues(), fetchSummary()]);

      toast.success(`MO ${moNumber} deleted successfully from M3`);
    } catch (error) {
      console.error('Failed to delete MO:', error);
      toast.error('Failed to delete MO from M3. Please try again.');
    } finally {
      setIsDeleting(false);
    }
  };

  // Close MO handlers
  const handleCloseMOClick = (issue: Issue) => {
    setIssueToClose(issue);
    setCloseMOModalOpen(true);
  };

  const handleCloseMOConfirm = async () => {
    if (!issueToClose) return;

    setIsClosing(true);
    try {
      const moNumber = issueToClose.productionOrderNumber;
      await api.closeMO(issueToClose.id);

      setCloseMOModalOpen(false);
      setIssueToClose(null);
      await Promise.all([fetchIssues(), fetchSummary()]);

      toast.success(`MO ${moNumber} closed successfully in M3`);
    } catch (error) {
      console.error('Failed to close MO:', error);
      toast.error('Failed to close MO in M3. Please try again.');
    } finally {
      setIsClosing(false);
    }
  };

  const handleCloseMOCancel = () => {
    setCloseMOModalOpen(false);
    setIssueToClose(null);
  };

  // Helper to build alignment confirmation message
  const getAlignmentMessage = (issue: Issue) => {
    const minDate = issue.issueData?.min_date;
    const numOrders = issue.issueData?.num_production_orders || 0;

    if (!minDate) {
      return `This will reschedule ${numOrders} production orders. Continue?`;
    }

    // Check if date is in the past
    const minDateInt = parseInt(String(minDate));
    const today = new Date();
    today.setHours(0, 0, 0, 0);

    const minDateObj = new Date(
      Math.floor(minDateInt / 10000),
      (Math.floor(minDateInt / 100) % 100) - 1,
      minDateInt % 100
    );

    const isPast = minDateObj < today;

    if (isPast) {
      return `The earliest date (${formatM3Date(minDate)}) is in the past. This will reschedule ${numOrders} production orders to the next business day instead. Continue?`;
    }

    return `This will reschedule ${numOrders} production orders to align with the earliest date (${formatM3Date(minDate)}). This action will update orders in M3. Continue?`;
  };

  const handleAlignEarliestClick = (issue: Issue) => {
    setIssueToAlign(issue);
    setAlignModalOpen(true);
  };

  const handleAlignEarliestConfirm = async () => {
    if (!issueToAlign) return;

    setIsAligning(true);
    try {
      const result = await api.alignEarliestMOs(issueToAlign.id);

      setAlignModalOpen(false);
      setIssueToAlign(null);

      // Refresh data
      await Promise.all([fetchIssues(), fetchSummary()]);

      // Build success message with date adjustment info
      let successMessage = `Successfully aligned ${result.aligned_count} production orders to ${formatM3Date(result.target_date)}`;
      if (result.date_adjusted && result.original_min_date) {
        successMessage += ` (adjusted from ${formatM3Date(result.original_min_date)} to next business day)`;
      }

      // Show success/partial success message
      if (result.failed_count === 0) {
        toast.success(successMessage);
      } else if (result.aligned_count > 0) {
        const warningMsg = `Aligned ${result.aligned_count} orders, ${result.failed_count} failed.${result.date_adjusted ? ' Date adjusted to next business day.' : ''}`;
        toast.warning(warningMsg);
        console.error('Alignment failures:', result.failures);
      } else {
        toast.error('Failed to align any production orders. Please try again.');
      }
    } catch (error) {
      console.error('Failed to align orders:', error);
      toast.error('Failed to align production orders. Please try again.');
    } finally {
      setIsAligning(false);
    }
  };

  const handleAlignEarliestCancel = () => {
    setAlignModalOpen(false);
    setIssueToAlign(null);
  };

  return (
    <AppLayout>
      <ToastContainer toasts={toast.toasts} onClose={toast.removeToast} />
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
                  {detectorLabels[detector] || detector}
                </div>
                <div className="mt-2 text-3xl font-semibold text-slate-900">{count}</div>
              </div>
            ))}
          </div>
        )}

        {/* Filters */}
        <div className="mb-6 rounded-xl bg-white p-6 shadow-sm ring-1 ring-slate-200">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
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
                    {detectorLabels[detector] || detector} ({summary.by_detector[detector]})
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

            <div className="flex items-end">
              <label className="flex items-center cursor-pointer">
                <input
                  type="checkbox"
                  checked={showIgnored}
                  onChange={(e) => setShowIgnored(e.target.checked)}
                  className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-slate-300 rounded"
                />
                <span className="ml-2 text-sm text-slate-700">Show Ignored Issues</span>
              </label>
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
                      Actions
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-200 bg-white">
                  {issues.map((issue) => (
                    <tr
                      key={issue.id}
                      className={issue.isIgnored ? 'bg-slate-100 opacity-75 hover:bg-slate-150' : 'hover:bg-slate-50'}
                    >
                      <td className="whitespace-nowrap px-6 py-4 text-sm font-medium text-slate-900">
                        {detectorLabels[issue.detectorType] || issue.detectorType}
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
                            {issue.moTypeDescription && (
                              <div className="text-xs text-slate-600 mt-0.5">
                                {issue.moTypeDescription}
                              </div>
                            )}
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
                        <IssueDetailsCell
                          issueData={issue.issueData}
                          detectorType={issue.detectorType}
                          productionOrderType={issue.productionOrderType}
                          issue={issue}
                          onShowDetail={(issue) => {
                            setSelectedIssueForDetail(issue);
                            setDetailModalOpen(true);
                          }}
                        />
                      </td>
                      <td className="whitespace-nowrap px-6 py-4 text-sm">
                        <div className="flex items-center gap-2">
                          {issue.isIgnored ? (
                            <button
                              onClick={() => handleUnignore(issue.id)}
                              className="inline-flex items-center px-3 py-1.5 border border-primary-300 text-sm font-medium rounded-md text-primary-700 bg-primary-50 hover:bg-primary-100 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
                            >
                              Unignore
                            </button>
                          ) : (
                            <button
                              onClick={() => handleIgnore(issue.id)}
                              className="inline-flex items-center px-3 py-1.5 border border-slate-300 text-sm font-medium rounded-md text-slate-700 bg-white hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-slate-500"
                            >
                              Ignore
                            </button>
                          )}
                          {issue.detectorType === 'unlinked_production_orders' &&
                           issue.productionOrderType === 'MOP' && (
                            <button
                              onClick={() => handleDeleteMOPClick(issue)}
                              className="inline-flex items-center px-3 py-1.5 border border-red-300 text-sm font-medium rounded-md text-red-700 bg-red-50 hover:bg-red-100 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
                            >
                              Delete
                            </button>
                          )}

                          {/* Delete MO button (status <= 22) */}
                          {issue.detectorType === 'unlinked_production_orders' &&
                           canDeleteMO(issue) && (
                            <button
                              onClick={() => handleDeleteMOClick(issue)}
                              className="inline-flex items-center px-3 py-1.5 border border-red-300 text-sm font-medium rounded-md text-red-700 bg-red-50 hover:bg-red-100 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
                            >
                              Delete
                            </button>
                          )}

                          {/* Close MO button (status > 22) */}
                          {issue.detectorType === 'unlinked_production_orders' &&
                           canCloseMO(issue) && (
                            <button
                              onClick={() => handleCloseMOClick(issue)}
                              className="inline-flex items-center px-3 py-1.5 border border-orange-300 text-sm font-medium rounded-md text-orange-700 bg-orange-50 hover:bg-orange-100 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-orange-500"
                            >
                              Close
                            </button>
                          )}

                          {/* Align Earliest button - only for joint_delivery_date_mismatch */}
                          {issue.detectorType === 'joint_delivery_date_mismatch' && (
                            <button
                              onClick={() => handleAlignEarliestClick(issue)}
                              className="inline-flex items-center px-3 py-1.5 border border-blue-300 text-sm font-medium rounded-md text-blue-700 bg-blue-50 hover:bg-blue-100 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                              title="Align all orders to earliest date"
                            >
                              Align Earliest
                            </button>
                          )}
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}

          {/* Pagination Controls */}
          {!loading && issues.length > 0 && (
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
                  Showing {Math.min((currentPage - 1) * pageSize + 1, totalCount)} to {Math.min(currentPage * pageSize, totalCount)} of {totalCount} issues
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

        {/* Delete MOP/MO Confirmation Modal */}
        <ConfirmModal
          isOpen={deleteModalOpen}
          title={issueToDelete?.productionOrderType === 'MOP' ? 'Delete Manufacturing Order Proposal' : 'Delete Manufacturing Order'}
          message={`Are you sure you want to delete ${issueToDelete?.productionOrderType} ${issueToDelete?.productionOrderNumber}? This action cannot be undone and will permanently delete the order from M3.`}
          confirmLabel={isDeleting ? 'Deleting...' : `Delete ${issueToDelete?.productionOrderType || 'Order'}`}
          cancelLabel="Cancel"
          onConfirm={issueToDelete?.productionOrderType === 'MOP' ? handleDeleteMOPConfirm : handleDeleteMOConfirm}
          onCancel={handleDeleteMOPCancel}
          isDestructive={true}
        />

        {/* Close MO Confirmation Modal */}
        <ConfirmModal
          isOpen={closeMOModalOpen}
          title="Close Manufacturing Order"
          message={`Are you sure you want to close MO ${issueToClose?.productionOrderNumber}? This will mark the order as complete in M3. This action cannot be undone.`}
          confirmLabel={isClosing ? 'Closing...' : 'Close MO'}
          cancelLabel="Cancel"
          onConfirm={handleCloseMOConfirm}
          onCancel={handleCloseMOCancel}
          isDestructive={true}
        />

        {/* Align Earliest Confirmation Modal */}
        <ConfirmModal
          isOpen={alignModalOpen}
          title="Align Production Orders to Earliest Date"
          message={issueToAlign ? getAlignmentMessage(issueToAlign) : ''}
          confirmLabel={isAligning ? 'Aligning...' : 'Align Orders'}
          cancelLabel="Cancel"
          onConfirm={handleAlignEarliestConfirm}
          onCancel={handleAlignEarliestCancel}
          isDestructive={false}
        />

        {/* Joint Delivery Detail Modal */}
        {detailModalOpen && selectedIssueForDetail && (
          <JointDeliveryDetailModal
            isOpen={detailModalOpen}
            onClose={() => {
              setDetailModalOpen(false);
              setSelectedIssueForDetail(null);
            }}
            onAlignEarliest={() => {
              setDetailModalOpen(false);  // Close detail modal
              handleAlignEarliestClick(selectedIssueForDetail);  // Open confirm modal
            }}
            issueData={selectedIssueForDetail.issueData}
            coNumber={selectedIssueForDetail.coNumber}
            currentOrderNumber={selectedIssueForDetail.productionOrderNumber}
          />
        )}
      </div>
    </AppLayout>
  );
};

// Helper component to display issue-specific details
const IssueDetailsCell: React.FC<{
  issueData: Record<string, any>;
  detectorType: string;
  productionOrderType?: string;
  issue?: Issue;
  onShowDetail?: (issue: Issue) => void;
}> = ({ issueData, detectorType, productionOrderType, issue, onShowDetail }) => {
  if (detectorType === 'unlinked_production_orders') {
    const startDate = issueData.start_date ? formatM3DateRelative(issueData.start_date) : null;

    return (
      <div className="text-xs">
        {/* Show product_number for MOPs, item_number for MOs */}
        {productionOrderType === 'MOP' && issueData.product_number ? (
          <div>Product: {issueData.product_number}</div>
        ) : (
          <div>Item: {issueData.item_number}</div>
        )}
        {startDate && (
          <div>
            Start:{' '}
            <span
              className="cursor-help border-b border-dotted border-slate-400"
              title={startDate.absolute}
            >
              {startDate.relative}
            </span>
          </div>
        )}
        {issueData.mo_type && (
          <div className="text-slate-400">Type: {issueData.mo_type}</div>
        )}
      </div>
    );
  }

  if (detectorType === 'start_date_mismatch') {
    return (
      <div className="text-xs">
        {issueData.dates && issueData.dates.length > 0 && (
          <div>Dates: {issueData.dates.join(', ')}</div>
        )}
        {issueData.orders && issueData.orders.length > 0 && issueData.orders[0].mo_type && (
          <div className="text-slate-400">Type: {issueData.orders[0].mo_type}</div>
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
        {issueData.mo_type && (
          <div className="text-slate-400">Type: {issueData.mo_type}</div>
        )}
      </div>
    );
  }

  if (detectorType === 'joint_delivery_date_mismatch') {
    return (
      <button
        onClick={() => issue && onShowDetail && onShowDetail(issue)}
        className="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium text-blue-700 bg-blue-50 rounded-md hover:bg-blue-100 transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-1"
      >
        <InformationCircleIcon className="h-4 w-4" />
        Details
      </button>
    );
  }

  return <div className="text-xs">{JSON.stringify(issueData)}</div>;
};

export default Inconsistencies;
