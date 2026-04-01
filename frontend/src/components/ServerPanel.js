import React from 'react';

const ServerPanel = ({ servers, activeHost, serverStatuses, onServerSelect, isPanelOpen, onTogglePanel }) => {
  if (servers.length <= 1) return null;

  return (
    <>
      <div className="side-panel">
        <h3>Servers</h3>
        <ul className="server-list">
          {servers.map(server => (
            <li
              key={server.name}
              onClick={() => onServerSelect(server.url)}
              className={activeHost === server.url ? 'active' : ''}
            >
              <span className={`status-dot ${serverStatuses[server.name] || 'offline'}`}></span>
              {server.name}
            </li>
          ))}
        </ul>
      </div>
      <button onClick={onTogglePanel} className="panel-toggle">
        {isPanelOpen ? '<' : '>'}
      </button>
    </>
  );
};

export default ServerPanel;
