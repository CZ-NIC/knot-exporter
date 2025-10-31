# Knot DNS Exporter

A Prometheus exporter for Knot DNS server that provides comprehensive metrics
collection including global statistics, zone statistics, zone status, and memory
usage.

This project is a Go port of the original Python-based Knot DNS exporter found
in the [Knot DNS project](https://gitlab.nic.cz/knot/knot-dns/-/tree/master/python/knot_exporter).

The Go implementation leverages the same C libraries and control interface from
the [Knot DNS project](https://gitlab.nic.cz/knot/knot-dns) while providing
improved performance and easier deployment through static binary compilation.

## Features

- **Global Statistics**: Knot DNS server-wide metrics
- **Zone Statistics**: Per-zone metrics including query counts and response codes
- **Zone Serials**: Zone serial numbers
- **Zone Timers**: SOA record timers (refresh, retry, expiration)
- **Memory Usage**: Process memory consumption monitoring
- **Build Information**: Version and build metadata

## Architecture

The project is organized into clean, separated packages:

- `main` package: Prometheus exporter logic and HTTP server
- `libknot` package: Clean Go wrapper around libknot C interface

## Requirements

- Go 1.23 or later
- libknot development files
- pkg-config
- Knot DNS server running with control socket enabled

### Installing Dependencies

**Ubuntu/Debian:**
```bash
sudo apt-get install libknot-dev pkg-config
```

**CentOS/RHEL/Fedora:**
```bash
sudo yum install knot-devel pkg-config
# or
sudo dnf install knot-devel pkg-config
```

## Binary releases

Binary releases on GitHub are versioned in sync with appropriate Knot DNS
release. Cross compatibility between mismatched versions of exporter and Knot
daemon isn't guaranteed.

## Building

```bash
# Check dependencies
make check-deps

# Build the application
make build

# Build with debug information
make dev

# Build with race detector
make build-race
```

## Usage

Basic usage:
```bash
./knot-exporter
```

Common options:
```bash
./knot-exporter \
  -web-listen-addr 0.0.0.0 \
  -web-listen-port 9433 \
  -knot-socket-path /run/knot/knot.sock \
  -knot-socket-timeout 5000 \
  -zone-timers \
  -debug
```

### Command Line Options

- `-web-listen-addr`: Address to listen on (default: 127.0.0.1)
- `-web-listen-port`: Port to listen on (default: 9433)
- `-knot-socket-path`: Path to Knot control socket (default: /run/knot/knot.sock)
- `-knot-socket-timeout`: Socket timeout in milliseconds (default: 2000)
- `-no-meminfo`: Disable memory usage collection
- `-no-global-stats`: Disable global statistics collection
- `-no-zone-stats`: Disable zone statistics collection
- `-no-zone-status`: Disable zone status collection
- `-no-zone-serial`: Disable zone serial collection
- `-zone-timers`: Enable SOA timer collection
- `-debug`: Enable debug logging
- `-version`: Show version information

## Metrics

Each metric comes in two variants, one as the prometheus gauge type and the
other, denoted with a `_total` suffix, as the counter type. This is useful for
the detection of overflows, service restarts or other such events.

### Global Metrics

- `knot_global_stats_*`: Dynamic global statistics from Knot DNS
- `knot_build_info`: Build and version information
- `knot_memory_usage_bytes`: Memory usage by process ID

### Zone Metrics

- `knot_zone_status`: Zone status (master/slave)
- `knot_zone_serial`: Zone serial numbers
- `knot_zone_stats_*`: Dynamic per-zone statistics
- `knot_zone_refresh_seconds`: SOA refresh timer
- `knot_zone_retry_seconds`: SOA retry timer
- `knot_zone_expiration_seconds`: SOA expiration timer
- `knot_zone_status_refresh_seconds`: Zone status refresh timer
- `knot_zone_status_expiration_seconds`: Zone status expiration timer

## Configuration

### Knot DNS Configuration

Ensure your Knot DNS server has the control interface enabled in `/etc/knot/knot.conf`:

```yaml
control:
    listen: /run/knot/knot.sock
```

### Systemd Service

Create `/etc/systemd/system/knot-exporter.service`:

```ini
[Unit]
Description=Knot DNS Prometheus Exporter
After=network.target knot.service
Requires=knot.service

[Service]
Type=simple
User=knot
Group=knot
ExecStart=/usr/local/bin/knot-exporter \
    -web-listen-addr 0.0.0.0 \
    -web-listen-port 9433 \
    -zone-timers \
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl enable knot-exporter
sudo systemctl start knot-exporter
```

### Prometheus Configuration

Add to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'knot-dns'
    static_configs:
      - targets: ['localhost:9433']
    scrape_interval: 30s
    scrape_timeout: 10s
```

Optionally if you'd like to disable the counter metrics:

```yaml
metric_relabel_configs:
  - source_labels: [__name__]
    regex: '.*_total$'
    action: drop
```

## Development

### Project Structure

```
knot-exporter/
├── go.mod                 # Go module definition
├── Makefile               # Build configuration
├── main.go                # Main application logic
├── libknot/               # Libknot C interface package
│   └── libknot.go         # Libknot bindings
└── README.md              # This file
```

### Testing

```bash
# Run tests
make test

# Run tests with race detector
make test-race

# Format and lint code
make fmt
make vet
```

## License

This project is licensed under the GNU General Public License v3.0 or later
(GPL-3.0-or-later).

### License Compatibility

- **Knot DNS**: Licensed under GPL-2.0-or-later
- **This project**: Licensed under GPL-3.0-or-later (compatible)
- **libknot C libraries**: GPL-2.0-or-later (linked as permitted by GPL)

## Contributing

We welcome contributions to improve this Knot DNS exporter! This project
maintains the spirit of the original Python implementation while leveraging Go's
advantages, chief among them the standalone nature of resulting binaries.

### How to Contribute

1. **Fork the repository** and create your feature branch from `main`
2. **Make your changes** following the existing code style
3. **Add tests** if you're adding new functionality
4. **Update documentation** if needed (README, comments, etc.)
5. **Test your changes** thoroughly:
   ```bash
   make test
   make build
   ./knot-exporter -version
   ```
6. **Submit a pull request** with a clear description of your changes

### Development Guidelines

- **Code Style**: Follow standard Go conventions (`gofmt`, `go vet`)
- **Testing**: Add tests for new features and bug fixes
- **Documentation**: Update README and code comments as needed
- **Compatibility**: Maintain compatibility with supported Knot DNS versions
- **Dependencies**: Keep external dependencies minimal and well-justified

### Types of Contributions Welcome

- **Bug fixes** and performance improvements
- **New metrics** or collection methods
- **Documentation** improvements and examples
- **Testing** enhancements and edge case coverage
- **Build system** improvements and platform support
- **Configuration** options and usability features

### Building and Testing

```bash
# Check dependencies
make check-deps

# Build
make build

# Run tests
make test

# Run with debug output
./knot-exporter -debug

# Test specific functionality
curl http://localhost:9433/metrics
curl http://localhost:9433/health
```

### Reporting Issues

When reporting issues, please include:

- **Version information**: `./knot-exporter -version`
- **Environment details**: OS, Knot DNS version, system configuration
- **Error logs**: Include relevant log output with `-debug` flag
- **Reproduction steps**: Clear steps to reproduce the issue
- **Expected vs actual behavior**

### Code of Conduct

This project follows the same collaborative spirit as the upstream Knot DNS
project. Please:

- Be respectful and constructive in discussions
- Focus on technical merit and project improvement
- Help newcomers and share knowledge
- Maintain compatibility with the GPL-3.0-or-later license

### Upstream Coordination

We aim to stay compatible with the original Python implementation and coordinate
with the upstream Knot DNS project when appropriate. Consider contributing
improvements back to the original Python exporter when applicable.

For questions or discussions, please use the project's issue tracker or reach
out to the maintainers.
