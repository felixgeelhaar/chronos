# Release Checklist

Use this checklist before every production release to ensure quality and completeness.

## Pre-Release

### Code Quality
- [ ] All tests passing (`npm test`)
- [ ] Test coverage >= 70%
- [ ] No linting errors (`npm run lint`)
- [ ] No TypeScript errors (`npm run type-check`)
- [ ] Code reviewed and approved
- [ ] All TODOs addressed or documented

### Functionality
- [ ] All features working as expected
- [ ] No critical bugs
- [ ] Error handling tested
- [ ] Offline mode tested
- [ ] Sync functionality verified
- [ ] Push notifications working
- [ ] Video upload/playback tested

### Performance
- [ ] App launches in < 3 seconds
- [ ] Navigation transitions smooth (60 FPS)
- [ ] No memory leaks
- [ ] Images loading efficiently
- [ ] Database queries optimized
- [ ] No ANR (Application Not Responding) errors

### Security
- [ ] API keys secured (not in version control)
- [ ] HTTPS enforced for all requests
- [ ] Sensitive data encrypted
- [ ] Authentication tokens secured
- [ ] Permission requests justified
- [ ] No hardcoded secrets

### Testing
- [ ] Tested on physical iOS device
- [ ] Tested on physical Android device
- [ ] Tested on various screen sizes
- [ ] Tested with poor network conditions
- [ ] Tested in airplane mode (offline)
- [ ] Tested push notifications
- [ ] Tested background sync

## Version Management

### Version Numbers
- [ ] Updated version in `package.json`
- [ ] Updated version in `ios/Info.plist`
- [ ] Updated version in `android/app/build.gradle`
- [ ] Version follows semantic versioning (MAJOR.MINOR.PATCH)

### Build Numbers
- [ ] Incremented iOS build number
- [ ] Incremented Android version code
- [ ] Build numbers documented in release notes

## Documentation

- [ ] README.md updated
- [ ] CHANGELOG.md updated with new features/fixes
- [ ] API documentation updated (if API changes)
- [ ] User-facing documentation updated
- [ ] Migration guide (if breaking changes)

## iOS Build

### Configuration
- [ ] Bundle identifier correct
- [ ] Display name correct
- [ ] App icon set (all sizes)
- [ ] Launch screen configured
- [ ] Provisioning profile valid
- [ ] Signing certificate valid

### Capabilities
- [ ] Push notifications enabled
- [ ] Background modes configured
- [ ] Camera/photo library permissions
- [ ] Network usage described

### Build Process
- [ ] Clean build successful
- [ ] Archive created successfully
- [ ] No warnings in build output
- [ ] App validated in Xcode
- [ ] Uploaded to App Store Connect

### App Store
- [ ] Screenshots uploaded (all device sizes)
- [ ] App preview video (if applicable)
- [ ] App description accurate
- [ ] Keywords optimized
- [ ] Privacy policy URL updated
- [ ] Support URL updated
- [ ] Marketing materials ready

## Android Build

### Configuration
- [ ] Application ID correct
- [ ] App name correct
- [ ] App icon set (all densities)
- [ ] Splash screen configured
- [ ] Signing key configured

### Permissions
- [ ] Camera permission
- [ ] Storage permission
- [ ] Internet permission
- [ ] Notification permission (Android 13+)
- [ ] Permissions justified in privacy policy

### Build Process
- [ ] Clean build successful
- [ ] Release APK/AAB generated
- [ ] APK signed with release key
- [ ] ProGuard/R8 configuration tested
- [ ] No obfuscation issues

### Play Store
- [ ] Screenshots uploaded (all device types)
- [ ] Feature graphic uploaded
- [ ] Video preview (if applicable)
- [ ] Short description (80 chars)
- [ ] Full description (4000 chars)
- [ ] Privacy policy URL updated
- [ ] Content rating completed

## Backend Compatibility

- [ ] API endpoints compatible
- [ ] Database migrations complete
- [ ] Server capacity verified
- [ ] CDN configured for media
- [ ] Monitoring and alerts set up

## Deployment

### Pre-Deployment
- [ ] Staging environment tested
- [ ] Beta testing complete
- [ ] Crash reporting configured
- [ ] Analytics configured
- [ ] Feature flags ready (if applicable)

### Deployment Process
- [ ] Deployment time scheduled
- [ ] Team notified
- [ ] Rollback plan ready
- [ ] Support team briefed
- [ ] Release notes published

### Post-Deployment
- [ ] App available in stores
- [ ] Download and open tested
- [ ] No critical crashes reported
- [ ] Analytics showing expected metrics
- [ ] User feedback monitored
- [ ] Support channels monitored

## Monitoring

### First 24 Hours
- [ ] Crash rate < 1%
- [ ] ANR rate < 0.5%
- [ ] App store rating maintained
- [ ] No spike in support tickets
- [ ] Server response times normal
- [ ] Sync success rate > 95%

### First Week
- [ ] User retention tracked
- [ ] Feature adoption monitored
- [ ] Performance metrics reviewed
- [ ] User feedback analyzed
- [ ] Bug reports triaged

## Rollback Plan

If critical issues are discovered:

1. **Immediate Actions**
   - [ ] Halt staged rollout
   - [ ] Document the issue
   - [ ] Notify team and stakeholders

2. **Assess Severity**
   - [ ] Determine user impact
   - [ ] Estimate fix timeline
   - [ ] Decide: hotfix or rollback

3. **Execute Plan**
   - [ ] If hotfix: Create emergency release
   - [ ] If rollback: Revert to previous version
   - [ ] Monitor deployment
   - [ ] Communicate with users

## Release Notes Template

```markdown
# Version X.Y.Z

Released: YYYY-MM-DD

## ✨ New Features
- Feature description with user benefit

## 🐛 Bug Fixes
- Fix description and impact

## ⚡ Improvements
- Performance/UX improvement details

## 🔧 Technical
- Technical changes for developers

## 📝 Notes
- Migration steps if required
- Breaking changes if any
```

## Post-Release Tasks

- [ ] Update internal documentation
- [ ] Share release announcement
- [ ] Plan next sprint features
- [ ] Review release process
- [ ] Document lessons learned
- [ ] Celebrate with team! 🎉

---

**Release Manager:** _________________
**Release Date:** _________________
**Version:** _________________
**Status:** ⬜ Not Started | ⬜ In Progress | ⬜ Complete
