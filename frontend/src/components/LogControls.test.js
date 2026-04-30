import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import LogControls from './LogControls';

describe('LogControls download button', () => {
  const baseProps = {
    logFile: '',
    onLogFileChange: jest.fn(),
    onKeyDown: jest.fn(),
    isInputFocused: false,
    onInputFocus: jest.fn(),
    onInputBlur: jest.fn(),
    suggestions: [],
    activeSuggestionIndex: -1,
    suggestionsListRef: { current: null },
    onSuggestionClick: jest.fn(),
    onStart: jest.fn(),
    onStop: jest.fn(),
    onClear: jest.fn(),
    onIncreaseFontSize: jest.fn(),
    onDecreaseFontSize: jest.fn(),
    onToggleLineNumbers: jest.fn(),
    showLineNumbers: true,
    autoScroll: true,
    onToggleAutoScroll: jest.fn(),
    isStreaming: false,
    highlightText: '',
    onHighlightChange: jest.fn(),
    onClearHighlight: jest.fn(),
    matchCount: 0,
    theme: 'dark',
    onToggleTheme: jest.fn(),
    onDownload: jest.fn(),
    onCancelDownload: jest.fn(),
    isDownloading: false,
    downloadProgress: 0,
  };

  test('Download button is disabled when logFile is empty', () => {
    render(<LogControls {...baseProps} logFile="" />);
    const btn = screen.getByRole('button', { name: /^download$/i });
    expect(btn).toBeDisabled();
  });

  test('Download button calls onDownload with current logFile', () => {
    const onDownload = jest.fn();
    render(<LogControls {...baseProps} logFile="app.log" onDownload={onDownload} />);
    fireEvent.click(screen.getByRole('button', { name: /^download$/i }));
    expect(onDownload).toHaveBeenCalledWith('app.log');
  });

  test('shows Cancel button with percentage during download', () => {
    render(<LogControls {...baseProps} logFile="app.log" isDownloading={true} downloadProgress={42} />);
    expect(screen.queryByRole('button', { name: /^download$/i })).toBeNull();
    expect(screen.getByRole('button', { name: /cancel \(42%\)/i })).toBeInTheDocument();
  });

  test('Cancel button calls onCancelDownload', () => {
    const onCancelDownload = jest.fn();
    render(
      <LogControls
        {...baseProps}
        logFile="app.log"
        isDownloading={true}
        downloadProgress={50}
        onCancelDownload={onCancelDownload}
      />
    );
    fireEvent.click(screen.getByRole('button', { name: /cancel/i }));
    expect(onCancelDownload).toHaveBeenCalled();
  });
});
