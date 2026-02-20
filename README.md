<p align="center">
  <img src="logo.png" alt="Aceplay" width="200">
</p>

<h1 align="center">Aceplay</h1>

<p align="center">
  <a href="https://github.com/crstian19/aceplay/releases">
    <img src="https://img.shields.io/github/v/release/crstian19/aceplay?sort=semver&style=for-the-badge" alt="Release">
  </a>
  <a href="https://aur.archlinux.org/packages/aceplay">
    <img src="https://img.shields.io/aur/version/aceplay?style=for-the-badge&color=blue" alt="AUR">
  </a>
  <a href="https://github.com/crstian19/aceplay/actions/workflows/ci.yml">
    <img src="https://img.shields.io/github/actions/workflow/status/crstian19/aceplay/ci.yml?style=for-the-badge" alt="CI">
  </a>
  <a href="LICENSE">
    <img src="https://img.shields.io/github/license/crstian19/aceplay?style=for-the-badge" alt="License">
  </a>
</p>

---

> **Fun fact**: In Spain, aceplay is also known as "Herakleopolis" — the ancient Greek name for Neni-Nesut, the eternal rival of Thebes. 🏛️

---

A modern reimplementation of `acestream-launcher` in **Go** with an elegant visual interface using the Charm ecosystem.

## Features

- **Auto-start engine**: Automatically starts acestream-engine if not running (just like original acestream-launcher)
- **Compiled binary**: No runtime dependencies
- **Elegant visual interface**: Spinners, progress bars, and styles with Lipgloss
- **Multiple players**: Support for mpv, vlc, ffplay, and more
- **Flexible configuration**: CLI flags + configuration file
- **HLS mode**: HLS streaming support
- **Desktop notifications**: Integration with libnotify
- **Browser integration**: Open `acestream://` links directly from your browser

## Installation

### From AUR (Arch Linux)

```bash
yay -S aceplay
# or
paru -S aceplay
```

### From source

Requirements:
- Go 1.22+
- make

```bash
git clone https://github.com/crstian19/aceplay.git
cd aceplay
make build
sudo make install
```

### Pre-compiled binaries

Download from the [releases](https://github.com/crstian19/aceplay/releases) page.

## Usage

```bash
# Play stream with mpv (default)
aceplay acestream://abcd1234...

# Use VLC
aceplay acestream://abcd1234... --player vlc

# HLS mode
aceplay acestream://abcd1234... --hls

# Verbose mode
aceplay acestream://abcd1234... --verbose

# Configure default player
aceplay config set player vlc

# Show configuration
aceplay config show
```

### Opening acestream:// links from your browser

After installation, register the protocol handler:

```bash
# Register acestream:// protocol
aceplay register-protocol
```

This allows you to click `acestream://` links directly in your browser and have Aceplay automatically handle them.

## Configuration

The configuration file is located at `~/.config/aceplay/config.yaml`:

```yaml
player: mpv
engine:
  host: localhost
  port: 6878
timeout: 60s
hls: false
verbose: false
```

You can also use environment variables with the prefix `ACEPLAY_`:
- `ACEPLAY_PLAYER=vlc`
- `ACEPLAY_ENGINE_HOST=192.168.1.100`
- `ACEPLAY_HLS=true`

## Development

```bash
# Run tests
make test

# Run tests with coverage
make test-coverage

# Build
make build

# Build for all platforms
make build-all

# Install locally
make install

# Clean
make clean

# View all commands
make help
```

## Architecture

```
aceplay/
├── cmd/aceplay/          # Entry point
├── internal/
│   ├── acestream/        # HTTP client API acestream-engine
│   ├── config/           # Configuration
│   ├── player/           # Video player launcher
│   ├── notify/           # Desktop notifications
│   └── ui/               # Charm UI components
├── pkg/acestream/        # acestream:// URL parser
├── scripts/              # Utility scripts
├── Makefile              # Build, test, install
└── README.md             # Documentation
```

## Dependencies

### Build
- Go 1.22+

### Runtime
- **acestream-engine** (will be auto-started if not running - same behavior as original acestream-launcher)
- Video player (mpv, vlc, or ffplay)

### Go Libraries
- Cobra v1.10.2 - CLI framework
- Viper v1.21.0 - Configuration
- Lipgloss v1.1.0 - Terminal styles
- Bubbles - UI components
- Log v0.4.2 - Structured logging
- Resty v2.17.2 - HTTP client

## License

MIT License - see [LICENSE](LICENSE) for details.

## Credits

Inspired by [acestream-launcher](https://github.com/jonian/acestream-launcher) by Jonian Guveli.

UI built with the [Charm](https://charm.sh/) ecosystem.

## Contributing

Contributions are welcome. Please:

1. Fork the repository
2. Create a branch for your feature (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Make sure to:
- Run `make check` before committing
- Add tests for new features
- Update documentation
