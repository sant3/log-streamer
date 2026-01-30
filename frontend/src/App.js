import React, { useState, useRef, useEffect, useCallback } from 'react';
import './App.css';
import ThemeSwitcher from './components/ThemeSwitcher'; // Import the ThemeSwitcher component
import Loader from './components/Loader'; // Import the Loader component

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

const POLLING_ALIVE_MS = 10000;

function App() {
  const [theme, setTheme] = useState('dark'); // Initialize theme state to 'dark'
  const [logs, setLogs] = useState([]);
  const [logFile, setLogFile] = useState('');
  const [error, setError] = useState('');
  const eventSourceRef = useRef(null);
  const [fontSize, setFontSize] = useState(14);
  const [showLineNumbers, setShowLineNumbers] = useState(true);
  const logContainerRef = useRef(null);
  const [autoScroll, setAutoScroll] = useState(true);
  const [availableLogFiles, setAvailableLogFiles] = useState([]);
  const [isLoadingServers, setIsLoadingServers] = useState(false); // New state for loader
  const [isStreaming, setIsStreaming] = useState(false); // New state for streaming status

  // New state for side panel and servers
  const [isPanelOpen, setIsPanelOpen] = useState(true);
  const [servers, setServers] = useState([]);
  const [serverStatuses, setServerStatuses] = useState({});

  // Use a more descriptive name for the currently selected host
  const [activeHost, setActiveHost] = useState('');

  const queryFileName = getQueryParam('file');

  // Autocomplete state
  const [suggestions, setSuggestions] = useState([]);
  const [isInputFocused, setIsInputFocused] = useState(false);
  const [activeSuggestionIndex, setActiveSuggestionIndex] = useState(-1);
  const suggestionsListRef = useRef(null);

  // Load servers from global variable
  useEffect(() => {
    const serversData = window.APP_SERVERS || [];
    setServers(serversData);
    const hostParam = getQueryParam('host');
    if (hostParam) {
      setActiveHost(ensureUrlSchema(hostParam));
    } else if (serversData.length > 0) {
      setActiveHost(serversData[0].url);
    } else {
      setActiveHost('http://localhost:5005');
    }
  }, []);

  // Effect to update suggestions when dependencies change
  useEffect(() => {
    if (isInputFocused) {
      const filtered = availableLogFiles.filter(file =>
        file.toLowerCase().includes(logFile.toLowerCase())
      );
      setSuggestions(filtered);
    } else {
      setSuggestions([]);
    }
  }, [isInputFocused, logFile, availableLogFiles]);


  const toggleTheme = () => {
    setTheme((prevTheme) => (prevTheme === 'dark' ? 'light' : 'dark'));
  };

  const handleLogFileChange = (e) => {
    setLogFile(e.target.value);
    setActiveSuggestionIndex(-1); // Reset active suggestion on text change
  };

  const handleSuggestionClick = (suggestion) => {
    setLogFile(suggestion);
    setIsInputFocused(false);
    setActiveSuggestionIndex(-1);
  };

  const handleKeyDown = (e) => {
    if (suggestions.length === 0) return;

    if (e.key === 'ArrowDown') {
      e.preventDefault();
      setActiveSuggestionIndex(prevIndex =>
        prevIndex < suggestions.length - 1 ? prevIndex + 1 : 0
      );
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      setActiveSuggestionIndex(prevIndex =>
        prevIndex > 0 ? prevIndex - 1 : suggestions.length - 1
      );
    } else if (e.key === 'Enter') {
      e.preventDefault();
      if (activeSuggestionIndex !== -1) {
        handleSuggestionClick(suggestions[activeSuggestionIndex]);
      }
    } else if (e.key === 'Escape') {
      setIsInputFocused(false);
    }
  };

  useEffect(() => {
    if (activeSuggestionIndex !== -1 && suggestionsListRef.current) {
      const activeItem = suggestionsListRef.current.children[activeSuggestionIndex];
      if (activeItem) {
        activeItem.scrollIntoView({
          block: 'nearest',
        });
      }
    }
  }, [activeSuggestionIndex]);

  // Function to check the status of all servers
  const updateServerStatuses = useCallback(async () => {
    if (servers.length === 0) return;

    const statuses = {};
    for (const server of servers) {
      try {
        const response = await fetch(`${server.url}/alive`, { signal: AbortSignal.timeout(POLLING_ALIVE_MS) });
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
    setIsLoadingServers(true); // Start loading
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
    } finally {
      setIsLoadingServers(false); // End loading
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
      setIsStreaming(true);
    };
    newEventSource.onerror = () => {
      console.error('EventSource failed.');
      setError('EventSource failed.');
      setIsStreaming(false);
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
    // loadLogFiles will be triggered by activeHost change in useEffect
  };

  const handleStop = () => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      setIsStreaming(false);
    }
  };

  const handleClear = () => {
    setLogs([]);
    setError('');
    handleStop();
    setLogFile('');
    setSuggestions([]);
    setActiveSuggestionIndex(-1);
  };

  const increaseFontSize = () => {
    if (fontSize < 20)
      setFontSize(fontSize + 2);
  }
  const decreaseFontSize = () => {
    if (fontSize > 10)
      setFontSize(fontSize - 2);
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
    <div className={`app-wrapper ${isPanelOpen && servers.length > 1 ? 'panel-open' : ''} ${theme}`}>
      {isLoadingServers && <Loader />}
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
            <div style={{ display: 'flex', alignItems: 'center' }}>
              <img src={process.env.PUBLIC_URL + '/main_logo.svg'} alt="Log Streamer Logo" className="main-logo" />
            </div>
            <div className="controls">
              <div className="autocomplete-wrapper">
                <input
                  type="text"
                  value={logFile}
                  onChange={handleLogFileChange}
                  onFocus={() => setIsInputFocused(true)}
                  onBlur={() => setTimeout(() => setIsInputFocused(false), 200)} // Delay to allow click
                  onKeyDown={handleKeyDown}
                  placeholder="Enter log file name"
                  autoComplete="off"
                />
                {isInputFocused && suggestions.length > 0 && (
                  <ul className="suggestions-list" ref={suggestionsListRef}>
                    {suggestions.map((file, index) => (
                      <li
                        key={index}
                        className={index === activeSuggestionIndex ? 'active' : ''}
                        onMouseDown={() => handleSuggestionClick(file)}
                      >
                        {file}
                      </li>
                    ))}
                  </ul>
                )}
              </div>

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
                <div className={`streaming-indicator ${isStreaming ? 'active' : 'inactive'}`} title={isStreaming ? 'Streaming Active' : 'Streaming Inactive'}></div>
                <ThemeSwitcher theme={theme} toggleTheme={toggleTheme} /> {/* Moved ThemeSwitcher here */}
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