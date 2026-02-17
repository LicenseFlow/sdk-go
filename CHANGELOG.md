# Changelog

All notable changes to the LicenseFlow Go SDK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres on [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.3.0] - 2026-02-17

### Added
- Environment scoping support: `environmentID` parameter added to all license operation methods
- Cache isolation between environments to prevent cross-environment cache collisions
- Optional environment parameter in `Activate()`, `Verify()`, and `Deactivate()` methods

### Changed
- **Method Signatures Updated**:
  - `Activate(licenseKey, deviceName, environmentID string)` - added `environmentID` parameter
  - `Verify(licenseKey, environmentID string)` - added `environmentID` parameter
  - `Deactivate(licenseKey, environmentID string)` - added `environmentID` parameter
- Updated cache key generation to include environment context
- Cache key format now: `"verify:{licenseKey}:{deviceID}:{environmentID}"`
- Defaults to `"default"` environment when `environmentID` is empty string

### Technical Details
- **Breaking Change**: Method signatures changed (new parameter added)
- **Migration**: Update method calls to include `environmentID` parameter (use `""` for no environment)
- **Backward Compatibility**: Pass empty string `""` for `environmentID` to maintain previous behavior

## [v0.2.0] - Previous Release

Initial release with core license management features.
