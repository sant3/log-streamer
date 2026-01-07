import { render, screen, fireEvent, act } from '@testing-library/react';
import App from './App';

// Mock the global fetch function
globalThis.fetch = jest.fn();

// Mock for EventSource
class MockEventSource {
  constructor(url) {
    this.url = url;
    this.onmessage = null;
    this.onerror = null;
  }
  close() {}
}
globalThis.EventSource = MockEventSource;


describe('App component', () => {
  beforeEach(() => {
    // Reset mocks before each test
    fetch.mockClear();
    fetch.mockImplementation((url) => {
      if (url.endsWith('/list-files')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve(['system.log', 'application.log']),
        });
      }
      if (url.endsWith('/servers.json')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve([
            { name: "Test Server 1", url: "http://test1.local" },
            { name: "Test Server 2", url: "http://test2.local" },
          ]),
        });
      }
      return Promise.resolve({ ok: true }); // Default success for other calls like /alive
    });
  });

  test('renders the main application controls', async () => {
    // Use act to ensure all initial renders and effects are processed
    await act(async () => {
      render(<App />);
    });
    expect(screen.getByPlaceholderText(/enter log file name/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /start/i })).toBeInTheDocument();
  });

  test('loads and displays available log files on initial render', async () => {
    await act(async () => {
      render(<App />);
    });
    
    // It will wait for the text to appear in the DOM.
    expect(await screen.findByText('system.log')).toBeInTheDocument();
    expect(await screen.findByText('application.log')).toBeInTheDocument();
  });

  test('clears logs and input when clear button is clicked', async () => {
    await act(async () => {
      render(<App />);
    });
    const logInput = screen.getByPlaceholderText(/enter log file name/i);
    const clearButton = screen.getByRole('button', { name: /clear/i });

    // Set a file name
    fireEvent.change(logInput, { target: { value: 'system.log' } });
    expect(logInput.value).toBe('system.log');
    
    // Click the clear button
    fireEvent.click(clearButton);

    // Assert that the input field is now empty
    expect(logInput.value).toBe('');
  });

  test('font size buttons modify the font size of the log container', async () => {
    await act(async () => {
      render(<App />);
    });
    const logContainer = screen.getByTestId('log-container');
    const increaseButton = screen.getByRole('button', { name: /\+ font size/i });
    const decreaseButton = screen.getByRole('button', { name: /- font size/i });

    const initialSize = 14;
    expect(logContainer).toHaveStyle(`font-size: ${initialSize}px`);

    fireEvent.click(increaseButton);
    expect(logContainer).toHaveStyle(`font-size: ${initialSize + 2}px`);

    fireEvent.click(decreaseButton);
    expect(logContainer).toHaveStyle(`font-size: ${initialSize}px`);
  });
});
