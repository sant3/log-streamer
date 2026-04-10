import { useState, useRef, useCallback } from 'react';

const useLogStream = (activeHost) => {
  const [logs, setLogs] = useState([]);
  const [error, setError] = useState('');
  const [isStreaming, setIsStreaming] = useState(false);
  const eventSourceRef = useRef(null);

  const checkBackendConnectivity = useCallback(async () => {
    if (!activeHost) return false;
    try {
      const response = await fetch(`${activeHost}/alive`);
      if (response.ok) {
        setError('');
        return true;
      } else {
        setError('Backend is unreachable. Please check the connection.');
        return false;
      }
    } catch (err) {
      setError('Failed to connect to backend. Please make sure the server is running.');
      return false;
    }
  }, [activeHost]);

  const startStream = useCallback(async (logFile, logContainerRef) => {
    if (!logFile) {
      setError('Please specify a log file.');
      return;
    }
    if (!activeHost) {
      setError('No active server selected.');
      return;
    }

    if (eventSourceRef.current) {
      eventSourceRef.current.close();
    }

    const isBackendReachable = await checkBackendConnectivity();
    if (!isBackendReachable) {
      if (logContainerRef && logContainerRef.current) {
        logContainerRef.current.scrollTop = 0;
      }
      return;
    }

    const encodedLogFile = encodeURIComponent(logFile);
    const newEventSource = new EventSource(`${activeHost}/stream-logs?file=${encodedLogFile}`);
    newEventSource.onmessage = (event) => {
      if (event.data.startsWith("Error:")) {
        setError(event.data);
        return;
      }
      setError('');
      setLogs((prevLogs) => [...prevLogs, event.data]);
      setIsStreaming(true);
    };
    newEventSource.onerror = () => {
      console.error('EventSource failed.');
      setError('EventSource failed.');
      setIsStreaming(false);
      newEventSource.close();
    };
    eventSourceRef.current = newEventSource;
  }, [activeHost, checkBackendConnectivity]);

  const stopStream = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      setIsStreaming(false);
    }
  }, []);

  const clearLogs = useCallback(() => {
    setLogs([]);
    setError('');
    stopStream();
  }, [stopStream]);

  return {
    logs,
    error,
    isStreaming,
    startStream,
    stopStream,
    clearLogs,
    setError,
  };
};

export default useLogStream;
