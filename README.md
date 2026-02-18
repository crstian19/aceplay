# Aceplay

A modern reimplementation of `acestream-launcher` in **Go** with an elegant visual interface using the Charm ecosystem.

## Features

- **Auto-start engine**: Automatically starts acestream-engine if not running (just like original acestream-launcher)
- **Compiled binary**: No runtime dependencies
- **Elegant visual interface**: Spinners, progress bars, and styles with Lipgloss
- **Multiple players**: Support for mpv, vlc, ffplay, and more
- **Flexible configuration**: CLI flags + configuration file
- **HLS mode**: HLS streaming support
- **Desktop notifications**: Integration with libnotify
- **Auto-completion**: Bash, zsh, fish

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
git clone https://github.com/crstian/aceplay.git
cd aceplay
make build
sudo make install
```

### Pre-compiled binaries

Download from the [releases](https://github.com/crstian/aceplay/releases) page.

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
- Cobra v1.9.1 - CLI framework
- Viper v1.20.1 - Configuration
- Lipgloss v0.12.1 - Terminal styles
- Bubbles - UI components
- Log - Structured logging
- Resty v2 - HTTP client

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
