# Week 12: Video Upload & Management - Setup Instructions

## New Dependencies

Week 12 adds video recording, uploading, and playback functionality. The following new dependencies are required:

### react-native-image-picker (^7.1.0)
For selecting videos from the device library or recording new videos.

### Installation Steps

```bash
# Install the new dependency
npm install

# iOS: Install pods
cd ios && pod install && cd ..

# iOS: Add camera and photo library permissions to Info.plist
# Add these keys to ios/AscendMobile/Info.plist:
<key>NSCameraUsageDescription</key>
<string>We need access to your camera to record workout videos</string>
<key>NSPhotoLibraryUsageDescription</key>
<string>We need access to your photo library to select workout videos</string>
<key>NSMicrophoneUsageDescription</key>
<string>We need access to your microphone to record workout videos</string>

# Android: Add camera and storage permissions to AndroidManifest.xml
# Add these permissions to android/app/src/main/AndroidManifest.xml:
<uses-permission android:name="android.permission.CAMERA" />
<uses-permission android:name="android.permission.READ_EXTERNAL_STORAGE" />
<uses-permission android:name="android.permission.WRITE_EXTERNAL_STORAGE" />

# Android: Update gradle settings if needed (should be automatic with react-native 0.74+)
```

## Features Implemented

### Video Service (`src/services/api/video.service.ts`)
- Upload videos with progress tracking
- Get video details
- List videos by user or set
- Delete videos
- Get video analysis results
- Trigger video reprocessing
- Stream video URLs with quality selection
- Thumbnail URLs

### Video Components

#### VideoPicker (`src/shared/components/VideoPicker.tsx`)
- Modal dialog for video selection
- Record new video with camera
- Select existing video from library
- Video duration validation (max 60s by default)
- File size validation (max 100MB)
- Processing indicators

#### VideoUpload (`src/shared/components/VideoUpload.tsx`)
- Complete upload workflow
- Progress bar with percentage
- Upload cancellation
- Success/error handling
- Optional set association
- Callback hooks for upload events

#### VideoPlayer (`src/shared/components/VideoPlayer.tsx`)
- Video playback with react-native-video
- Play/pause controls
- Progress bar
- Time display
- Quality selection support
- Error handling
- Loading states

### Session Integration

#### SessionDetailScreen Updates
- Video players embedded in set display
- Automatic video loading for sets with videos
- Proper video URL resolution
- Thumbnail support

### API Client Updates (`src/services/api/client.ts`)
- Upload progress callback support
- `onUploadProgress` event handling
- `getBaseURL()` method for video URL construction

## Usage

### Recording/Uploading Videos

Videos can be attached to sets during or after workout logging:

```typescript
import { VideoUpload } from '../../shared/components';

<VideoUpload
  setId={setId}
  onUploadComplete={(video) => {
    console.log('Video uploaded:', video.id);
  }}
  onUploadError={(error) => {
    console.error('Upload failed:', error);
  }}
  buttonText="Add Video"
  showProgress={true}
/>
```

### Playing Videos

Videos are automatically displayed in SessionDetailScreen when associated with sets:

```typescript
import { VideoPlayer } from '../../shared/components';

{set.video_id && (
  <VideoPlayer
    videoId={set.video_id}
    videoUrl={videoService.getVideoUrl(set.video_id)}
    thumbnailUrl={videoService.getThumbnailUrl(set.video_id)}
    autoPlay={false}
    controls={true}
    height={220}
  />
)}
```

## Backend Requirements

The backend API must support:
- `POST /v1/videos/upload` - multipart/form-data upload with optional set_id
- `GET /v1/videos/:id` - get video metadata
- `GET /v1/videos` - list videos (with set_id filter)
- `DELETE /v1/videos/:id` - delete video
- `GET /v1/videos/:id/stream?quality=high` - stream video file
- `GET /v1/videos/:id/thumbnail` - get video thumbnail
- `GET /v1/videos/:id/analysis` - get video analysis results
- `POST /v1/videos/:id/reprocess` - trigger reprocessing

Video processing should be asynchronous (status: pending → processing → completed/failed).

## Environment Variables

Update your `.env` file:

```bash
# Backend API URL (should already be set from Week 9)
API_BASE_URL=http://localhost:8080
```

## Testing

After setup:

1. Run the app: `npm run ios` or `npm run android`
2. Log in to your account
3. Navigate to a workout session detail screen
4. If a set has a video attached, it should appear below the set data
5. Test video playback controls
6. Test recording a new video (permission prompts should appear)
7. Test selecting from library
8. Monitor upload progress

## Troubleshooting

### iOS Issues

**Camera not working:**
- Check Info.plist permissions are properly set
- Rebuild the app after adding permissions
- Check device settings allow camera access

**Pod install fails:**
- Try `cd ios && pod deintegrate && pod install`
- Clear cache: `rm -rf ~/Library/Caches/CocoaPods`

### Android Issues

**Camera permission not requested:**
- Verify AndroidManifest.xml permissions
- Rebuild app: `cd android && ./gradlew clean && cd ..`
- Check Android system settings

**Build fails:**
- Ensure gradle is up to date
- Check `android/build.gradle` and `android/app/build.gradle` for conflicts

### Video Upload Issues

**Upload fails with 413:**
- Backend has file size limit, reduce max video size in VideoPicker
- Check server configuration (nginx/apache client_max_body_size)

**Upload stuck at 0%:**
- Check network connection
- Verify API_BASE_URL is correct
- Check backend CORS settings for multipart uploads

**Video won't play:**
- Check video URL is accessible
- Verify backend video processing completed
- Check video format is supported (MP4 H.264 recommended)

## Next Steps

Week 13 will add:
- Video analysis features (form checking, rep counting)
- Video comparison (side-by-side previous sets)
- Video trimming before upload
- Batch video management
