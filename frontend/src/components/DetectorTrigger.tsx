import React, { useState, useEffect } from 'react';

function PlayIcon({ className }: { className?: string }) {
  return (
    <svg className={className} fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" d="M5.25 5.653c0-.856.917-1.398 1.667-.986l11.54 6.348a1.125 1.125 0 010 1.971l-11.54 6.347a1.125 1.125 0 01-1.667-.985V5.653z" />
    </svg>
  );
}

interface Detector {
  name: string;
  label: string;
  description: string;
  enabled: boolean;
}

interface DetectorTriggerProps {
  environment: string;
  disabled?: boolean;
  onTrigger?: (jobId: string) => void;
}

export const DetectorTrigger: React.FC<DetectorTriggerProps> = ({
  environment,
  disabled = false,
  onTrigger,
}) => {
  const [detectors, setDetectors] = useState<Detector[]>([]);
  const [selectedDetectors, setSelectedDetectors] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);
  const [triggering, setTriggering] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showPanel, setShowPanel] = useState(false);

  // Load available detectors
  useEffect(() => {
    if (showPanel) {
      loadDetectors();
    }
  }, [showPanel, environment]);

  const loadDetectors = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await fetch(`/api/detection/detectors?environment=${environment}`);
      if (!response.ok) {
        throw new Error('Failed to load detectors');
      }
      const data = await response.json();
      setDetectors(data);

      // Auto-select all enabled detectors
      setSelectedDetectors(data.filter((d: Detector) => d.enabled).map((d: Detector) => d.name));
    } catch (err) {
      console.error('Failed to load detectors:', err);
      setError(err instanceof Error ? err.message : 'Failed to load detectors');
    } finally {
      setLoading(false);
    }
  };

  const handleToggleDetector = (name: string) => {
    setSelectedDetectors(prev =>
      prev.includes(name)
        ? prev.filter(n => n !== name)
        : [...prev, name]
    );
  };

  const handleTrigger = async () => {
    if (selectedDetectors.length === 0) {
      setError('Please select at least one detector');
      return;
    }

    setTriggering(true);
    setError(null);

    try {
      const response = await fetch('/api/detection/trigger', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          environment,
          detectorNames: selectedDetectors,
        }),
      });

      if (!response.ok) {
        const errorData = await response.text();
        throw new Error(errorData || 'Failed to trigger detectors');
      }

      const data = await response.json();
      console.log('Detectors triggered:', data);

      // Close panel and notify parent
      setShowPanel(false);
      if (onTrigger) {
        onTrigger(data.jobId);
      }
    } catch (err) {
      console.error('Failed to trigger detectors:', err);
      setError(err instanceof Error ? err.message : 'Failed to trigger detectors');
    } finally {
      setTriggering(false);
    }
  };

  return (
    <div className="relative">
      {/* Trigger Button */}
      <button
        onClick={() => setShowPanel(!showPanel)}
        disabled={disabled}
        className="inline-flex items-center gap-2 rounded-md bg-green-600 px-3 py-1.5 text-sm font-medium text-white shadow-sm transition-all hover:bg-green-500 focus:outline-none focus:ring-2 focus:ring-green-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed"
      >
        <PlayIcon className="h-4 w-4" />
        Re-run Detectors
      </button>

      {/* Detector Selection Panel */}
      {showPanel && (
        <div className="absolute right-0 top-full mt-2 w-96 rounded-lg bg-white shadow-lg ring-1 ring-slate-200 z-10">
          <div className="p-4">
            {/* Header */}
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold text-slate-900">
                Select Detectors to Run
              </h3>
              <button
                onClick={() => setShowPanel(false)}
                className="text-slate-400 hover:text-slate-600"
              >
                <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>

            {/* Loading State */}
            {loading && (
              <div className="flex items-center justify-center py-8">
                <div className="h-8 w-8 animate-spin rounded-full border-4 border-slate-200 border-t-green-600" />
              </div>
            )}

            {/* Error State */}
            {error && (
              <div className="mb-4 rounded-lg bg-red-50 p-3 text-sm text-red-800">
                {error}
              </div>
            )}

            {/* Detector List */}
            {!loading && detectors.length > 0 && (
              <>
                <div className="space-y-2 mb-4 max-h-64 overflow-y-auto">
                  {detectors.map((detector) => (
                    <label
                      key={detector.name}
                      className={`flex items-start gap-3 p-3 rounded-md cursor-pointer transition-colors ${
                        detector.enabled
                          ? 'hover:bg-slate-50'
                          : 'opacity-50 cursor-not-allowed bg-slate-50'
                      }`}
                    >
                      <input
                        type="checkbox"
                        checked={selectedDetectors.includes(detector.name)}
                        onChange={() => handleToggleDetector(detector.name)}
                        disabled={!detector.enabled}
                        className="mt-0.5 h-4 w-4 rounded border-slate-300 text-green-600 focus:ring-green-500 disabled:cursor-not-allowed disabled:opacity-50"
                      />
                      <div className="flex-1">
                        <div className="text-sm font-medium text-slate-900">
                          {detector.label}
                        </div>
                        {!detector.enabled && (
                          <div className="text-xs text-slate-500 mt-1">
                            (Disabled)
                          </div>
                        )}
                      </div>
                    </label>
                  ))}
                </div>

                {/* Selection Summary */}
                <div className="mb-4 text-sm text-slate-600">
                  {selectedDetectors.length} of {detectors.filter(d => d.enabled).length} enabled detectors selected
                </div>

                {/* Actions */}
                <div className="flex gap-2">
                  <button
                    onClick={handleTrigger}
                    disabled={triggering || selectedDetectors.length === 0}
                    className="flex-1 inline-flex items-center justify-center gap-2 rounded-md bg-green-600 px-4 py-2 text-sm font-medium text-white shadow-sm transition-all hover:bg-green-500 focus:outline-none focus:ring-2 focus:ring-green-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {triggering ? (
                      <>
                        <div className="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent" />
                        Triggering...
                      </>
                    ) : (
                      <>
                        <PlayIcon className="h-4 w-4" />
                        Run Selected
                      </>
                    )}
                  </button>
                  <button
                    onClick={() => setShowPanel(false)}
                    className="px-4 py-2 text-sm font-medium text-slate-700 hover:text-slate-900 rounded-md hover:bg-slate-100 transition-colors"
                  >
                    Cancel
                  </button>
                </div>
              </>
            )}

            {/* Empty State */}
            {!loading && detectors.length === 0 && (
              <div className="text-center py-8 text-sm text-slate-500">
                No detectors available
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
};
