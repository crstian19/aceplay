# Changelog

## [v0.4.6](https://github.com/crstian19/aceplay/releases/tag/v0.4.6)

[Compare to previous version](https://github.com/crstian19/aceplay/compare/v0.4.5...v0.4.6)

### Bug Fixes

- **deps**: update module charm.land/lipgloss/v2 to v2.1.0 (#32) ([908d195](https://github.com/crstian19/aceplay/commit/908d195ce142d9a28caaf79195c1a50a83558b27))

## [0.4.5] - 2026-11-08

### Security
- Updated golang.org/x/text to v0.41.0

### Changed
- Updated golangci-lint to v2.15.0 in CI
- Switched Renovate to the shared preset

## [0.4.4] - 2026-09-09

### Changed
- Updated charm.land/lipgloss/v2 to v2.0.6
- Updated charm.land/log/v2 to v2.0.1
- Updated stretchr/testify to v1.12.1
- Updated Go toolchain to 1.27 in CI workflows
- Updated actions/checkout to v7 and actions/setup-go to v7

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
