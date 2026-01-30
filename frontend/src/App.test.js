import { render, screen, fireEvent, act, waitFor } from '@testing-library/react';
import App from './App';

// Mock scrollIntoView for JSDOM
window.HTMLElement.prototype.scrollIntoView = jest.fn();

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
          json: () => Promise.resolve(['system.log', 'application.log', 'another.log']),
        });
      }
      return Promise.resolve({ ok: true }); // Default success for other calls like /alive
    });

    // Reset window.APP_SERVERS
    delete window.APP_SERVERS;
  });

  test('renders the main application controls', async () => {
    // Use act to ensure all initial renders and effects are processed
    await act(async () => {
      render(<App />);
    });
    expect(screen.getByPlaceholderText(/enter log file name/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /start/i })).toBeInTheDocument();
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

  describe('Theme Switcher', () => {
    test('defaults to dark theme and toggles correctly', async () => {
      await act(async () => {
        render(<App />);
      });
      const appWrapper = screen.getByTestId('log-container').closest('.app-wrapper');
  
      // Defaults to dark
      expect(appWrapper).toHaveClass('dark');
  
      const themeSwitcher = screen.getByRole('button', { name: '' }).closest('.theme-switcher-button');
  
      // Toggle to light
      await act(async () => {
        fireEvent.click(themeSwitcher);
      });
      expect(appWrapper).not.toHaveClass('dark');
  
      // Toggle back to dark
      await act(async () => {
        fireEvent.click(themeSwitcher);
      });
      expect(appWrapper).toHaveClass('dark');
    });
  });
  
  describe('Server Loading', () => {
    test('loads servers from window.APP_SERVERS when more than one', async () => {
      window.APP_SERVERS = [
        { name: "Global Test Server 1", url: "http://global.test1" },
        { name: "Global Test Server 2", url: "http://global.test2" },
      ];
      await act(async () => {
        render(<App />);
      });
  
      expect(screen.getByText('Global Test Server 1')).toBeInTheDocument();
      expect(screen.getByText('Global Test Server 2')).toBeInTheDocument();
    });
  });
  
  describe('Autocomplete Feature', () => {
    beforeEach(() => {
      window.APP_SERVERS = [{ name: "Test Server", url: "http://test.local" }];
    });
  
    test('filters suggestions based on input', async () => {
      render(<App />);
      const logInput = screen.getByPlaceholderText(/enter log file name/i);
      
      // Wait for files to be loaded
      await waitFor(() => expect(fetch).toHaveBeenCalledWith('http://test.local/list-files'));
  
      // Focus to show initial suggestions
      fireEvent.focus(logInput);
      
      // Type to filter
      fireEvent.change(logInput, { target: { value: 'app' } });
  
      const suggestions = await screen.findAllByRole('listitem');
      expect(suggestions).toHaveLength(1);
      expect(suggestions[0]).toHaveTextContent('application.log');
      expect(screen.queryByText('system.log')).not.toBeInTheDocument();
    });
  
    test('keyboard navigation works correctly', async () => {
      render(<App />);
      const logInput = screen.getByPlaceholderText(/enter log file name/i);
      
      // Wait for files to be loaded and focus
      await waitFor(() => expect(fetch).toHaveBeenCalledWith('http://test.local/list-files'));
      fireEvent.focus(logInput);
  
      const suggestions = await screen.findAllByRole('listitem');
      expect(suggestions.length).toBe(3);
  
      // Navigate down
      fireEvent.keyDown(logInput, { key: 'ArrowDown' });
      expect(suggestions[0]).toHaveClass('active');
      expect(suggestions[1]).not.toHaveClass('active');
  
      fireEvent.keyDown(logInput, { key: 'ArrowDown' });
      expect(suggestions[0]).not.toHaveClass('active');
      expect(suggestions[1]).toHaveClass('active');
  
      // Navigate up
      fireEvent.keyDown(logInput, { key: 'ArrowUp' });
      expect(suggestions[0]).toHaveClass('active');
      expect(suggestions[1]).not.toHaveClass('active');
  
      // Select with Enter
      fireEvent.keyDown(logInput, { key: 'Enter' });
  
      // Input value should be updated and list should disappear
      expect(logInput.value).toBe('system.log');
      expect(screen.queryAllByRole('listitem')).toHaveLength(0);
    });
  });
});
