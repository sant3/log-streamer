import { renderHook, act, waitFor } from '@testing-library/react';
import useFileDownload from './useFileDownload';

const HOST = 'http://localhost:5005';

function makeStreamResponse({ chunks, contentLength, status = 200 }) {
  let i = 0;
  const body = {
    getReader() {
      return {
        read() {
          if (i >= chunks.length) return Promise.resolve({ done: true });
          const value = chunks[i++];
          return Promise.resolve({ done: false, value });
        },
      };
    },
  };
  return {
    ok: status >= 200 && status < 300,
    status,
    headers: {
      get: (h) => (h.toLowerCase() === 'content-length' ? contentLength : null),
    },
    body,
  };
}

describe('useFileDownload', () => {
  let originalFetch;
  let originalCreateObjectURL;
  let originalRevokeObjectURL;
  let createdUrls;

  beforeEach(() => {
    originalFetch = global.fetch;
    originalCreateObjectURL = global.URL.createObjectURL;
    originalRevokeObjectURL = global.URL.revokeObjectURL;
    createdUrls = [];
    global.URL.createObjectURL = jest.fn(() => {
      const u = `blob:fake-${createdUrls.length}`;
      createdUrls.push(u);
      return u;
    });
    global.URL.revokeObjectURL = jest.fn();
  });

  afterEach(() => {
    global.fetch = originalFetch;
    global.URL.createObjectURL = originalCreateObjectURL;
    global.URL.revokeObjectURL = originalRevokeObjectURL;
    jest.restoreAllMocks();
  });

  test('downloads file successfully and updates progress', async () => {
    const chunks = [
      new Uint8Array([1, 2, 3, 4, 5]),
      new Uint8Array([6, 7, 8, 9, 10]),
    ];
    global.fetch = jest.fn(() =>
      Promise.resolve(makeStreamResponse({ chunks, contentLength: '10' }))
    );

    const { result } = renderHook(() => useFileDownload(HOST));
    expect(result.current.isDownloading).toBe(false);

    await act(async () => {
      await result.current.download('app.log');
    });

    expect(global.fetch).toHaveBeenCalledWith(
      `${HOST}/download-file?file=app.log`,
      expect.objectContaining({ signal: expect.any(AbortSignal) })
    );
    expect(result.current.progress).toBe(100);
    expect(result.current.isDownloading).toBe(false);
    expect(result.current.error).toBeNull();
    expect(global.URL.createObjectURL).toHaveBeenCalled();
    expect(global.URL.revokeObjectURL).toHaveBeenCalled();
  });

  test('sets error on HTTP 404', async () => {
    global.fetch = jest.fn(() =>
      Promise.resolve({ ok: false, status: 404, headers: { get: () => null }, body: null })
    );
    const { result } = renderHook(() => useFileDownload(HOST));
    await act(async () => {
      await result.current.download('missing.log');
    });
    expect(result.current.error).toMatch(/404/);
    expect(result.current.isDownloading).toBe(false);
  });

  test('sets error on network failure', async () => {
    global.fetch = jest.fn(() => Promise.reject(new Error('network error')));
    const { result } = renderHook(() => useFileDownload(HOST));
    await act(async () => {
      await result.current.download('app.log');
    });
    expect(result.current.error).toBe('network error');
    expect(result.current.isDownloading).toBe(false);
  });

  test('cancels download via AbortController without setting error', async () => {
    let abortFired = false;
    global.fetch = jest.fn((url, opts) => {
      return new Promise((_, reject) => {
        opts.signal.addEventListener('abort', () => {
          abortFired = true;
          const err = new Error('aborted');
          err.name = 'AbortError';
          reject(err);
        });
      });
    });

    const { result } = renderHook(() => useFileDownload(HOST));
    let downloadPromise;
    act(() => {
      downloadPromise = result.current.download('app.log');
    });
    await waitFor(() => expect(result.current.isDownloading).toBe(true));

    act(() => {
      result.current.cancel();
    });
    await act(async () => {
      await downloadPromise;
    });

    expect(abortFired).toBe(true);
    expect(result.current.isDownloading).toBe(false);
    expect(result.current.error).toBeNull();
  });

  test('handles missing Content-Length gracefully', async () => {
    const chunks = [new Uint8Array([1, 2, 3])];
    global.fetch = jest.fn(() =>
      Promise.resolve(makeStreamResponse({ chunks, contentLength: null }))
    );
    const { result } = renderHook(() => useFileDownload(HOST));
    await act(async () => {
      await result.current.download('app.log');
    });
    expect(result.current.progress).toBe(0);
    expect(result.current.error).toBeNull();
    expect(result.current.isDownloading).toBe(false);
  });

  test('no-op when filename is empty', async () => {
    global.fetch = jest.fn();
    const { result } = renderHook(() => useFileDownload(HOST));
    await act(async () => {
      await result.current.download('');
    });
    expect(global.fetch).not.toHaveBeenCalled();
    expect(result.current.isDownloading).toBe(false);
  });

  test('no-op when host is empty', async () => {
    global.fetch = jest.fn();
    const { result } = renderHook(() => useFileDownload(''));
    await act(async () => {
      await result.current.download('app.log');
    });
    expect(global.fetch).not.toHaveBeenCalled();
  });
});
