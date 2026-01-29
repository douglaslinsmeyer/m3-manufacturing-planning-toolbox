import React, { useState } from 'react';
import { AppLayout } from '../components/AppLayout';
import { FilterBuilder } from '../components/bulk-operations/FilterBuilder';
import { PreviewPanel } from '../components/bulk-operations/PreviewPanel';
import { OperationPanel } from '../components/bulk-operations/OperationPanel';
import { BulkOperationModal } from '../components/BulkOperationModal';
import { ConfirmModal } from '../components/ConfirmModal';
import { api } from '../services/api';
import { ToastContainer } from '../components/Toast';
import { useToast } from '../hooks/useToast';

export interface IssueCriteria {
  detector_type: string;
  facility: string;
  warehouse: string;
  product_number: string;
  include_ignored: boolean;
}

export interface PreviewSummary {
  by_order_type: Record<string, number>;
  by_facility: Record<string, number>;
  by_detector: Record<string, number>;
}

export interface SampleIssue {
  id: number;
  production_order_number: string;
  production_order_type: string;
  detector_type: string;
  facility: string;
  warehouse?: string;
}

export interface PreviewData {
  total_count: number;
  summary: PreviewSummary;
  sample_issues: SampleIssue[];
}

export const BulkOperations: React.FC = () => {
  const toast = useToast();

  // Filter state
  const [filters, setFilters] = useState<IssueCriteria>({
    detector_type: '',
    facility: '',
    warehouse: '',
    product_number: '',
    include_ignored: false,
  });

  // Preview state
  const [previewData, setPreviewData] = useState<PreviewData | null>(null);
  const [isLoadingPreview, setIsLoadingPreview] = useState(false);

  // Operation state
  const [selectedOperation, setSelectedOperation] = useState<string>('');
  const [operationParams, setOperationParams] = useState<Record<string, any>>({});

  // Job tracking state
  const [activeJobId, setActiveJobId] = useState<string | null>(null);

  // Confirmation modal state
  const [showConfirm, setShowConfirm] = useState(false);
  const [confirmConfig, setConfirmConfig] = useState<{
    title: string;
    message: string;
    onConfirm: () => void;
  } | null>(null);

  // Handler: Preview matching issues
  const handlePreview = async () => {
    setIsLoadingPreview(true);
    try {
      const response = await api.post('/issues/bulk-preview', {
        criteria: filters,
      });
      setPreviewData(response.data);
      toast.success('Preview loaded successfully');
    } catch (error: any) {
      console.error('Preview failed:', error);
      toast.error(error.response?.data?.message || 'Failed to load preview');
      setPreviewData(null);
    } finally {
      setIsLoadingPreview(false);
    }
  };

  // Handler: Execute bulk operation
  const handleExecute = async () => {
    if (!previewData || !selectedOperation) {
      toast.warning('Please preview and select an operation first');
      return;
    }

    // Build confirmation message
    const operationLabels: Record<string, string> = {
      delete: 'delete',
      close: 'close',
      reschedule: 'reschedule',
    };

    const operationLabel = operationLabels[selectedOperation] || selectedOperation;
    const message = `This will ${operationLabel} ${previewData.total_count.toLocaleString()} production orders.`;

    // Show confirmation modal
    setConfirmConfig({
      title: `Execute Bulk ${operationLabel.charAt(0).toUpperCase() + operationLabel.slice(1)}`,
      message,
      onConfirm: async () => {
        setShowConfirm(false);
        await executeOperation();
      },
    });
    setShowConfirm(true);
  };

  // Execute the operation after confirmation
  const executeOperation = async () => {
    try {
      const payload: any = {
        criteria: filters,
      };

      // Add operation-specific params
      if (selectedOperation === 'reschedule' && operationParams.new_date) {
        payload.params = {
          new_date: operationParams.new_date,
        };
      }

      const response = await api.post(`/issues/bulk-${selectedOperation}`, payload);

      setActiveJobId(response.data.job_id);
      toast.success(`Bulk ${selectedOperation} operation started`);

      // Clear preview after successful submission
      setPreviewData(null);
      setSelectedOperation('');
      setOperationParams({});
    } catch (error: any) {
      console.error('Execution failed:', error);
      toast.error(error.response?.data?.message || 'Failed to execute operation');
    }
  };

  return (
    <AppLayout>
      <div className="container mx-auto px-4 py-6">
        <div className="mb-6">
          <h1 className="text-2xl font-bold text-gray-900">Bulk Operations Tool</h1>
          <p className="mt-1 text-sm text-gray-600">
            Apply criteria-based filters to preview and execute bulk operations on detected issues.
          </p>
        </div>

        {/* Three-column layout */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Left Column: Filter Builder */}
          <div className="lg:col-span-1">
            <FilterBuilder
              filters={filters}
              onChange={setFilters}
              onPreview={handlePreview}
              isLoading={isLoadingPreview}
            />
          </div>

          {/* Middle Column: Preview & Summary */}
          <div className="lg:col-span-1">
            <PreviewPanel
              previewData={previewData}
              isLoading={isLoadingPreview}
            />
          </div>

          {/* Right Column: Operation Selection & Execution */}
          <div className="lg:col-span-1">
            <OperationPanel
              previewData={previewData}
              selectedOperation={selectedOperation}
              onOperationChange={setSelectedOperation}
              operationParams={operationParams}
              onParamsChange={setOperationParams}
              onExecute={handleExecute}
            />
          </div>
        </div>

        {/* Progress Modal (reuse existing) */}
        {activeJobId && (
          <BulkOperationModal
            jobId={activeJobId}
            onClose={() => {
              setActiveJobId(null);
              // Optionally refresh preview after job completes
            }}
          />
        )}

        {/* Confirmation Modal */}
        {showConfirm && confirmConfig && (
          <ConfirmModal
            isOpen={showConfirm}
            title={confirmConfig.title}
            message={confirmConfig.message}
            confirmLabel="Execute"
            cancelLabel="Cancel"
            variant="danger"
            onConfirm={confirmConfig.onConfirm}
            onCancel={() => setShowConfirm(false)}
          />
        )}

        <ToastContainer toasts={toast.toasts} onClose={toast.removeToast} />
      </div>
    </AppLayout>
  );
};
