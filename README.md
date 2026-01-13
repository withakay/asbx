# sbx & asbx

This repository contains two macOS sandbox tools:

- **sbx** - A low-level CLI tool for running commands with macOS sandbox-exec policies using a flag-based interface.
- **asbx** - A higher-level CLI tool for running AI coding agents in a secure sandbox with pre-configured profiles.

Both tools are heavily inspired by [littledivy](https://github.com/littledivy)'s [sh-deno](https://github.com/littledivy/sh-deno).

## Important Notes

- **These tools use the deprecated `sandbox-exec` feature.**
- **These tools are experimental and unstable.**
- **macOS only** (darwin/amd64 and darwin/arm64)

---

# asbx - Agent Sandbox

`asbx` runs AI coding agents in a secure macOS sandbox with pre-configured profiles that restrict file system access to protect your system while allowing the agent to work effectively.

## Features

- Pre-configured profiles for popular AI coding agents (opencode, claude, codex)
- Restricts file operations to the project directory and essential paths
- Allows full outbound network for API calls
- Supports additional path customization via flags
- Print profile option for inspection and debugging

## Installation

```bash
go install github.com/syumai/sbx/cmd/asbx@latest
```

Or download from [releases](https://github.com/syumai/sbx/releases).

## Usage

```
asbx --agent-harness=<agent> [flags] <command> [command-args...]
asbx --agent-harness=<agent> [flags] -- <command> [command-flags] [command-args...]
```

### Supported Agents

| Agent | Description |
|-------|-------------|
| `opencode` | [SST OpenCode](https://opencode.ai) - Open source AI coding agent |
| `claude` | Anthropic Claude Code |
| `codex` | OpenAI Codex CLI |

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--agent-harness` | `-a` | Agent harness type (required): opencode, claude, codex |
| `--project-path` | `-p` | Project directory path (defaults to current directory) |
| `--print-profile` | `-P` | Print the generated sandbox profile and exit |
| `--extra-read` | | Additional read-only paths (can be specified multiple times) |
| `--extra-write` | | Additional write paths (can be specified multiple times) |
| `--extra-exec` | | Additional executable paths (can be specified multiple times) |

### Examples

Run opencode in sandbox for current directory:
```bash
asbx --agent-harness=opencode opencode
```

Run opencode for a specific project:
```bash
asbx --agent-harness=opencode --project-path=/path/to/project opencode
```

Run with extra paths:
```bash
asbx --agent-harness=opencode --extra-read=/custom/config -- opencode
```

Print the generated sandbox profile:
```bash
asbx --agent-harness=opencode --print-profile
```

## Agent Profiles

### OpenCode Profile

The opencode profile is designed for [OpenCode](https://opencode.ai)'s client/server architecture:

**Read Access:**
- System binaries: `/usr`, `/bin`, `/sbin`, `/System`, `/Library`, `/opt`
- Homebrew: `/opt/homebrew`, `/usr/local/Homebrew`
- OpenCode config: `~/.opencode.json`, `~/.config/opencode/`
- Git config: `~/.gitconfig`, `~/.config/git/`
- SSH (limited): `~/.ssh/known_hosts`, `~/.ssh/config`
- Shell configs: `~/.zshrc`, `~/.bashrc`, etc.
- Version managers: `~/.nvm`, `~/.fnm`, `~/.pyenv`, `~/.cargo`, `~/.rustup`, etc.

**Write Access:**
- Project directory (full read/write)
- Temp directories: `/tmp`, `/var/tmp`, `/var/folders`
- Cache: `~/.cache`
- OpenCode data: `~/.local/share/opencode/`

**Network:**
- Full outbound network (for API calls)
- Inbound network (for server mode - `opencode serve`)

### Claude Code Profile

The Claude Code profile is designed for [Anthropic's Claude Code](https://code.claude.com):

**Configuration Paths** (based on [Claude Code settings docs](https://code.claude.com/docs/en/settings)):
- `~/.claude.json` - Main global config (OAuth, MCP servers, preferences, per-project state)
- `~/.claude/settings.json` - User-specific global settings
- `~/.claude/settings.local.json` - User-specific local settings
- `.claude/settings.json` - Project settings (checked into git)
- `.claude/settings.local.json` - Project-specific local settings
- `.mcp.json` - Project MCP server configuration

**Read Access:**
- System binaries: `/usr`, `/bin`, `/sbin`, `/System`, `/Library`, `/opt`
- Homebrew: `/opt/homebrew`, `/usr/local/Homebrew`
- Git config: `~/.gitconfig`, `~/.config/git/`
- SSH (limited): `~/.ssh/known_hosts`, `~/.ssh/config`
- Shell configs: `~/.zshrc`, `~/.bashrc`, etc.
- Version managers: `~/.nvm`, `~/.fnm`, `~/.pyenv`, `~/.cargo`, `~/.rustup`, etc.
- Editor configs: `~/.vscode`, `~/.cursor`

**Write Access:**
- Project directory (full read/write)
- Temp directories: `/tmp`, `/var/tmp`, `/var/folders`
- Cache: `~/.cache`, `~/.npm`
- Claude config: `~/.claude/`, `~/.claude.json`

**Network:**
- Full outbound network (for Anthropic API calls)
- No inbound network

### Codex CLI Profile

The Codex profile is designed for [OpenAI's Codex CLI](https://developers.openai.com/codex/cli/):

**Configuration Paths** (based on [Codex config reference](https://developers.openai.com/codex/config-reference/)):
- `~/.codex/` - CODEX_HOME directory (main config home)
- `~/.codex/config.toml` - Main configuration (model, provider, approval policies, MCP)
- `~/.codex/AGENTS.md` - Global custom instructions
- `~/.codex/skills/**/SKILL.md` - Skills definitions
- `~/.codex/rules/` - Execution policy rules
- `.codex/` - Project-specific configuration layers
- `AGENTS.md` - Project-specific instructions

**Read Access:**
- System binaries: `/usr`, `/bin`, `/sbin`, `/System`, `/Library`, `/opt`
- Homebrew: `/opt/homebrew`, `/usr/local/Homebrew`
- Git config: `~/.gitconfig`, `~/.config/git/`
- SSH (limited): `~/.ssh/known_hosts`, `~/.ssh/config`
- Shell configs: `~/.zshrc`, `~/.bashrc`, etc.
- Version managers: `~/.nvm`, `~/.fnm`, `~/.pyenv`, `~/.cargo`, `~/.rustup`, etc.

**Write Access:**
- Project directory (full read/write)
- Temp directories: `/tmp`, `/var/tmp`, `/var/folders`
- Cache: `~/.cache`, `~/.npm`
- Codex home: `~/.codex/` (or `$CODEX_HOME`)

**Network:**
- Full outbound network (for OpenAI API calls)
- No inbound network

## Security Model

The sandbox uses macOS's native `sandbox-exec` (Seatbelt) to enforce:

1. **File System Isolation** - Agents can only read/write to explicitly allowed paths
2. **Network Control** - Outbound network is allowed; inbound can be enabled for server mode
3. **Process Execution** - Only binaries in approved paths can be executed

### Blocked by Default

- `~/Documents`, `~/Desktop`, `~/Downloads`
- `~/.ssh` (except known_hosts and config)
- `~/.aws`, `~/.gnupg`, `~/.kube`
- Any paths not explicitly allowed

---

# sbx - Low-level Sandbox

`sbx` is a low-level CLI tool for running commands with custom macOS sandbox-exec policies.

## Features

- Easy allow/deny configuration for common operations (file, network, process, etc.)
- Supports both relative and absolute path filtering.

## Notes

- This command implicitly applies `Common system sandbox rules` which is defined in "system.sb".
- This command also allows access to dylibs because it's required for process exec.
- When you specify `-network` flag, this command allows access to unix-socket.

## Installation

```bash
go install github.com/syumai/sbx/cmd/sbx@latest
```

Or download from [releases](https://github.com/syumai/sbx/releases).

## Usage

```
sbx [flags] <command> [command-args...]
sbx [flags] -- <command> [command-flags] [command-args...]
```

### Flags

- By default, `sbx` denies all operations.
  - To allow all operations for investigation purposes, use the `--allow-all` flag.
- You can allow operations by specifying the corresponding flags.
- You can deny operations by specifying the corresponding flags with `deny-` prefix.
- `-all` flags are boolean flags that allow / deny all operations for the corresponding operation type.
  - Other flags require path arguments as comma-separated values.
- `-network` flags support only settings below:
  - Only `ip` protocol.
  - Only `localhost` or `*` for host.

#### Special Flag

- `--allow-all`               Allow all operations (without this flag, deny all operations by default)

#### File Operations
- `--allow-file`              Allow file operations
- `--deny-file`               Deny file operations
- `--allow-file-all`          Allow all file operations
- `--deny-file-all`           Deny all file operations

##### File Read Operations
- `--allow-file-read`         Allow file read operations
- `--deny-file-read`          Deny file read operations
- `--allow-file-read-all`     Allow all file read operations
- `--deny-file-read-all`      Deny all file read operations

##### File Write Operations
- `--allow-file-write`        Allow file write operations
- `--deny-file-write`         Deny file write operations
- `--allow-file-write-all`    Allow all file write operations
- `--deny-file-write-all`     Deny all file write operations

#### Network Operations
- `--allow-network-all`       Allow all network operations
- `--deny-network-all`        Deny all network operations
- `--allow-network-inbound`   Allow inbound network operations
- `--deny-network-inbound`    Deny inbound network operations
- `--allow-network-outbound`  Allow outbound network operations
- `--deny-network-outbound`   Deny outbound network operations

#### Process Operations
- `--allow-process-exec`      Allow process execution
- `--deny-process-exec`       Deny process execution
- `--allow-process-exec-all`  Allow all process execution
- `--deny-process-exec-all`   Deny all process execution

### Examples

Allow read operation for current directory:
```console
sbx --allow-file-read . ls .
# same as above
sbx --allow-file-read='.' ls .

# with command flags
sbx --allow-file-read='.' -- ls -l .
```

Allow network operation for `localhost:8080`:
```console
sbx --allow-network='localhost:8080' curl http://localhost:8080
```

Allow network operation for remote host (with CA certificate access):
```console
sbx --allow-network='*:443' --allow-file-read='/opt/local' curl https://syum.ai/ascii
```

---

## License

MIT
