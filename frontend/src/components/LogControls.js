import React from 'react';
import ThemeSwitcher from './ThemeSwitcher';

const LogControls = ({
  logFile,
  onLogFileChange,
  onKeyDown,
  isInputFocused,
  onInputFocus,
  onInputBlur,
  suggestions,
  activeSuggestionIndex,
  suggestionsListRef,
  onSuggestionClick,
  onStart,
  onStop,
  onClear,
  onIncreaseFontSize,
  onDecreaseFontSize,
  onToggleLineNumbers,
  showLineNumbers,
  autoScroll,
  onToggleAutoScroll,
  isStreaming,
  highlightText,
  onHighlightChange,
  onClearHighlight,
  matchCount,
  theme,
  onToggleTheme,
}) => {
  return (
    <div className="controls">
      <div className="autocomplete-wrapper">
        <input
          type="text"
          value={logFile}
          onChange={onLogFileChange}
          onFocus={onInputFocus}
          onBlur={onInputBlur}
          onKeyDown={onKeyDown}
          placeholder="Enter log file name"
          autoComplete="off"
        />
        {isInputFocused && suggestions.length > 0 && (
          <ul className="suggestions-list" ref={suggestionsListRef}>
            {suggestions.map((file, index) => (
              <li
                key={index}
                className={index === activeSuggestionIndex ? 'active' : ''}
                onMouseDown={() => onSuggestionClick(file)}
              >
                {file}
              </li>
            ))}
          </ul>
        )}
      </div>

      <div className="button-group">
        <button onClick={onStart}>Start</button>
        <button onClick={onStop}>Stop</button>
        <button onClick={onClear}>Clear</button>
        <button onClick={onIncreaseFontSize}>+ Font Size</button>
        <button onClick={onDecreaseFontSize}>- Font Size</button>
        <button onClick={onToggleLineNumbers}>
          {showLineNumbers ? 'Hide Line Numbers' : 'Show Line Numbers'}
        </button>
        <label>
          <input
            type="checkbox"
            checked={autoScroll}
            onChange={onToggleAutoScroll}
          />
          Auto-Scroll
        </label>
        <div className={`streaming-indicator ${isStreaming ? 'active' : 'inactive'}`} title={isStreaming ? 'Streaming Active' : 'Streaming Inactive'}></div>
        <div className="highlight-wrapper">
          <input
            type="text"
            value={highlightText}
            onChange={onHighlightChange}
            placeholder="Highlight..."
            className="highlight-input"
          />
          {highlightText && (
            <button className="clear-highlight" onClick={onClearHighlight}>×</button>
          )}
          {highlightText.trim() && (
            <span className="highlight-count">{matchCount} matches</span>
          )}
        </div>
        <ThemeSwitcher theme={theme} toggleTheme={onToggleTheme} />
      </div>
    </div>
  );
};

export default LogControls;
