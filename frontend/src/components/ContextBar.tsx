import React, { useEffect } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { useContextManagement } from '../hooks/useContextManagement';
import './ContextBar.css';

interface ContextBarProps {
  onOpenSwitcher: () => void;
}

export const ContextBar: React.FC<ContextBarProps> = ({ onOpenSwitcher }) => {
  const { isAuthenticated, loading: authLoading } = useAuth();
  const { effectiveContext, loadEffectiveContext } = useContextManagement();

  useEffect(() => {
    // Only load context if auth check completed AND user is authenticated
    if (authLoading || !isAuthenticated) {
      return;
    }
    loadEffectiveContext();
  }, [isAuthenticated, authLoading, loadEffectiveContext]);

  if (!effectiveContext) {
    return null;
  }

  const contextDisplay = `${effectiveContext.company} • ${effectiveContext.division} • ${effectiveContext.facility} • ${effectiveContext.warehouse}`;

  return (
    <div
      className={`context-bar ${effectiveContext.hasTemporaryOverrides ? 'has-overrides' : ''}`}
      onClick={onOpenSwitcher}
      title="Click to change context"
    >
      <span className="context-label">Context:</span>
      <span className="context-value">{contextDisplay}</span>
      {effectiveContext.hasTemporaryOverrides && (
        <span className="override-badge">Temporary</span>
      )}
      <span className="context-icon">⚙</span>
    </div>
  );
};
