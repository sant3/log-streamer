import React, { useState, useRef, useEffect, useCallback, useMemo } from 'react';
import './App.css';
import Loader from './components/Loader';
import ServerPanel from './components/ServerPanel';
import LogControls from './components/LogControls';
import LogViewer from './components/LogViewer';
import useLogStream from './hooks/useLogStream';
import useServerStatus from './hooks/useServerStatus';
import useFileDownload from './hooks/useFileDownload';

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
  const [theme, setTheme] = useState('dark');
  const [logFile, setLogFile] = useState('');
  const [fontSize, setFontSize] = useState(14);
  const [showLineNumbers, setShowLineNumbers] = useState(true);
  const logContainerRef = useRef(null);
  const [autoScroll, setAutoScroll] = useState(true);
  const [availableLogFiles, setAvailableLogFiles] = useState([]);
  const [isLoadingServers, setIsLoadingServers] = useState(false);
  const [highlightText, setHighlightText] = useState('');

  const [isPanelOpen, setIsPanelOpen] = useState(true);
  const [servers, setServers] = useState([]);
  const [activeHost, setActiveHost] = useState('');

  const queryFileName = getQueryParam('file');

  // Autocomplete state
  const [suggestions, setSuggestions] = useState([]);
  const [isInputFocused, setIsInputFocused] = useState(false);
  const [activeSuggestionIndex, setActiveSuggestionIndex] = useState(-1);
  const suggestionsListRef = useRef(null);

  // Custom hooks
  const serverStatuses = useServerStatus(servers);
  const { logs, error, isStreaming, startStream, stopStream, clearLogs } = useLogStream(activeHost);
  const {
    download: downloadFile,
    cancel: cancelDownload,
    isDownloading,
    progress: downloadProgress,
    error: downloadError,
  } = useFileDownload(activeHost);

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
    setActiveSuggestionIndex(-1);
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

  const loadLogFiles = useCallback(async () => {
    if (!activeHost) return;
    setIsLoadingServers(true);
    try {
      const response = await fetch(`${activeHost}/list-files`);
      if (response.ok) {
        const logFiles = await response.json();
        setAvailableLogFiles(logFiles);
      } else {
        setAvailableLogFiles([]);
      }
    } catch (err) {
      setAvailableLogFiles([]);
    } finally {
      setIsLoadingServers(false);
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

  const handleStart = () => {
    startStream(logFile, logContainerRef);
  };

  const handleServerSelect = (serverUrl) => {
    setActiveHost(serverUrl);
    handleClear();
  };

  const handleClear = () => {
    clearLogs();
    setLogFile('');
    setSuggestions([]);
    setActiveSuggestionIndex(-1);
  };

  const increaseFontSize = () => {
    if (fontSize < 20) setFontSize(fontSize + 2);
  };
  const decreaseFontSize = () => {
    if (fontSize > 10) setFontSize(fontSize - 2);
  };

  const matchCount = useMemo(() => {
    if (!highlightText.trim() || logs.length === 0) return 0;
    const escapedText = highlightText.replace(/[-[\]{}()*+?.,\\^$|#\s]/g, '\\$&');
    const regex = new RegExp(escapedText, 'gi');
    return logs.reduce((acc, log) => {
      const matches = log.match(regex);
      return acc + (matches ? matches.length : 0);
    }, 0);
  }, [logs, highlightText]);

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
      <ServerPanel
        servers={servers}
        activeHost={activeHost}
        serverStatuses={serverStatuses}
        onServerSelect={handleServerSelect}
        isPanelOpen={isPanelOpen}
        onTogglePanel={() => setIsPanelOpen(!isPanelOpen)}
      />

      <div className="main-content">
        <div className="container">
          <div className="header">
            <div style={{ display: 'flex', alignItems: 'center' }}>
              <img src={process.env.PUBLIC_URL + '/main_logo.svg'} alt="Log Streamer Logo" className="main-logo" />
            </div>
            <LogControls
              logFile={logFile}
              onLogFileChange={handleLogFileChange}
              onKeyDown={handleKeyDown}
              isInputFocused={isInputFocused}
              onInputFocus={() => setIsInputFocused(true)}
              onInputBlur={() => setTimeout(() => setIsInputFocused(false), 200)}
              suggestions={suggestions}
              activeSuggestionIndex={activeSuggestionIndex}
              suggestionsListRef={suggestionsListRef}
              onSuggestionClick={handleSuggestionClick}
              onStart={handleStart}
              onStop={stopStream}
              onClear={handleClear}
              onIncreaseFontSize={increaseFontSize}
              onDecreaseFontSize={decreaseFontSize}
              onToggleLineNumbers={() => setShowLineNumbers(!showLineNumbers)}
              showLineNumbers={showLineNumbers}
              autoScroll={autoScroll}
              onToggleAutoScroll={() => setAutoScroll(!autoScroll)}
              isStreaming={isStreaming}
              highlightText={highlightText}
              onHighlightChange={(e) => setHighlightText(e.target.value)}
              onClearHighlight={() => setHighlightText('')}
              matchCount={matchCount}
              theme={theme}
              onToggleTheme={toggleTheme}
              onDownload={downloadFile}
              onCancelDownload={cancelDownload}
              isDownloading={isDownloading}
              downloadProgress={downloadProgress}
            />
          </div>
          <LogViewer
            logs={logs}
            error={error}
            fontSize={fontSize}
            showLineNumbers={showLineNumbers}
            highlightText={highlightText}
            logContainerRef={logContainerRef}
          />
          {downloadError && (
            <div className="download-error" role="alert">
              Download error: {downloadError}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

export default App;
