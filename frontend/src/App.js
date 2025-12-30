import React, { useState, useRef, useEffect } from 'react';
import './App.css';


const getQueryParam = (param) => {
  const queryParams = new URLSearchParams(window.location.search);
  return queryParams.get(param);
};

const ensureUrlSchema = (url) => {
  if (!/^https?:\/\//i.test(url)) {
    return `http://${url}`; // add "http://" if missing
  }
  return url;
};

const getFileNameQueryParam = () => {
  if (window.location.hash !==''){
      const prefix = getQueryParam('file');
      return prefix + window.location.hash;
  }else {
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

  // const jwtToken = "b"; // Il token JWT dovrebbe essere memorizzato qui
  // const jwtToken = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiOjE3MjM2MzE4MzIsImV4cCI6MTkxMjkzNDIwNywiYXVkIjoid3d3LmV4YW1wbGUuY29tIiwic3ViIjoianJvY2"; // Il token JWT non valido

  const defaultHost = 'http://localhost:5005'; // default Host
  const queryHost = getQueryParam('host');
  const backendHost = ensureUrlSchema(queryHost || defaultHost);

  const queryFileName = getQueryParam('file');

  

  const loadLogFiles = async () => {
    try {
      const response = await fetch(`${backendHost}/list-files`);
      if (response.ok) {
        const logFiles = await response.json();
        setAvailableLogFiles(logFiles);
      } else {
        setError('Failed to load log files from backend.');
      }
    } catch (error) {
      setError('Error fetching log files.');
    }
  };

  useEffect(() => {
    loadLogFiles();
  }, []);

  useEffect(() => {
    if (queryFileName != null){
      const completFileName = getFileNameQueryParam();
      console.log(completFileName)
      setLogFile(completFileName)
    
    }
  }, []);

  const handleStart = async () => {
    if (!logFile) {
      setError('Please specify a log file.');
      return;
    }

    if (eventSourceRef.current) {
      eventSourceRef.current.close();
    }

    // Check BE connection before start streaming log
    const isBackendReachable = await checkBackendConnectivity();
    if (!isBackendReachable) {
      const logContainer = logContainerRef.current;
      if (logContainer) {
        logContainer.scrollTop = 0; // Scroll to the top
      }

      return; // Don't start streaming if BE did not respond 
    }

    const encodedLogFile = encodeURIComponent(logFile);

    const newEventSource = new EventSource(`${backendHost}/stream-logs?file=${encodedLogFile}`);
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
    try {
      
      // axios.get(`${backendHost}/alive`, {
      //   headers: {
      //     'Authorization': `Bearer ${jwtToken}`,
      //     "Content-Type": "application/json",
      //   },
      // });
      
      const response = await fetch(`${backendHost}/alive`, {
        headers: {
          // "Authorization": `Bearer ${jwtToken}`,
          "Content-Type": "application/json",
        },
      });

      if (response.ok) {
        setError(''); // Backend is reachable
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

  const increaseFontSize = () => {
    setFontSize(fontSize + 2); // Increase font size by 2px
  };

  const decreaseFontSize = () => {
    if (fontSize > 10) { // Set minimum font size to 10px
      setFontSize(fontSize - 2);
    }
  };

  const toggleLineNumbers = () => {
    setShowLineNumbers(!showLineNumbers); // Toggle state
  };


  useEffect(() => {
    if (autoScroll) {
      const logContainer = logContainerRef.current;
      if (logContainer) {
        logContainer.scrollTop = logContainer.scrollHeight; // Scroll to the bottom
      }
    }
  }, [logs, autoScroll]);

  return (
    <div className="container">
      <div className="header">

        <input
          type="text"
          value={logFile}
          onChange={(e) => setLogFile(e.target.value)}
          placeholder="Enter log file name or select from combo"
          list="log-files" // Associate list file name to input
        />
        
       

        <datalist id="log-files">
          {availableLogFiles && availableLogFiles.length > 0 ? (
            availableLogFiles.map((file, index) => (
              <option key={index} value={file}>
                {file}
              </option>
            ))
          ) : (
            <option disabled>No files available</option>
          )}
        </datalist>


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
      <div className="log-container" ref={logContainerRef} style={{ fontSize: `${fontSize}px` }} data-testid="log-container">
        {error && <p className="error-message">{error}</p>}
        {logs.map((log, index) => (
          <div key={index} className="log-message">
            {showLineNumbers && <span className="line-number">{index + 1}. </span>} {/* Conditionally render line number */}
            {log}
          </div>
        ))}
      </div>
    </div>
  );
}

export default App;
