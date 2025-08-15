import { useState, useEffect, useRef, useCallback } from 'react';
import { billService, BillStatus } from '../services/billService';

interface UseBillStatusOptions {
  billId: string;
  pollInterval?: number; // in milliseconds
  onStatusChange?: (status: string) => void;
  onComplete?: (billStatus: BillStatus) => void;
  onError?: (error: Error) => void;
}

export function useBillStatus({
  billId,
  pollInterval = 2000, // Default to 2 seconds
  onStatusChange,
  onComplete,
  onError
}: UseBillStatusOptions) {
  const [status, setStatus] = useState<string>('pending');
  const [isPolling, setIsPolling] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);
  
  const intervalRef = useRef<NodeJS.Timeout | null>(null);
  const abortControllerRef = useRef<AbortController | null>(null);

  const startPolling = useCallback(() => {
    if (isPolling) return;
    
    setIsPolling(true);
    setError(null);
    
    // Create abort controller for cleanup
    abortControllerRef.current = new AbortController();
    
    const poll = async () => {
      try {
        if (abortControllerRef.current?.signal.aborted) return;
        
        const billStatus = await billService.getBillStatus(billId);
        setStatus(billStatus.status);
        setLastUpdated(new Date());
        
        // Call status change callback
        onStatusChange?.(billStatus.status);
        
        // Check if processing is complete
        if (billStatus.status === 'completed' || billStatus.status === 'failed') {
          setIsPolling(false);
          onComplete?.(billStatus);
          
          // Clear interval
          if (intervalRef.current) {
            clearInterval(intervalRef.current);
            intervalRef.current = null;
          }
          return;
        }
      } catch (err) {
        if (err instanceof Error && err.name === 'AbortError') return;
        
        const error = err instanceof Error ? err : new Error('Unknown error occurred');
        setError(error);
        onError?.(error);
        
        // Stop polling on error
        setIsPolling(false);
        if (intervalRef.current) {
          clearInterval(intervalRef.current);
          intervalRef.current = null;
        }
      }
    };

    // Initial poll
    poll();
    
    // Set up interval for subsequent polls
    intervalRef.current = setInterval(poll, pollInterval);
  }, [billId, pollInterval, isPolling, onStatusChange, onComplete, onError]);

  const stopPolling = () => {
    setIsPolling(false);
    
    // Clear interval
    if (intervalRef.current) {
      clearInterval(intervalRef.current);
      intervalRef.current = null;
    }
    
    // Abort any pending requests
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
      abortControllerRef.current = null;
    }
  };

  const reset = () => {
    stopPolling();
    setStatus('pending');
    setError(null);
    setLastUpdated(null);
  };

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      stopPolling();
    };
  }, []);

  // Auto-start polling when billId changes
  useEffect(() => {
    if (billId) {
      startPolling();
    }
    
    return () => {
      stopPolling();
    };
  }, [billId, startPolling]);

  return {
    status,
    isPolling,
    error,
    lastUpdated,
    startPolling,
    stopPolling,
    reset
  };
}
