<img src="assets/michishirube-logo.png" alt="Michishirube Logo" width="100" align="left" style="margin-right: 20px; margin-top: -30px;">

# Michishirube

A desktop application for organizing links to websites, built with Go and Fyne.

## Features

- **Hierarchical Organization**: Organize your links in a tree structure of topics and subtopics
- **Search Functionality**: Quickly find links by searching for their names
- **Easy Movement**: Move links between topics with a simple drag-and-drop interface
- **Thumbnails**: Visual representation of links with thumbnails
- **Cross-Platform**: Works on macOS, Windows, and Linux

## Installation

### Prerequisites

- Go 1.18 or later
- For macOS: Xcode Command Line Tools and GLFW (`brew install glfw`)
- For Linux: Required development libraries (`apt-get install libgl1-mesa-dev xorg-dev`)
- For Windows: GCC compiler (via MinGW or TDM-GCC)

### Building from Source

1. Clone the repository:
   ```
   git clone https://github.com/jparrill/michishirube.git
   cd michishirube
   ```

2. Build the application:
   ```
   go build -o michishirube ./cmd
   ```

3. Run the application:
   ```
   ./michishirube
   ```

### Using Docker

You can also run Michishirube in Docker:

1. Build and run (auto-detects platform):
   ```
   make docker-run
   ```

2. For more Docker options:
   ```
   make help
   ```

See [DOCKER.md](DOCKER.md) for detailed instructions on running with Docker.

### Cross-Platform Builds

Build for multiple platforms at once:

```
make cross-build
```

This creates binaries for Linux and macOS (both amd64 and arm64).

## Usage

1. **Adding Topics**: Click the "+" button in the Topics section to create a new topic
2. **Adding Links**: Select a topic, then click the "+" button in the Links section to add a new link
3. **Searching**: Type in the search bar at the top to find links by name
4. **Moving Links**: Click the "move" icon on a link to move it to a different topic
5. **Opening Links**: Click the "open" icon on a link to open it in your default browser

## Development

### Project Structure

- `cmd/`: Main application entry point
- `internal/`: Internal packages
  - `data/`: Database models and operations
  - `service/`: Business logic
  - `ui/`: User interface components

### Running Tests

```
go test ./...
```

### Docker Development

For Docker-based development:

```
# Set up Docker buildx
make docker-setup

# Build multi-platform images
make docker-build

# Build for local platform only
make docker-build-local
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.
