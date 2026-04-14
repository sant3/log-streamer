import { render, screen } from '@testing-library/react';
import LogViewer from './LogViewer';
import React from 'react';

describe('LogViewer', () => {
  const defaultProps = {
    logs: [],
    error: '',
    fontSize: 14,
    showLineNumbers: true,
    highlightText: '',
    logContainerRef: React.createRef(),
  };

  test('renders empty log container', () => {
    render(<LogViewer {...defaultProps} />);
    const container = screen.getByTestId('log-container');
    expect(container).toBeInTheDocument();
  });

  test('renders log messages', () => {
    render(<LogViewer {...defaultProps} logs={['Line one', 'Line two']} />);
    expect(screen.getByText('Line one')).toBeInTheDocument();
    expect(screen.getByText('Line two')).toBeInTheDocument();
  });

  test('shows line numbers when enabled', () => {
    const { container } = render(
      <LogViewer {...defaultProps} logs={['Test line']} showLineNumbers={true} />
    );
    const lineNumber = container.querySelector('.line-number');
    expect(lineNumber).toBeInTheDocument();
    expect(lineNumber.textContent).toBe('1. ');
  });

  test('hides line numbers when disabled', () => {
    const { container } = render(
      <LogViewer {...defaultProps} logs={['Test line']} showLineNumbers={false} />
    );
    const lineNumber = container.querySelector('.line-number');
    expect(lineNumber).not.toBeInTheDocument();
  });

  test('displays error message', () => {
    render(<LogViewer {...defaultProps} error="Connection failed" />);
    expect(screen.getByText('Connection failed')).toBeInTheDocument();
  });

  test('applies font size style', () => {
    render(<LogViewer {...defaultProps} fontSize={18} />);
    const container = screen.getByTestId('log-container');
    expect(container).toHaveStyle('font-size: 18px');
  });

  test('highlights matching text', () => {
    const { container } = render(
      <LogViewer {...defaultProps} logs={['Error occurred in module']} highlightText="Error" />
    );
    const highlight = container.querySelector('.log-highlight');
    expect(highlight).toBeInTheDocument();
    expect(highlight.textContent).toBe('Error');
  });

  test('shows match count when highlighting', () => {
    render(
      <LogViewer
        {...defaultProps}
        logs={['Error one', 'Error two', 'No match']}
        highlightText="Error"
      />
    );
    expect(screen.getByText('2 matches')).toBeInTheDocument();
  });

  test('shows 0 matches for non-matching highlight', () => {
    render(
      <LogViewer
        {...defaultProps}
        logs={['Line one', 'Line two']}
        highlightText="xyz"
      />
    );
    expect(screen.getByText('0 matches')).toBeInTheDocument();
  });
});
