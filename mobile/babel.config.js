module.exports = {
  presets: ['module:@react-native/babel-preset'],
  plugins: [
    [
      'module-resolver',
      {
        root: ['./src'],
        extensions: ['.ios.js', '.android.js', '.js', '.ts', '.tsx', '.json'],
        alias: {
          '@': './src',
          '@app': './src/app',
          '@features': './src/features',
          '@shared': './src/shared',
          '@services': './src/services',
          '@config': './src/config',
        },
      },
    ],
    'react-native-reanimated/plugin',
  ],
};
