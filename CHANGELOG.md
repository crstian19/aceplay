# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.0] - 2026-02-26

### Added
- **Interactive TUI Configuration**: New `aceplay config` command with full interactive menu
  - Select video player (mpv, vlc, ffplay)
  - Configure engine host/port
  - Set timeout
  - Toggle HLS and verbose modes
- **Stream Status Display**: Shows download/upload speed and peers while playing
- **GoReleaser Configuration**: Auto-generates .deb and .rpm packages
- **Logo Display**: Shows logo on `aceplay version` and `aceplay config`

### Changed
- Improved terminal UX with Charm ecosystem
- Configuration menu now shows current settings

### Removed
- Windows and macOS support (Linux only)

## [0.2.0] - 2026-02-26

### Changed
- Linux only version

## [0.1.x] - 2026-02-18

### Added
- Initial release
- Play Ace Stream links
- Auto-start acestream-engine
- Multiple player support (mpv, vlc, ffplay)
- Desktop notifications
- Browser protocol handler registration
