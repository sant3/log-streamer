import { render } from '@testing-library/react';
import Loader from './Loader';

describe('Loader', () => {
  test('renders loader overlay', () => {
    const { container } = render(<Loader />);
    const overlay = container.querySelector('.loader-overlay');
    expect(overlay).toBeInTheDocument();
  });

  test('renders spinner inside overlay', () => {
    const { container } = render(<Loader />);
    const spinner = container.querySelector('.loader-spinner');
    expect(spinner).toBeInTheDocument();
  });
});
