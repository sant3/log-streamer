import { render, screen, fireEvent } from '@testing-library/react';
import ThemeSwitcher from './ThemeSwitcher';

describe('ThemeSwitcher', () => {
  test('renders a button', () => {
    const toggleTheme = jest.fn();
    render(<ThemeSwitcher theme="dark" toggleTheme={toggleTheme} />);
    const button = document.querySelector('.theme-switcher-button');
    expect(button).toBeInTheDocument();
  });

  test('renders moon icon in dark mode', () => {
    render(<ThemeSwitcher theme="dark" toggleTheme={() => {}} />);
    const svg = document.querySelector('.theme-switcher-button svg');
    expect(svg).toBeInTheDocument();
    // Moon icon path contains the specific d attribute
    const path = svg.querySelector('path');
    expect(path.getAttribute('d')).toContain('21 12.79');
  });

  test('renders sun icon in light mode', () => {
    render(<ThemeSwitcher theme="light" toggleTheme={() => {}} />);
    const svg = document.querySelector('.theme-switcher-button svg');
    expect(svg).toBeInTheDocument();
    const path = svg.querySelector('path');
    expect(path.getAttribute('d')).toContain('M8 11a3 3 0');
  });

  test('calls toggleTheme when clicked', () => {
    const toggleTheme = jest.fn();
    render(<ThemeSwitcher theme="dark" toggleTheme={toggleTheme} />);
    const button = document.querySelector('.theme-switcher-button');
    fireEvent.click(button);
    expect(toggleTheme).toHaveBeenCalledTimes(1);
  });
});
