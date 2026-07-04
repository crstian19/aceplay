# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.4.3] - 2026-07-04

### Changed
- Updated charm.land/lipgloss/v2 to v2.0.3 (fixes background color query hang)
- Updated indirect dependencies (charmbracelet/x/ansi, go-colorful, go-runewidth, golang.org/x/sys)
- Configured Renovate for automated dependency updates

### Fixed
- Resolved golangci-lint errcheck and staticcheck issues across codebase and tests

## [0.4.2] - 2026-03-20

### Fixed
- Updated colorprofile to fix "short write" errors

## [0.4.1] - 2026-03-20

### Security
- Updated AUR SSH key configuration

## [0.4.0] - 2026-03-20

### Changed
- Updated charm.land libraries to v2 (breaking changes in import paths)
  - `charm.land/huh/v2` (was `github.com/charmbracelet/huh`)
  - `charm.land/log/v2` (was `github.com/charmbracelet/log`)
  - `charm.land/lipgloss/v2` (already updated)
- Updated GitHub Actions to latest versions

## [0.3.3] - 2026-02-28

### Fixed
- AUR workflow version placeholder (use __VERSION__ for proper sed replacement)

## [0.3.2] - 2026-02-28

### Fixed
- AUR package download URL (use tarball format instead of direct binary)

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
