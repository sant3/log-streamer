import { useState, useEffect, useCallback } from 'react';

const POLLING_ALIVE_MS = 10000;

const useServerStatus = (servers) => {
  const [serverStatuses, setServerStatuses] = useState({});

  const updateServerStatuses = useCallback(async () => {
    if (servers.length === 0) return;

    const statuses = {};
    for (const server of servers) {
      try {
        const response = await fetch(`${server.url}/alive`, { signal: AbortSignal.timeout(POLLING_ALIVE_MS) });
        statuses[server.name] = response.ok ? 'online' : 'offline';
      } catch (error) {
        statuses[server.name] = 'offline';
      }
    }
    setServerStatuses(statuses);
  }, [servers]);

  useEffect(() => {
    updateServerStatuses();
    const intervalId = setInterval(updateServerStatuses, 15000);
    return () => clearInterval(intervalId);
  }, [updateServerStatuses]);

  return serverStatuses;
};

export default useServerStatus;
