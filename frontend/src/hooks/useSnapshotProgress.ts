import { useState, useEffect, useRef } from 'react';
import { SnapshotStatus } from '../types';

interface UseSnapshotProgressReturn {
  status: SnapshotStatus | null;
  isConnected: boolean;
  error: string | null;
}

const MAX_RETRIES = 3;
const INITIAL_RETRY_DELAY = 1000; // 1 second

// EventSource doesn't respect Vite proxy, so we need the full URL
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

export function useSnapshotProgress(jobId: string | null): UseSnapshotProgressReturn {
  const [status, setStatus] = useState<SnapshotStatus | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const eventSourceRef = useRef<EventSource | null>(null);
  const retryCountRef = useRef(0);
  const retryTimeoutRef = useRef<number | null>(null);
  const jobCompletedRef = useRef(false);

  useEffect(() => {
    console.log('useSnapshotProgress - jobId:', jobId);
    if (!jobId) {
      setStatus(null);
      setIsConnected(false);
      setError(null);
      jobCompletedRef.current = false;
      return;
    }

    // Reset completion flag for new job
    jobCompletedRef.current = false;

    // Check if EventSource is supported
    if (typeof EventSource === 'undefined') {
      setError('Live updates unavailable (browser not supported)');
      return;
    }

    const connectToSSE = () => {
      // Clean up existing connection
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
      }

      try {
        // EventSource doesn't respect Vite proxy, use full URL with credentials
        const url = `${API_BASE_URL}/api/snapshot/progress/${jobId}`;
        console.log('Connecting to SSE:', url);
        const eventSource = new EventSource(url, { withCredentials: true });
        eventSourceRef.current = eventSource;

        eventSource.addEventListener('connected', () => {
          console.log('SSE connection established');
          setIsConnected(true);
          setError(null);
          retryCountRef.current = 0; // Reset retry count on successful connection
        });

        eventSource.addEventListener('progress', (event) => {
          console.log('Received progress event, data:', event.data);
          try {
            const progressData: SnapshotStatus = JSON.parse(event.data);
            console.log('Parsed progress data:', progressData);
            setStatus(progressData);
            setIsConnected(true);
            setError(null);

            // Mark as completed if status is completed or cancelled
            if (progressData.status === 'completed' || progressData.status === 'cancelled') {
              jobCompletedRef.current = true;
              // Close connection after a brief delay
              setTimeout(() => {
                eventSource.close();
                setIsConnected(false);
              }, 500);
            }
          } catch (err) {
            console.error('Failed to parse progress data:', err);
            console.error('Raw event.data:', event.data);
          }
        });

        eventSource.addEventListener('complete', (event) => {
          try {
            const completeData: SnapshotStatus = JSON.parse(event.data);
            setStatus(completeData);
            jobCompletedRef.current = true;
            // Close connection on completion
            setTimeout(() => {
              eventSource.close();
              setIsConnected(false);
            }, 500);
          } catch (err) {
            console.error('Failed to parse completion data:', err);
          }
        });

        eventSource.addEventListener('error', (event) => {
          // This handles error messages sent by the server as SSE events
          const messageEvent = event as MessageEvent;
          if (messageEvent.data) {
            try {
              const errorData: SnapshotStatus = JSON.parse(messageEvent.data);
              setStatus(errorData);
            } catch (err) {
              console.error('Failed to parse error event data:', err);
            }
          }
          // Connection-level errors are handled by onerror below
        });

        eventSource.onerror = (event) => {
          console.error('SSE connection error:', event);
          console.log('EventSource readyState:', eventSource.readyState);
          console.log('EventSource url:', eventSource.url);
          setIsConnected(false);

          // Close the failed connection
          eventSource.close();

          // Don't retry if job has completed - connection close is expected
          if (jobCompletedRef.current) {
            console.log('Job completed, not retrying SSE connection');
            return;
          }

          // Implement exponential backoff retry
          if (retryCountRef.current < MAX_RETRIES) {
            const delay = INITIAL_RETRY_DELAY * Math.pow(2, retryCountRef.current);
            console.log(`Retrying SSE connection in ${delay}ms (attempt ${retryCountRef.current + 1}/${MAX_RETRIES})`);

            retryTimeoutRef.current = setTimeout(() => {
              retryCountRef.current++;
              connectToSSE();
            }, delay);
          } else {
            // Max retries exceeded
            console.error('Max SSE connection retries exceeded');
            setError('Live updates unavailable');
            setIsConnected(false);
          }
        };
      } catch (err) {
        console.error('Failed to create EventSource:', err);
        setError('Live updates unavailable');
        setIsConnected(false);
      }
    };

    // Initial connection
    connectToSSE();

    // Cleanup function
    return () => {
      if (retryTimeoutRef.current) {
        clearTimeout(retryTimeoutRef.current);
      }
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
        eventSourceRef.current = null;
      }
      setIsConnected(false);
    };
  }, [jobId]);

  return { status, isConnected, error };
}
