import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { AppLayout } from '../components/AppLayout';
import { buildM3BookmarkURL, M3Config } from '../utils/m3Links';
import { dateDiffDays, getVarianceBadgeColor } from '../utils/m3DateUtils';
import { api } from '../services/api';
import { Issue, AnomalySummary } from '../types';
import { ConfirmModal } from '../components/ConfirmModal';
import { JointDeliveryDetailModal } from '../components/JointDeliveryDetailModal';
import { ToastContainer } from '../components/Toast';
import { useToast } from '../hooks/useToast';
import { BulkActionToolbar, BulkAction } from '../components/BulkActionToolbar';
import { BulkOperationModal, BulkOperationResult } from '../components/BulkOperationModal';
import { IssueActionsMenu } from '../components/IssueActionsMenu';
import { Trash2, XCircle, Calendar } from 'lucide-react';

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

const Issues: React.FC = () => {
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
  const [alignLatestModalOpen, setAlignLatestModalOpen] = useState(false);
  const [issueToAlignLatest, setIssueToAlignLatest] = useState<Issue | null>(null);
  const [isAligningLatest, setIsAligningLatest] = useState(false);
  const [detailModalOpen, setDetailModalOpen] = useState(false);
  const [selectedIssueForDetail, setSelectedIssueForDetail] = useState<Issue | null>(null);
  const [isInitialized, setIsInitialized] = useState(false);
  const [currentPage, setCurrentPage] = useState<number>(1);
  const [pageSize, setPageSize] = useState<number>(50);
  const [totalCount, setTotalCount] = useState<number>(0);
  const [totalPages, setTotalPages] = useState<number>(0);
  const toast = useToast();

  // Multi-select state
  const [selectedIssues, setSelectedIssues] = useState<Set<number>>(new Set());

  // Bulk operation modal state
  const [bulkModalOpen, setBulkModalOpen] = useState(false);
  const [bulkModalTitle, setBulkModalTitle] = useState('');
  const [bulkResults, setBulkResults] = useState<BulkOperationResult[]>([]);
  const [isBulkProcessing, setIsBulkProcessing] = useState(false);
  const [bulkTotal, setBulkTotal] = useState(0);
  const [bulkCompleted, setBulkCompleted] = useState(0);

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
      // Clear selection when data changes
      clearSelection();
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
      const result = await api.listIssues({
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
      return `This will reschedule ${numOrders.toLocaleString()} production orders. Continue?`;
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
      return `The earliest date (${formatM3Date(minDate)}) is in the past. This will reschedule ${numOrders.toLocaleString()} production orders to the next business day instead. Continue?`;
    }

    return `This will reschedule ${numOrders.toLocaleString()} production orders to align with the earliest date (${formatM3Date(minDate)}). This action will update orders in M3. Continue?`;
  };

  // Helper to build latest alignment confirmation message
  const getLatestAlignmentMessage = (issue: Issue) => {
    const maxDate = issue.issueData?.max_date;
    const numOrders = issue.issueData?.num_production_orders || 0;

    if (!maxDate) {
      return `This will reschedule ${numOrders.toLocaleString()} production orders. Continue?`;
    }

    // Check if date is in the past
    const maxDateInt = parseInt(String(maxDate));
    const today = new Date();
    today.setHours(0, 0, 0, 0);

    const maxDateObj = new Date(
      Math.floor(maxDateInt / 10000),
      (Math.floor(maxDateInt / 100) % 100) - 1,
      maxDateInt % 100
    );

    const isPast = maxDateObj < today;

    if (isPast) {
      return `The latest date (${formatM3Date(maxDate)}) is in the past. This will reschedule ${numOrders.toLocaleString()} production orders to the next business day instead. Continue?`;
    }

    return `This will reschedule ${numOrders.toLocaleString()} production orders to align with the latest date (${formatM3Date(maxDate)}). This action will update orders in M3. Continue?`;
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

  const handleAlignLatestClick = (issue: Issue) => {
    setIssueToAlignLatest(issue);
    setAlignLatestModalOpen(true);
  };

  const handleAlignLatestConfirm = async () => {
    if (!issueToAlignLatest) return;

    setIsAligningLatest(true);
    try {
      const result = await api.alignLatestMOs(issueToAlignLatest.id);

      setAlignLatestModalOpen(false);
      setIssueToAlignLatest(null);

      // Refresh data
      await Promise.all([fetchIssues(), fetchSummary()]);

      // Build success message with date adjustment info
      let successMessage = `Successfully aligned ${result.aligned_count} production orders to ${formatM3Date(result.target_date)}`;
      if (result.date_adjusted && result.original_max_date) {
        successMessage += ` (adjusted from ${formatM3Date(result.original_max_date)} to next business day)`;
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
      setIsAligningLatest(false);
    }
  };

  const handleAlignLatestCancel = () => {
    setAlignLatestModalOpen(false);
    setIssueToAlignLatest(null);
  };

  // Multi-select handlers
  const handleSelectAll = (checked: boolean) => {
    if (checked) {
      setSelectedIssues(new Set(issues.map(issue => issue.id)));
    } else {
      setSelectedIssues(new Set());
    }
  };

  const handleSelectIssue = (issueId: number, checked: boolean) => {
    const newSelection = new Set(selectedIssues);
    if (checked) {
      newSelection.add(issueId);
    } else {
      newSelection.delete(issueId);
    }
    setSelectedIssues(newSelection);
  };

  const clearSelection = () => {
    setSelectedIssues(new Set());
  };

  // Calculate duplicate info from selected issues
  const getDuplicateInfo = () => {
    if (selectedIssues.size === 0) {
      return { uniqueOrders: 0, duplicates: 0 };
    }

    const orderNumbers = new Set<string>();
    issues.forEach((issue) => {
      if (selectedIssues.has(issue.id)) {
        orderNumbers.add(issue.production_order_number);
      }
    });

    const uniqueOrders = orderNumbers.size;
    const duplicates = selectedIssues.size - uniqueOrders;

    return { uniqueOrders, duplicates };
  };

  const duplicateInfo = getDuplicateInfo();

  // Get available bulk actions based on selected issues
  const getAvailableActions = (): BulkAction[] => {
    const selected = Array.from(selectedIssues)
      .map(id => issues.find(i => i.id === id))
      .filter((issue): issue is Issue => issue !== undefined);

    if (selected.length === 0) {
      return [];
    }

    const allMOPs = selected.every(i => i.productionOrderType === 'MOP');
    const allMOs = selected.every(i => i.productionOrderType === 'MO');
    const allDeletableMOs = allMOs && selected.every(i => {
      const status = parseInt(i.issueData.status || '99', 10);
      return !isNaN(status) && status <= 22;
    });
    const allCloseableMOs = allMOs && selected.every(i => {
      const status = parseInt(i.issueData.status || '99', 10);
      return !isNaN(status) && status > 22;
    });

    const actions: BulkAction[] = [];

    // Delete action
    actions.push({
      id: 'delete',
      label: 'Delete',
      icon: <Trash2 className="h-4 w-4" />,
      variant: 'danger',
      enabled: allMOPs || allDeletableMOs,
      disabledReason: !allMOPs && !allDeletableMOs
        ? 'Select only MOPs or deletable MOs (status ≤22)'
        : undefined,
    });

    // Close action
    actions.push({
      id: 'close',
      label: 'Close',
      icon: <XCircle className="h-4 w-4" />,
      variant: 'warning',
      enabled: allCloseableMOs,
      disabledReason: !allCloseableMOs
        ? 'Select only MOs with status >22'
        : undefined,
    });

    // Reschedule action
    actions.push({
      id: 'reschedule',
      label: 'Reschedule',
      icon: <Calendar className="h-4 w-4" />,
      variant: 'primary',
      enabled: true,
      disabledReason: undefined,
    });

    return actions;
  };

  // Bulk operation handlers
  const handleBulkDelete = async () => {
    const issueIds = Array.from(selectedIssues);

    // Initialize modal with pending state
    const pendingResults: BulkOperationResult[] = issueIds.map((id) => ({
      issue_id: id,
      production_order: issues.find((i) => i.id === id)?.production_order_number || 'Unknown',
      status: 'pending' as const,
    }));

    setBulkModalTitle('Bulk Delete Production Orders');
    setBulkTotal(issueIds.length);
    setBulkCompleted(0);
    setBulkResults(pendingResults);
    setBulkModalOpen(true);
    setIsBulkProcessing(true);

    try {
      // Call bulk delete API (returns job_id)
      const response = await api.bulkDelete(issueIds);
      const jobId = response.job_id;

      // Poll for issue results
      const pollInterval = setInterval(async () => {
        try {
          const issueResults = await api.getBulkOperationIssueResults(jobId);

          // Convert to modal format
          const modalResults: BulkOperationResult[] = issueResults.results.map(r => ({
            issue_id: r.issue_id,
            production_order: r.production_order,
            status: r.status,
            message: r.message,
            error: r.error,
            is_duplicate: r.is_duplicate,
            primary_issue_id: r.primary_issue_id,
          }));

          setBulkResults(modalResults);
          setBulkCompleted(modalResults.filter(r => r.status !== 'pending').length);

          // Stop polling when all complete
          const allComplete = modalResults.every(r => r.status !== 'pending');
          if (allComplete) {
            clearInterval(pollInterval);
            setIsBulkProcessing(false);

            // Refresh data
            await Promise.all([fetchIssues(), fetchSummary()]);
            clearSelection();

            // Show toast
            const successCount = modalResults.filter(r => r.status === 'success').length;
            const failCount = modalResults.filter(r => r.status === 'error').length;

            if (failCount === 0) {
              toast.success(`Successfully deleted ${successCount} production orders`);
            } else if (successCount > 0) {
              toast.warning(`Deleted ${successCount} orders, ${failCount} failed`);
            } else {
              toast.error('Failed to delete any production orders');
            }
          }
        } catch (error) {
          clearInterval(pollInterval);
          console.error('Failed to fetch issue results:', error);
          toast.error('Failed to fetch operation results');
          setIsBulkProcessing(false);
        }
      }, 1000); // Poll every 1 second
    } catch (error: any) {
      console.error('Bulk delete failed:', error);
      toast.error(error.message || 'Bulk delete failed');
      setIsBulkProcessing(false);
    }
  };

  const handleBulkClose = async () => {
    const issueIds = Array.from(selectedIssues);

    // Initialize modal with pending state
    const pendingResults: BulkOperationResult[] = issueIds.map((id) => ({
      issue_id: id,
      production_order: issues.find((i) => i.id === id)?.production_order_number || 'Unknown',
      status: 'pending' as const,
    }));

    setBulkModalTitle('Bulk Close Manufacturing Orders');
    setBulkTotal(issueIds.length);
    setBulkCompleted(0);
    setBulkResults(pendingResults);
    setBulkModalOpen(true);
    setIsBulkProcessing(true);

    try {
      // Call bulk close API (returns job_id)
      const response = await api.bulkClose(issueIds);
      const jobId = response.job_id;

      // Poll for issue results
      const pollInterval = setInterval(async () => {
        try {
          const issueResults = await api.getBulkOperationIssueResults(jobId);

          // Convert to modal format
          const modalResults: BulkOperationResult[] = issueResults.results.map(r => ({
            issue_id: r.issue_id,
            production_order: r.production_order,
            status: r.status,
            message: r.message,
            error: r.error,
            is_duplicate: r.is_duplicate,
            primary_issue_id: r.primary_issue_id,
          }));

          setBulkResults(modalResults);
          setBulkCompleted(modalResults.filter(r => r.status !== 'pending').length);

          // Stop polling when all complete
          const allComplete = modalResults.every(r => r.status !== 'pending');
          if (allComplete) {
            clearInterval(pollInterval);
            setIsBulkProcessing(false);

            // Refresh data
            await Promise.all([fetchIssues(), fetchSummary()]);
            clearSelection();

            // Show toast
            const successCount = modalResults.filter(r => r.status === 'success').length;
            const failCount = modalResults.filter(r => r.status === 'error').length;

            if (failCount === 0) {
              toast.success(`Successfully closed ${successCount} manufacturing orders`);
            } else if (successCount > 0) {
              toast.warning(`Closed ${successCount} orders, ${failCount} failed`);
            } else {
              toast.error('Failed to close any manufacturing orders');
            }
          }
        } catch (error) {
          clearInterval(pollInterval);
          console.error('Failed to fetch issue results:', error);
          toast.error('Failed to fetch operation results');
          setIsBulkProcessing(false);
        }
      }, 1000); // Poll every 1 second
    } catch (error: any) {
      console.error('Bulk close failed:', error);
      toast.error(error.message || 'Bulk close failed');
      setIsBulkProcessing(false);
    }
  };

  const handleBulkReschedule = async () => {
    // Prompt for new date
    const dateInput = prompt('Enter new date (YYYYMMDD format):');
    if (!dateInput) {
      return; // User cancelled
    }

    // Validate date format
    const dateRegex = /^\d{8}$/;
    if (!dateRegex.test(dateInput)) {
      toast.error('Invalid date format. Please use YYYYMMDD (e.g., 20260201)');
      return;
    }

    const issueIds = Array.from(selectedIssues);

    // Initialize modal with pending state
    const pendingResults: BulkOperationResult[] = issueIds.map((id) => ({
      issue_id: id,
      production_order: issues.find((i) => i.id === id)?.production_order_number || 'Unknown',
      status: 'pending' as const,
    }));

    setBulkModalTitle('Bulk Reschedule Production Orders');
    setBulkTotal(issueIds.length);
    setBulkCompleted(0);
    setBulkResults(pendingResults);
    setBulkModalOpen(true);
    setIsBulkProcessing(true);

    try {
      // Call bulk reschedule API (returns job_id)
      const response = await api.bulkReschedule(issueIds, dateInput);
      const jobId = response.job_id;

      // Poll for issue results
      const pollInterval = setInterval(async () => {
        try {
          const issueResults = await api.getBulkOperationIssueResults(jobId);

          // Convert to modal format
          const modalResults: BulkOperationResult[] = issueResults.results.map(r => ({
            issue_id: r.issue_id,
            production_order: r.production_order,
            status: r.status,
            message: r.message,
            error: r.error,
            is_duplicate: r.is_duplicate,
            primary_issue_id: r.primary_issue_id,
          }));

          setBulkResults(modalResults);
          setBulkCompleted(modalResults.filter(r => r.status !== 'pending').length);

          // Stop polling when all complete
          const allComplete = modalResults.every(r => r.status !== 'pending');
          if (allComplete) {
            clearInterval(pollInterval);
            setIsBulkProcessing(false);

            // Refresh data
            await Promise.all([fetchIssues(), fetchSummary()]);
            clearSelection();

            // Show toast
            const successCount = modalResults.filter(r => r.status === 'success').length;
            const failCount = modalResults.filter(r => r.status === 'error').length;

            if (failCount === 0) {
              toast.success(`Successfully rescheduled ${successCount} production orders to ${formatM3Date(dateInput)}`);
            } else if (successCount > 0) {
              toast.warning(`Rescheduled ${successCount} orders, ${failCount} failed`);
            } else {
              toast.error('Failed to reschedule any production orders');
            }
          }
        } catch (error) {
          clearInterval(pollInterval);
          console.error('Failed to fetch issue results:', error);
          toast.error('Failed to fetch operation results');
          setIsBulkProcessing(false);
        }
      }, 1000); // Poll every 1 second
    } catch (error: any) {
      console.error('Bulk reschedule failed:', error);
      toast.error(error.message || 'Bulk reschedule failed');
      setIsBulkProcessing(false);
    }
  };

  const handleBulkAction = (actionId: string) => {
    switch (actionId) {
      case 'delete':
        handleBulkDelete();
        break;
      case 'close':
        handleBulkClose();
        break;
      case 'reschedule':
        handleBulkReschedule();
        break;
      default:
        console.warn('Unknown bulk action:', actionId);
    }
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
              <div className="mt-2 text-3xl font-semibold text-slate-900">{summary.total.toLocaleString()}</div>
            </div>

            {Object.entries(summary.by_detector).slice(0, 3).map(([detector, count]) => (
              <div key={detector} className="rounded-xl bg-white p-6 shadow-sm ring-1 ring-slate-200">
                <div className="text-sm font-medium text-slate-500">
                  {detectorLabels[detector] || detector}
                </div>
                <div className="mt-2 text-3xl font-semibold text-slate-900">{count.toLocaleString()}</div>
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
                    {detectorLabels[detector] || detector} ({summary.by_detector[detector].toLocaleString()})
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
                    {warehouse} ({summary.by_warehouse[warehouse].toLocaleString()})
                  </option>
                ))}
              </select>
            </div>

            <div className="flex items-end">
              <label className="relative inline-flex items-center cursor-pointer">
                <input
                  type="checkbox"
                  checked={showIgnored}
                  onChange={(e) => setShowIgnored(e.target.checked)}
                  className="sr-only peer"
                />
                <div className="w-11 h-6 bg-slate-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-slate-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary-600"></div>
                <span className="ml-3 text-sm font-medium text-slate-700">Show Ignored Issues</span>
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
                    <th className="px-6 py-3 text-left">
                      <input
                        type="checkbox"
                        checked={issues.length > 0 && selectedIssues.size === issues.length}
                        onChange={(e) => handleSelectAll(e.target.checked)}
                        className="h-4 w-4 rounded border-slate-300 text-primary-600 focus:ring-primary-500"
                      />
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-slate-500">
                      Issue Type
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-slate-500">
                      Affected Orders
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-slate-500">
                      Warehouse
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
                      <td className="px-6 py-4">
                        <input
                          type="checkbox"
                          checked={selectedIssues.has(issue.id)}
                          onChange={(e) => handleSelectIssue(issue.id, e.target.checked)}
                          className="h-4 w-4 rounded border-slate-300 text-primary-600 focus:ring-primary-500"
                        />
                      </td>
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
                            {issue.moTypeDescription && issue.detectorType !== 'unlinked_production_orders' && (
                              <div className="text-xs text-slate-600 mt-0.5">
                                {issue.moTypeDescription}
                              </div>
                            )}
                          </div>
                        )}
                        {issue.coNumber &&
                          issue.detectorType !== 'joint_delivery_date_mismatch' &&
                          issue.detectorType !== 'dlix_date_mismatch' && (
                          <div className="text-xs text-slate-400">
                            CO: {issue.coNumber}-{issue.coLine}
                          </div>
                        )}
                      </td>
                      <td className="whitespace-nowrap px-6 py-4 text-sm text-slate-500">
                        {issue.warehouse || '-'}
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
                        <IssueActionsMenu
                          issue={issue}
                          onIgnore={handleIgnore}
                          onUnignore={handleUnignore}
                          onDeleteMOP={handleDeleteMOPClick}
                          onDeleteMO={handleDeleteMOClick}
                          onCloseMO={handleCloseMOClick}
                          onAlignEarliest={handleAlignEarliestClick}
                          onAlignLatest={handleAlignLatestClick}
                          onShowDetails={(issue) => {
                            setSelectedIssueForDetail(issue);
                            setDetailModalOpen(true);
                          }}
                        />
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
                  Showing {Math.min((currentPage - 1) * pageSize + 1, totalCount).toLocaleString()} to {Math.min(currentPage * pageSize, totalCount).toLocaleString()} of {totalCount.toLocaleString()} issues
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

        {/* Align Latest Confirmation Modal */}
        <ConfirmModal
          isOpen={alignLatestModalOpen}
          title="Align Production Orders to Latest Date"
          message={issueToAlignLatest ? getLatestAlignmentMessage(issueToAlignLatest) : ''}
          confirmLabel={isAligningLatest ? 'Aligning...' : 'Align Orders'}
          cancelLabel="Cancel"
          onConfirm={handleAlignLatestConfirm}
          onCancel={handleAlignLatestCancel}
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
            onAlignLatest={() => {
              setDetailModalOpen(false);  // Close detail modal
              handleAlignLatestClick(selectedIssueForDetail);  // Open confirm modal
            }}
            issueData={selectedIssueForDetail.issueData}
            coNumber={selectedIssueForDetail.coNumber}
            currentOrderNumber={selectedIssueForDetail.productionOrderNumber}
            issueType={selectedIssueForDetail.detectorType as 'joint_delivery_date_mismatch' | 'dlix_date_mismatch'}
          />
        )}

        {/* Bulk Action Toolbar */}
        <BulkActionToolbar
          selectedCount={selectedIssues.size}
          uniqueOrderCount={duplicateInfo.uniqueOrders}
          duplicateCount={duplicateInfo.duplicates}
          availableActions={getAvailableActions()}
          onExecute={handleBulkAction}
          onClear={clearSelection}
        />

        {/* Bulk Operation Modal */}
        <BulkOperationModal
          isOpen={bulkModalOpen}
          title={bulkModalTitle}
          total={bulkTotal}
          completed={bulkCompleted}
          results={bulkResults}
          onClose={() => {
            setBulkModalOpen(false);
            setBulkResults([]);
          }}
          isProcessing={isBulkProcessing}
        />
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
        {issue?.moTypeDescription && (
          <div className="text-slate-400">Type: {issue.moTypeDescription}</div>
        )}
      </div>
    );
  }

  if (detectorType === 'joint_delivery_date_mismatch' || detectorType === 'dlix_date_mismatch') {
    // Validate required data exists
    if (!issueData.min_date || !issueData.max_date) {
      return <div className="text-xs text-slate-400">N/A</div>;
    }

    try {
      const varianceDays = dateDiffDays(issueData.min_date, issueData.max_date);

      return (
        <button
          onClick={() => issue && onShowDetail && onShowDetail(issue)}
          className="inline-flex items-center gap-1.5 hover:opacity-80 transition-opacity focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-1 rounded-md"
          title="Click to view details"
        >
          <span className="text-xs text-slate-600">Variance:</span>
          <span
            className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-semibold ring-1 ${getVarianceBadgeColor(varianceDays)}`}
          >
            {varianceDays} {varianceDays === 1 ? 'day' : 'days'}
          </span>
        </button>
      );
    } catch (error) {
      // Fallback if date parsing fails
      return <div className="text-xs text-slate-400">Invalid dates</div>;
    }
  }

  return <div className="text-xs">{JSON.stringify(issueData)}</div>;
};

export default Issues;
