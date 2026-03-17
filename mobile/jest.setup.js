import '@testing-library/jest-native/extend-expect';
import 'react-native-gesture-handler/jestSetup';

// Mock AsyncStorage
jest.mock('@react-native-async-storage/async-storage', () =>
  require('@react-native-async-storage/async-storage/jest/async-storage-mock')
);

// Mock React Navigation
jest.mock('@react-navigation/native', () => {
  const actualNav = jest.requireActual('@react-navigation/native');
  return {
    ...actualNav,
    useNavigation: () => ({
      navigate: jest.fn(),
      goBack: jest.fn(),
      setOptions: jest.fn(),
    }),
    useRoute: () => ({
      params: {},
    }),
    useFocusEffect: jest.fn(),
  };
});

// Mock WatermelonDB
jest.mock('@nozbe/watermelondb', () => ({
  Q: {
    where: jest.fn(),
    sortBy: jest.fn(),
    take: jest.fn(),
    gte: jest.fn(),
    lte: jest.fn(),
    desc: jest.fn(),
    asc: jest.fn(),
    oneOf: jest.fn(),
    lt: jest.fn(),
    gt: jest.fn(),
  },
}));

// Mock NetInfo
jest.mock('@react-native-community/netinfo', () => ({
  fetch: jest.fn(() =>
    Promise.resolve({
      isConnected: true,
      isInternetReachable: true,
      type: 'wifi',
    })
  ),
  addEventListener: jest.fn(() => jest.fn()),
}));

// Mock Firebase
jest.mock('@react-native-firebase/app', () => ({
  firebase: {
    app: jest.fn(),
  },
}));

jest.mock('@react-native-firebase/messaging', () => ({
  __esModule: true,
  default: jest.fn(() => ({
    requestPermission: jest.fn(() => Promise.resolve(1)),
    getToken: jest.fn(() => Promise.resolve('mock-token')),
    onMessage: jest.fn(),
    onTokenRefresh: jest.fn(),
    setBackgroundMessageHandler: jest.fn(),
    getInitialNotification: jest.fn(() => Promise.resolve(null)),
  })),
}));

// Mock Notifee
jest.mock('@notifee/react-native', () => ({
  displayNotification: jest.fn(),
  createChannel: jest.fn(),
  requestPermission: jest.fn(() => Promise.resolve({ authorizationStatus: 1 })),
  onBackgroundEvent: jest.fn(),
  onForegroundEvent: jest.fn(),
  createTriggerNotification: jest.fn(),
  cancelNotification: jest.fn(),
  cancelAllNotifications: jest.fn(),
  getBadgeCount: jest.fn(() => Promise.resolve(0)),
  setBadgeCount: jest.fn(),
}));

// Mock Background Fetch
jest.mock('react-native-background-fetch', () => ({
  configure: jest.fn((config, success, failure) => {
    success('mock-task-id');
    return Promise.resolve(0);
  }),
  scheduleTask: jest.fn(() => Promise.resolve()),
  finish: jest.fn(),
  status: jest.fn(() => Promise.resolve(0)),
  stop: jest.fn(),
  start: jest.fn(),
  registerHeadlessTask: jest.fn(),
}));

// Mock React Native Video
jest.mock('react-native-video', () => 'Video');

// Mock Image Picker
jest.mock('react-native-image-picker', () => ({
  launchCamera: jest.fn(),
  launchImageLibrary: jest.fn(),
}));

// Mock Fast Image
jest.mock('react-native-fast-image', () => ({
  __esModule: true,
  default: jest.fn(),
  priority: {
    low: 'low',
    normal: 'normal',
    high: 'high',
  },
  resizeMode: {
    contain: 'contain',
    cover: 'cover',
    stretch: 'stretch',
    center: 'center',
  },
  cacheControl: {
    immutable: 'immutable',
    web: 'web',
    cacheOnly: 'cacheOnly',
  },
  preload: jest.fn(),
  clearMemoryCache: jest.fn(() => Promise.resolve()),
  clearDiskCache: jest.fn(() => Promise.resolve()),
}));

// Mock Haptic Feedback
jest.mock('react-native-haptic-feedback', () => ({
  trigger: jest.fn(),
}));

// Mock Share
jest.mock('react-native-share', () => ({
  open: jest.fn(() => Promise.resolve()),
}));

// Mock RNFS
jest.mock('react-native-fs', () => ({
  DocumentDirectoryPath: '/mock/documents',
  writeFile: jest.fn(() => Promise.resolve()),
  readFile: jest.fn(() => Promise.resolve('')),
  unlink: jest.fn(() => Promise.resolve()),
  exists: jest.fn(() => Promise.resolve(true)),
}));

// Mock Axios
jest.mock('axios');

// Silence console warnings during tests
global.console = {
  ...console,
  warn: jest.fn(),
  error: jest.fn(),
};

// Mock timers
jest.useFakeTimers();
