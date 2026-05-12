# cmd

This package holds the [urfave/cli](https://github.com/urfave/cli) sub-commands
that back the `bucky` CLI binary defined in `../main.go`.

| Command | File | Purpose |
|---|---|---|
| `bucky install` | `install.go` | Download whisper.cpp prebuilt libraries to a local `lib/` directory |
| `bucky system`  | `system.go`  | Show host + whisper.cpp system info (FFI hookups land in PR #2) |
| `bucky info`    | `info.go`    | Print the bucky banner / tagline |

The CLI structure mirrors [`hybridgroup/yzma/cmd`](https://github.com/hybridgroup/yzma/tree/main/cmd) so the two projects feel familiar side-by-side.
