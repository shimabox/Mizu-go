# Mizu-go

Mizu-go is a joke script that simulates water(H2o) generation in Go.

![Mizu-go](https://github.com/shimabox/Mizu-go/blob/main/images/demo.gif)

> [!NOTE]
> A Go + [Ebitengine](https://ebitengine.org/) port of [Mizu-ts](https://github.com/shimabox/Mizu-ts) (TypeScript), which itself is a port of [Mizu](https://github.com/shimabox/Mizu), originally written in JavaScript.

## Demo

https://shimabox.github.io/Mizu-go/

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

`make serve` builds the WebAssembly binary first (it depends on the `wasm` target) and then hosts `web/` at `:8080` using a small Go static file server (`tools/serve.go`), so no Python or other external tools are required.

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

### Benchmark Tool

`make bench` (or `go run ./cmd/bench`) automates the manual load-testing protocol above: it opens a real Ebitengine window per scenario, measures wall-clock frame intervals and `Simulator.Update()` time, and writes a Markdown report to `bench-reports/`. Pass `-compare <git-ref>` to A/B compare the current working tree against another ref (checked out into a temporary `git worktree`).

Requires a real, focused GUI environment — `ebiten.RunGame` opens an actual OS window, so this does not work over SSH or on displayless CI machines.

```sh
make bench
go run ./cmd/bench -compare main
go run ./cmd/bench -scenarios default,500 -frames 60 -warmup 1000
```

| option | default | description |
|:---|:---|:---|
| `-compare <git-ref>` | (none) | A/B compare against the given ref, measured in the same session |
| `-scenarios <a,b,c>` | `default,500,1000` | Scenarios to run (choices: `default`, `500`, `1000`, `3000`) |
| `-frames <N>` | `300` (`60` for the `3000` scenario) | Number of frames to sample per scenario |
| `-warmup <ms>` | `3000` | Warmup time before sampling starts |
| `-out <path>` | `bench-reports/report-<YYYYMMDD-HHmmss>.md` | Report output path |

`ebiten.RunGame` can only be called once per process, so `cmd/bench` re-executes itself as a subprocess per scenario (`-run-one <scenario> -json <path>`, an internal mode) and aggregates the JSON results into the final report.

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

cmd/bench/           - Benchmark tool CLI (see "Benchmark Tool" above)

cmd/devshot/         - Dev-only visual verification tool (see "Visual Verification" below)

tools/
└── serve.go         - Dev-only static file server for `make serve` (go run only, build-ignored)

internal/
├── core/            - Foundational types (Particle interface, Random, Bounds)
├── behavior/        - Behavior simulation
├── physics/         - Physics simulation and collision detection
├── reaction/        - Reaction rules and particle transformations
├── particle/        - Particle factory and particle types (H, H2, O, H2o)
├── sim/             - Simulator engine and world state management
├── render/          - Rendering logic and sprite management
├── debug/           - Debugging utilities
└── bench/           - Ebiten-independent stats/report generation for cmd/bench
```

### Visual Verification (devshot)

Ebitengine has no headless mode, so `cmd/devshot` is the tool for capturing what the game actually renders: it runs the same wiring as `cmd/mizu`, draws every frame to an offscreen image, dumps a PNG at the given tick, and exits. Useful for before/after comparisons of rendering changes (see [#1](https://github.com/shimabox/Mizu-go/pull/1)) and for checking overlay numbers under load — no screen-recording permission required.

```sh
go run ./cmd/devshot -out shot.png                  # default scene, captured after 300 ticks (~5s)
go run ./cmd/devshot -out shot.png -m -h 500 -o 500 # load test with the stats overlay
go run ./cmd/devshot -out shot.png -ticks 900       # capture a later moment (~15s)
```

| Flag | Default | Description |
|:---|:---|:---|
| `-out` | `devshot.png` | Output PNG path |
| `-h` / `-o` | 30 / 50 | Initial particle counts (before count scaling) |
| `-m` | false | Draw the measurement overlay into the capture |
| `-ticks` | 300 | Tick at which the frame is captured (60 ticks ≈ 1s) |
| `-width` / `-height` | 1280 / 720 | Logical resolution (CSS px) |

Like `make bench`, it opens a real window, so it does not work over SSH or on displayless CI machines.

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
