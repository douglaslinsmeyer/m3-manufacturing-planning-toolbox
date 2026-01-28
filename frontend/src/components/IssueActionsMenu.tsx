import React from 'react';
import { ActionMenuButton, ActionMenuItem } from './ActionMenuButton';
import { Trash2, XCircle, Calendar, EyeOff, Eye, InformationCircleIcon } from 'lucide-react';

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

export interface IssueActionsMenuProps {
  issue: Issue;
  onIgnore: (issueId: number) => void;
  onUnignore: (issueId: number) => void;
  onDeleteMOP: (issue: Issue) => void;
  onDeleteMO: (issue: Issue) => void;
  onCloseMO: (issue: Issue) => void;
  onAlignEarliest: (issue: Issue) => void;
  onAlignLatest: (issue: Issue) => void;
  onShowDetails: (issue: Issue) => void;
}

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

// Icon component for Information Circle (for Details action)
function InfoIcon({ className }: { className?: string }) {
  return (
    <svg className={className} fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" d="M11.25 11.25l.041-.02a.75.75 0 011.063.852l-.708 2.836a.75.75 0 001.063.853l.041-.021M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9-3.75h.008v.008H12V8.25z" />
    </svg>
  );
}

// Icon component for Wrench (for Fix button)
function WrenchIcon({ className }: { className?: string }) {
  return (
    <svg className={className} fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" d="M21.75 6.75a4.5 4.5 0 01-4.884 4.484c-1.076-.091-2.264.071-2.95.904l-7.152 8.684a2.548 2.548 0 11-3.586-3.586l8.684-7.152c.833-.686.995-1.874.904-2.95a4.5 4.5 0 016.336-4.486l-3.276 3.276a3.004 3.004 0 002.25 2.25l3.276-3.276c.256.565.398 1.192.398 1.852z" />
    </svg>
  );
}

export const IssueActionsMenu: React.FC<IssueActionsMenuProps> = ({
  issue,
  onIgnore,
  onUnignore,
  onDeleteMOP,
  onDeleteMO,
  onCloseMO,
  onAlignEarliest,
  onAlignLatest,
  onShowDetails,
}) => {
  // Build the list of available actions based on issue type and status
  const getAvailableActions = (): ActionMenuItem[] => {
    const actions: ActionMenuItem[] = [];

    // Ignore/Unignore - ALWAYS available (first item)
    if (issue.isIgnored) {
      actions.push({
        id: 'unignore',
        label: 'Unignore Issue',
        icon: <Eye className="h-4 w-4" />,
        variant: 'default',
        disabled: false,
      });
    } else {
      actions.push({
        id: 'ignore',
        label: 'Ignore Issue',
        icon: <EyeOff className="h-4 w-4" />,
        variant: 'default',
        disabled: false,
      });
    }

    // Delete - For unlinked_production_orders
    if (issue.detectorType === 'unlinked_production_orders') {
      if (issue.productionOrderType === 'MOP') {
        actions.push({
          id: 'delete-mop',
          label: 'Delete MOP',
          icon: <Trash2 className="h-4 w-4" />,
          variant: 'danger',
          disabled: false,
        });
      } else if (canDeleteMO(issue)) {
        actions.push({
          id: 'delete-mo',
          label: 'Delete MO',
          icon: <Trash2 className="h-4 w-4" />,
          variant: 'danger',
          disabled: false,
        });
      }

      // Close action - for closeable MOs
      if (canCloseMO(issue)) {
        actions.push({
          id: 'close-mo',
          label: 'Close MO',
          icon: <XCircle className="h-4 w-4" />,
          variant: 'warning',
          disabled: false,
        });
      }
    }

    // Alignment and Details actions - for date mismatch detectors
    if (
      issue.detectorType === 'joint_delivery_date_mismatch' ||
      issue.detectorType === 'dlix_date_mismatch'
    ) {
      // Details action first for date mismatch issues
      actions.push({
        id: 'show-details',
        label: 'View Details',
        icon: <InfoIcon className="h-4 w-4" />,
        variant: 'default',
        disabled: false,
      });

      actions.push({
        id: 'align-earliest',
        label: 'Align to Earliest Date',
        icon: <Calendar className="h-4 w-4" />,
        variant: 'info',
        disabled: false,
      });

      actions.push({
        id: 'align-latest',
        label: 'Align to Latest Date',
        icon: <Calendar className="h-4 w-4" />,
        variant: 'success',
        disabled: false,
      });
    }

    return actions;
  };

  const handleActionSelect = (actionId: string) => {
    switch (actionId) {
      case 'ignore':
        onIgnore(issue.id);
        break;
      case 'unignore':
        onUnignore(issue.id);
        break;
      case 'delete-mop':
        onDeleteMOP(issue);
        break;
      case 'delete-mo':
        onDeleteMO(issue);
        break;
      case 'close-mo':
        onCloseMO(issue);
        break;
      case 'align-earliest':
        onAlignEarliest(issue);
        break;
      case 'align-latest':
        onAlignLatest(issue);
        break;
      case 'show-details':
        onShowDetails(issue);
        break;
    }
  };

  const actions = getAvailableActions();

  return (
    <ActionMenuButton
      label="Fix"
      icon={<WrenchIcon className="h-4 w-4" />}
      actions={actions}
      onActionSelect={handleActionSelect}
    />
  );
};
