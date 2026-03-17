#!/bin/bash

# Build Android app for release
# Usage: ./scripts/build-android.sh

set -e  # Exit on error

echo "🤖 Building Android app for release..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if we're in the right directory
if [ ! -f "package.json" ]; then
    echo -e "${RED}Error: package.json not found. Run this script from the mobile directory.${NC}"
    exit 1
fi

# Check if Android directory exists
if [ ! -d "android" ]; then
    echo -e "${RED}Error: android directory not found.${NC}"
    exit 1
fi

# Install dependencies
echo -e "${YELLOW}📦 Installing dependencies...${NC}"
npm install

# Type check
echo -e "${YELLOW}🔍 Running type check...${NC}"
npm run type-check

# Lint
echo -e "${YELLOW}🔧 Running linter...${NC}"
npm run lint

# Run tests
echo -e "${YELLOW}🧪 Running tests...${NC}"
npm test -- --coverage

# Clean previous builds
echo -e "${YELLOW}🧹 Cleaning previous builds...${NC}"
cd android
./gradlew clean
cd ..

# Build release APK/AAB
echo -e "${YELLOW}🏗️  Building Android release...${NC}"
cd android

# Prompt for build type
echo "Select build type:"
echo "1) APK (for direct installation)"
echo "2) AAB (for Play Store)"
read -p "Enter choice (1 or 2): " BUILD_CHOICE

if [ "$BUILD_CHOICE" = "1" ]; then
    echo -e "${YELLOW}📦 Building release APK...${NC}"
    ./gradlew assembleRelease

    echo -e "${GREEN}✅ APK built successfully!${NC}"
    echo -e "APK location: android/app/build/outputs/apk/release/app-release.apk"

    # Check APK size
    APK_SIZE=$(du -h android/app/build/outputs/apk/release/app-release.apk | cut -f1)
    echo -e "APK size: ${APK_SIZE}"

elif [ "$BUILD_CHOICE" = "2" ]; then
    echo -e "${YELLOW}📦 Building release AAB...${NC}"
    ./gradlew bundleRelease

    echo -e "${GREEN}✅ AAB built successfully!${NC}"
    echo -e "AAB location: android/app/build/outputs/bundle/release/app-release.aab"

    # Check AAB size
    AAB_SIZE=$(du -h android/app/build/outputs/bundle/release/app-release.aab | cut -f1)
    echo -e "AAB size: ${AAB_SIZE}"
else
    echo -e "${RED}Invalid choice. Exiting.${NC}"
    exit 1
fi

cd ..

# Verify signing
echo -e "${YELLOW}🔐 Verifying APK/AAB signing...${NC}"
if [ "$BUILD_CHOICE" = "1" ]; then
    jarsigner -verify -verbose -certs android/app/build/outputs/apk/release/app-release.apk || {
        echo -e "${RED}Warning: APK verification failed. Make sure it's properly signed.${NC}"
    }
else
    jarsigner -verify -verbose -certs android/app/build/outputs/bundle/release/app-release.aab || {
        echo -e "${RED}Warning: AAB verification failed. Make sure it's properly signed.${NC}"
    }
fi

echo -e "${GREEN}🎉 Android build complete!${NC}"
echo -e "Next steps:"
if [ "$BUILD_CHOICE" = "1" ]; then
    echo -e "1. Test the APK on a device: adb install android/app/build/outputs/apk/release/app-release.apk"
    echo -e "2. Verify all features work correctly"
else
    echo -e "1. Upload AAB to Play Console"
    echo -e "2. Create a new release"
    echo -e "3. Submit for review"
fi
