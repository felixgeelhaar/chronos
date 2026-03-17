#!/bin/bash

# Build iOS app for release
# Usage: ./scripts/build-ios.sh

set -e  # Exit on error

echo "🍎 Building iOS app for release..."

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

# Check if iOS directory exists
if [ ! -d "ios" ]; then
    echo -e "${RED}Error: ios directory not found.${NC}"
    exit 1
fi

# Install dependencies
echo -e "${YELLOW}📦 Installing dependencies...${NC}"
npm install

# Install pods
echo -e "${YELLOW}📱 Installing CocoaPods dependencies...${NC}"
cd ios
pod install
cd ..

# Type check
echo -e "${YELLOW}🔍 Running type check...${NC}"
npm run type-check

# Lint
echo -e "${YELLOW}🔧 Running linter...${NC}"
npm run lint

# Run tests
echo -e "${YELLOW}🧪 Running tests...${NC}"
npm test -- --coverage

# Build iOS app
echo -e "${YELLOW}🏗️  Building iOS app...${NC}"
cd ios
xcodebuild \
    -workspace AscendMobile.xcworkspace \
    -scheme AscendMobile \
    -configuration Release \
    -archivePath build/AscendMobile.xcarchive \
    archive \
    -allowProvisioningUpdates

echo -e "${GREEN}✅ iOS archive created successfully!${NC}"
echo -e "Archive location: ios/build/AscendMobile.xcarchive"

# Export IPA (optional)
read -p "Export IPA for App Store? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}📤 Exporting IPA...${NC}"
    xcodebuild \
        -exportArchive \
        -archivePath build/AscendMobile.xcarchive \
        -exportPath build/Release-iphoneos \
        -exportOptionsPlist exportOptions.plist \
        -allowProvisioningUpdates

    echo -e "${GREEN}✅ IPA exported successfully!${NC}"
    echo -e "IPA location: ios/build/Release-iphoneos/AscendMobile.ipa"
fi

cd ..

echo -e "${GREEN}🎉 iOS build complete!${NC}"
echo -e "Next steps:"
echo -e "1. Open Xcode and validate the archive"
echo -e "2. Upload to App Store Connect"
echo -e "3. Submit for review"
