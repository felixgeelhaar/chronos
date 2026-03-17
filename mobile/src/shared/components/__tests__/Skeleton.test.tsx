import React from 'react';
import { render } from '@testing-library/react-native';
import { Skeleton, SessionListSkeleton, ExerciseCardSkeleton } from '../Skeleton';

describe('Skeleton', () => {
  it('should render with default props', () => {
    const { toJSON } = render(<Skeleton />);
    expect(toJSON()).toBeTruthy();
  });

  it('should render with custom width and height', () => {
    const { getByTestId } = render(
      <Skeleton width={200} height={50} testID="custom-skeleton" />
    );
    const skeleton = getByTestId('custom-skeleton');
    expect(skeleton).toBeTruthy();
  });

  it('should render with custom border radius', () => {
    const { toJSON } = render(<Skeleton borderRadius={12} />);
    expect(toJSON()).toBeTruthy();
  });
});

describe('SessionListSkeleton', () => {
  it('should render session skeleton structure', () => {
    const { toJSON } = render(<SessionListSkeleton />);
    expect(toJSON()).toMatchSnapshot();
  });
});

describe('ExerciseCardSkeleton', () => {
  it('should render exercise skeleton structure', () => {
    const { toJSON } = render(<ExerciseCardSkeleton />);
    expect(toJSON()).toMatchSnapshot();
  });
});
