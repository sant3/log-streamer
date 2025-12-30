import React, { useState, useRef, useEffect, useCallback } from 'react';
import './App.css';

const getQueryParam = (param) => {
  const queryParams = new URLSearchParams(window.location.search);
  return queryParams.get(param);
};

const ensureUrlSchema = (url) => {
  if (!url) return '';
  if (!/^https?:\/\//i.test(url)) {
    return `http://${url}`;
  }
  return url;
};

const getFileNameQueryParam = () => {
  if (window.location.hash !== '') {
    const prefix = getQueryParam('file');
    return prefix + window.location.hash;
  } else {
    return getQueryParam('file');
  }
}

function App() {
  const [logs, setLogs] = useState([]);
  const [logFile, setLogFile] = useState('');
  const [error, setError] = useState('');
  const eventSourceRef = useRef(null);
  const [fontSize, setFontSize] = useState(14);
  const [showLineNumbers, setShowLineNumbers] = useState(true);
  const logContainerRef = useRef(null);
  const [autoScroll, setAutoScroll] = useState(true);
  const [availableLogFiles, setAvailableLogFiles] = useState([]);

  // New state for side panel and servers
  const [isPanelOpen, setIsPanelOpen] = useState(true);
  const [servers, setServers] = useState([]);
  const [serverStatuses, setServerStatuses] = useState({});
  
  // Use a more descriptive name for the currently selected host
  const [activeHost, setActiveHost] = useState(() => ensureUrlSchema(getQueryParam('host') || 'http://localhost:5005'));

  const queryFileName = getQueryParam('file');

  // Fetch the list of servers from the public JSON file
  useEffect(() => {
    fetch('/servers.json')
      .then(response => response.json())
      .then(data => {
        setServers(data);
        // If host is not set by query param, default to the first server in the list
        if (!getQueryParam('host') && data.length > 0) {
          setActiveHost(data[0].url);
        }
      })
      .catch(err => console.error("Failed to load servers.json:", err));
  }, []);

  // Function to check the status of all servers
  const updateServerStatuses = useCallback(async () => {
    if (servers.length === 0) return;

    const statuses = {};
    for (const server of servers) {
      try {
        const response = await fetch(`${server.url}/alive`, { signal: AbortSignal.timeout(5000) });
        statuses[server.name] = response.ok ? 'online' : 'offline';
      } catch (error) {
        statuses[server.name] = 'offline';
      }
    }
    setServerStatuses(statuses);
  }, [servers]);

  // Periodically update server statuses
  useEffect(() => {
    updateServerStatuses(); // Initial check
    const intervalId = setInterval(updateServerStatuses, 15000); // Check every 15 seconds
    return () => clearInterval(intervalId); // Cleanup on unmount
  }, [updateServerStatuses]);


  const loadLogFiles = useCallback(async () => {
    if (!activeHost) return;
    try {
      const response = await fetch(`${activeHost}/list-files`);
      if (response.ok) {
        const logFiles = await response.json();
        setAvailableLogFiles(logFiles);
      } else {
        setAvailableLogFiles([]);
        setError('Failed to load log files from backend.');
      }
    } catch (error) {
      setAvailableLogFiles([]);
      setError('Error fetching log files.');
    }
  }, [activeHost]);

  useEffect(() => {
    loadLogFiles();
  }, [loadLogFiles]);

  useEffect(() => {
    if (queryFileName != null) {
      const completFileName = getFileNameQueryParam();
      console.log(completFileName)
      setLogFile(completFileName)
    }
  }, [queryFileName]);

  const handleStart = async () => {
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
      if (logContainerRef.current) {
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
    };
    newEventSource.onerror = () => {
      console.error('EventSource failed.');
      setError('EventSource failed.');
      newEventSource.close();
    };
    eventSourceRef.current = newEventSource;
  };

  const checkBackendConnectivity = async () => {
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
    } catch (error) {
      setError('Failed to connect to backend. Please make sure the server is running.');
      return false;
    }
  };

  const handleServerSelect = (serverUrl) => {
    setActiveHost(serverUrl);
    // Clear logs and file selection when changing servers
    handleClear();
  };

  const handleStop = () => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
    }
  };

  const handleClear = () => {
    setLogs([]);
    setError('');
    handleStop();
    setLogFile('');
  };

  const increaseFontSize = () => setFontSize(fontSize + 2);
  const decreaseFontSize = () => {
    if (fontSize > 10) setFontSize(fontSize - 2);
  };
  const toggleLineNumbers = () => setShowLineNumbers(!showLineNumbers);

  useEffect(() => {
    if (autoScroll) {
      if (logContainerRef.current) {
        logContainerRef.current.scrollTop = logContainerRef.current.scrollHeight;
      }
    }
  }, [logs, autoScroll]);

  return (
    <div className={`app-wrapper ${isPanelOpen && servers.length > 1 ? 'panel-open' : ''}`}>
      {servers.length > 1 && (
        <>
          <div className="side-panel">
            <h3>Servers</h3>
            <ul className="server-list">
              {servers.map(server => (
                <li
                  key={server.name}
                  onClick={() => handleServerSelect(server.url)}
                  className={activeHost === server.url ? 'active' : ''}
                >
                  <span className={`status-dot ${serverStatuses[server.name] || 'offline'}`}></span>
                  {server.name}
                </li>
              ))}
            </ul>
          </div>
          <button onClick={() => setIsPanelOpen(!isPanelOpen)} className="panel-toggle">
            {isPanelOpen ? '<' : '>'}
          </button>
        </>
      )}

      <div className="main-content">
        <div className="container">
          <div className="header">
            <img src={process.env.PUBLIC_URL + '/main_logo.svg'} alt="Log Streamer Logo" className="main-logo" />
            <div className="controls">
              <input
                type="text"
                value={logFile}
                onChange={(e) => setLogFile(e.target.value)}
                placeholder="Enter log file name or select from combo"
                list="log-files"
              />
              <datalist id="log-files">
                {availableLogFiles && availableLogFiles.length > 0 ? (
                  availableLogFiles.map((file, index) => (
                    <option key={index} value={file}>{file}</option>
                  ))
                ) : (
                  <option disabled>No files available</option>
                )}
              </datalist>
              
              <div className="button-group">
                <button onClick={handleStart}>Start</button>
                <button onClick={handleStop}>Stop</button>
                <button onClick={handleClear}>Clear</button>
                <button onClick={increaseFontSize}>+ Font Size</button>
                <button onClick={decreaseFontSize}>- Font Size</button>
                <button onClick={toggleLineNumbers}>
                  {showLineNumbers ? 'Hide Line Numbers' : 'Show Line Numbers'}
                </button>
                <label>
                  <input
                    type="checkbox"
                    checked={autoScroll}
                    onChange={() => setAutoScroll(!autoScroll)}
                  />
                  Auto-Scroll
                </label>
              </div>
            </div>
          </div>
          <div className="log-container" ref={logContainerRef} style={{ fontSize: `${fontSize}px` }} data-testid="log-container">
            {error && <p className="error-message">{error}</p>}
            {logs.map((log, index) => (
              <div key={index} className="log-message">
                {showLineNumbers && <span className="line-number">{index + 1}. </span>}
                {log}
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

export default App;