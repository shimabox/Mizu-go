# Mizu-go

Mizu-go is a joke script that simulates water(H2o) generation in Go.

> [!NOTE]
> A Go + [Ebitengine](https://ebitengine.org/) port of [Mizu-ts](https://github.com/shimabox/Mizu-ts) (TypeScript), which itself is a port of [Mizu](https://github.com/shimabox/Mizu), originally written in JavaScript.

## Features

- **Cross-platform**: Desktop (macOS, Linux, Windows) + WebAssembly support
- **Ebitengine v2**: Built with [Ebitengine](https://ebitengine.org/), a practical 2D game engine in pure Go
- **Performance monitoring**: Optional overlay showing FPS, frame time, particle counts, and more
- **URL parameters**: Control initial particle counts and measurement mode via query strings (WebAssembly only)

## Requirements

### Desktop

- **Go 1.24** or later
- **macOS**: Xcode Command Line Tools (required for cgo)
- **Linux**: See [Ebitengine installation guide](https://ebitengine.org/en/documents/install.html)

### WebAssembly

- **Go 1.24** or later (same as desktop)

## Usage

### Desktop

Run the simulation locally:

```sh
make run
# or
go run ./cmd/mizu
```

#### Flags

| Flag | Default | Description |
|:---|:---|:---|
| `-h` | 30 | Initial H atom count (scaled by screen width) |
| `-o` | 50 | Initial O atom count (scaled by screen width) |
| `-m` | false | Enable measurement overlay (FPS/frame time/particle counts) |

**Example:**

```sh
go run ./cmd/mizu -m -h 100 -o 100
```

All particle counts are automatically scaled by the simulation's `CountScale()` based on screen width. This allows consistent behavior across different resolutions.

### WebAssembly

Build the WebAssembly binary:

```sh
make wasm
```

This generates:
- `web/mizu.wasm` — The compiled WebAssembly binary
- `web/wasm_exec.js` — The glue script that bridges Go and JavaScript

Preview locally:

```sh
make serve
# Visit http://localhost:8080 in your browser
```

#### URL Parameters

Control the simulation via query strings (same semantics as Mizu-ts):

| Parameter | Default | Type | Description |
|:---|:---|:---|:---|
| `h` | 30 | number | Initial H atom count (scaled by screen width) |
| `o` | 50 | number | Initial O atom count (scaled by screen width) |
| `m` | 0 | number | Show performance overlay if set to 1 |

**Examples:**

- `https://shimabox.github.io/Mizu-go/?m=1` — Enable measurement overlay
- `https://shimabox.github.io/Mizu-go/?h=60&o=100` — Custom particle counts
- `https://shimabox.github.io/Mizu-go/?m=1&h=500&o=500` — Load test with overlay

#### Automatic Deployment

Push to the `main` branch to automatically deploy the WebAssembly demo to GitHub Pages. The deployment workflow is configured in `.github/workflows/deploy.yml`.

## Testing

Run the full test suite:

```sh
make test
# or
go test ./...
```

The simulation logic is independent of Ebitengine, so all game logic tests run as standard Go unit tests.

## Development

### Project Structure

```
cmd/mizu/
├── main.go          - Entry point; sets up Ebitengine window and game state
├── params_js.go     - URL parameter parsing (WebAssembly only)
└── params_default.go - Flag-based parameter handling (desktop)

internal/
├── core/            - Foundational types (Particle interface, Random, Bounds)
├── behavior/        - Behavior simulation
├── physics/         - Physics simulation and collision detection
├── reaction/        - Reaction rules and particle transformations
├── particle/        - Particle factory and particle types (H, H2, O, H2o)
├── sim/             - Simulator engine and world state management
├── render/          - Rendering logic and sprite management
└── debug/           - Debugging utilities
```

### Key Design Decisions

- **core** has no external dependencies, including Ebitengine, so it can be imported safely in tests.
- **Particle interface**: Defined in `core`, used across `behavior`, `physics`, `reaction`, `render`, and `sim`.
- **Simulation loop**: Managed by `sim.Simulator`, decoupled from rendering via `sim.World`.
- **Collision detection**: Grid-based spatial partitioning in `physics` for efficient collision queries.
- **Reactions**: Registry pattern in `reaction` allows composable particle transformations.

### Continuous Integration

- **CI**: Runs on every push and pull request (see `.github/workflows/ci.yml`)
  - Tests on Ubuntu with Go 1.24 and Ebitengine dependencies
  - Runs `go build`, `go vet`, and `go test ./...`
- **Deploy**: Automatically builds and deploys WebAssembly to GitHub Pages on pushes to `main` (see `.github/workflows/deploy.yml`)

## License

[MIT](./LICENSE)
