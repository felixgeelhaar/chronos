import { renderHook, act } from '@testing-library/react-hooks';
import { useDebounce } from '../useDebounce';

describe('useDebounce', () => {
  beforeEach(() => {
    jest.clearAllTimers();
  });

  it('should return initial value immediately', () => {
    const { result } = renderHook(() => useDebounce('test', 300));
    expect(result.current).toBe('test');
  });

  it('should debounce value changes', () => {
    const { result, rerender } = renderHook(
      ({ value, delay }) => useDebounce(value, delay),
      {
        initialProps: { value: 'initial', delay: 300 },
      }
    );

    expect(result.current).toBe('initial');

    // Update value
    rerender({ value: 'updated', delay: 300 });

    // Value should not change immediately
    expect(result.current).toBe('initial');

    // Fast-forward time
    act(() => {
      jest.advanceTimersByTime(300);
    });

    // Value should now be updated
    expect(result.current).toBe('updated');
  });

  it('should cancel previous timeout on rapid changes', () => {
    const { result, rerender } = renderHook(
      ({ value }) => useDebounce(value, 300),
      {
        initialProps: { value: 'initial' },
      }
    );

    // Rapidly change values
    rerender({ value: 'first' });
    act(() => {
      jest.advanceTimersByTime(100);
    });

    rerender({ value: 'second' });
    act(() => {
      jest.advanceTimersByTime(100);
    });

    rerender({ value: 'final' });

    // Advance time to complete debounce
    act(() => {
      jest.advanceTimersByTime(300);
    });

    // Should only have the final value
    expect(result.current).toBe('final');
  });

  it('should respect custom delay', () => {
    const { result, rerender } = renderHook(
      ({ value }) => useDebounce(value, 500),
      {
        initialProps: { value: 'initial' },
      }
    );

    rerender({ value: 'updated' });

    // Value should not change after 300ms
    act(() => {
      jest.advanceTimersByTime(300);
    });
    expect(result.current).toBe('initial');

    // Value should change after 500ms total
    act(() => {
      jest.advanceTimersByTime(200);
    });
    expect(result.current).toBe('updated');
  });
});
