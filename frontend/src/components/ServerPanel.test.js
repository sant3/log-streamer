import { render, screen, fireEvent } from '@testing-library/react';
import ServerPanel from './ServerPanel';

const mockServers = [
  { name: 'Server A', url: 'http://server-a' },
  { name: 'Server B', url: 'http://server-b' },
];

describe('ServerPanel', () => {
  test('renders nothing when only one server', () => {
    const { container } = render(
      <ServerPanel
        servers={[{ name: 'Solo', url: 'http://solo' }]}
        activeHost="http://solo"
        serverStatuses={{}}
        onServerSelect={() => {}}
        isPanelOpen={true}
        onTogglePanel={() => {}}
      />
    );
    expect(container.innerHTML).toBe('');
  });

  test('renders server list when multiple servers', () => {
    render(
      <ServerPanel
        servers={mockServers}
        activeHost="http://server-a"
        serverStatuses={{ 'Server A': 'online', 'Server B': 'offline' }}
        onServerSelect={() => {}}
        isPanelOpen={true}
        onTogglePanel={() => {}}
      />
    );
    expect(screen.getByText('Server A')).toBeInTheDocument();
    expect(screen.getByText('Server B')).toBeInTheDocument();
  });

  test('calls onServerSelect when a server is clicked', () => {
    const onSelect = jest.fn();
    render(
      <ServerPanel
        servers={mockServers}
        activeHost="http://server-a"
        serverStatuses={{}}
        onServerSelect={onSelect}
        isPanelOpen={true}
        onTogglePanel={() => {}}
      />
    );
    fireEvent.click(screen.getByText('Server B'));
    expect(onSelect).toHaveBeenCalledWith('http://server-b');
  });

  test('highlights active server', () => {
    render(
      <ServerPanel
        servers={mockServers}
        activeHost="http://server-a"
        serverStatuses={{}}
        onServerSelect={() => {}}
        isPanelOpen={true}
        onTogglePanel={() => {}}
      />
    );
    const activeItem = screen.getByText('Server A').closest('li');
    expect(activeItem).toHaveClass('active');
  });

  test('shows status dots with correct class', () => {
    const { container } = render(
      <ServerPanel
        servers={mockServers}
        activeHost="http://server-a"
        serverStatuses={{ 'Server A': 'online', 'Server B': 'offline' }}
        onServerSelect={() => {}}
        isPanelOpen={true}
        onTogglePanel={() => {}}
      />
    );
    const dots = container.querySelectorAll('.status-dot');
    expect(dots[0]).toHaveClass('online');
    expect(dots[1]).toHaveClass('offline');
  });

  test('toggle panel button calls onTogglePanel', () => {
    const onToggle = jest.fn();
    render(
      <ServerPanel
        servers={mockServers}
        activeHost="http://server-a"
        serverStatuses={{}}
        onServerSelect={() => {}}
        isPanelOpen={true}
        onTogglePanel={onToggle}
      />
    );
    fireEvent.click(screen.getByText('<'));
    expect(onToggle).toHaveBeenCalledTimes(1);
  });
});
