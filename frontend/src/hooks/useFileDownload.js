import { useState, useCallback, useRef } from 'react';

const useFileDownload = (host) => {
  const [isDownloading, setIsDownloading] = useState(false);
  const [progress, setProgress] = useState(0);
  const [error, setError] = useState(null);
  const abortControllerRef = useRef(null);

  const download = useCallback(async (filename) => {
    if (!filename || !host) return;

    setIsDownloading(true);
    setProgress(0);
    setError(null);

    const controller = new AbortController();
    abortControllerRef.current = controller;

    try {
      const url = `${host}/download-file?file=${encodeURIComponent(filename)}`;
      const response = await fetch(url, { signal: controller.signal });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
      }

      const totalHeader = response.headers.get('Content-Length');
      const total = totalHeader ? parseInt(totalHeader, 10) : 0;
      const reader = response.body.getReader();
      const chunks = [];
      let received = 0;

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;
        chunks.push(value);
        received += value.length;
        if (total > 0) {
          setProgress(Math.round((received / total) * 100));
        }
      }

      const blob = new Blob(chunks, { type: 'application/octet-stream' });
      const downloadUrl = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = downloadUrl;
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(downloadUrl);
    } catch (err) {
      if (err.name !== 'AbortError') {
        setError(err.message);
      }
    } finally {
      setIsDownloading(false);
      abortControllerRef.current = null;
    }
  }, [host]);

  const cancel = useCallback(() => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }
  }, []);

  return { download, cancel, isDownloading, progress, error };
};

export default useFileDownload;
