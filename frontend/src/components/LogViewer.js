import React, { useMemo } from 'react';

const LogViewer = ({ logs, error, fontSize, showLineNumbers, highlightText, logContainerRef }) => {
  const renderLogMessage = (text) => {
    if (!highlightText.trim()) return text;
    const regex = new RegExp(`(${highlightText.replace(/[-[\]{}()*+?.,\\^$|#\s]/g, '\\$&')})`, 'gi');
    const parts = text.split(regex);
    return parts.map((part, i) =>
      part.toLowerCase() === highlightText.toLowerCase() ?
        <span key={i} className="log-highlight">{part}</span> : part
    );
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

  return (
    <div className="log-container" ref={logContainerRef} style={{ fontSize: `${fontSize}px` }} data-testid="log-container">
      {error && <p className="error-message">{error}</p>}
      {logs.map((log, index) => (
        <div key={index} className="log-message">
          {showLineNumbers && <span className="line-number">{index + 1}. </span>}
          {renderLogMessage(log)}
        </div>
      ))}
      {highlightText.trim() && (
        <span className="highlight-count" data-testid="match-count">{matchCount} matches</span>
      )}
    </div>
  );
};

export default LogViewer;
