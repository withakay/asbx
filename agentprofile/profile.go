// Package agentprofile defines sandbox profiles for different coding agents
package agentprofile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AgentType represents a supported coding agent
type AgentType string

const (
	AgentTypeOpencode AgentType = "opencode"
	AgentTypeClaude   AgentType = "claude"
	AgentTypeCodex    AgentType = "codex"
)

// ParseAgentType parses a string into an AgentType
func ParseAgentType(s string) (AgentType, error) {
	switch strings.ToLower(s) {
	case "opencode":
		return AgentTypeOpencode, nil
	case "claude":
		return AgentTypeClaude, nil
	case "codex":
		return AgentTypeCodex, nil
	default:
		return "", fmt.Errorf("unknown agent type: %s (supported: opencode, claude, codex)", s)
	}
}

// Profile represents a sandbox profile for a coding agent
type Profile struct {
	// Paths the agent can read from (subpath)
	ReadPaths []string
	// Paths the agent can write to (subpath)
	WritePaths []string
	// Paths the agent can both read and write (subpath)
	ReadWritePaths []string
	// Specific binaries/paths that can be executed (literal)
	ExecPaths []string
	// Binary directories to allow execution from (subpath)
	ExecDirs []string
	// Allow all outbound network
	AllowOutboundNetwork bool
	// Allow inbound network (for server mode)
	AllowInboundNetwork bool
}

// GetProfile returns the sandbox profile for a given agent type
// projectPath is the path to the project directory that the agent should have access to
func GetProfile(agentType AgentType, projectPath string) (*Profile, error) {
	absProjectPath, err := filepath.Abs(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve project path: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome == "" {
		xdgConfigHome = filepath.Join(home, ".config")
	}

	xdgDataHome := os.Getenv("XDG_DATA_HOME")
	if xdgDataHome == "" {
		xdgDataHome = filepath.Join(home, ".local", "share")
	}

	xdgCacheHome := os.Getenv("XDG_CACHE_HOME")
	if xdgCacheHome == "" {
		xdgCacheHome = filepath.Join(home, ".cache")
	}

	switch agentType {
	case AgentTypeOpencode:
		return getOpencodeProfile(absProjectPath, home, xdgConfigHome, xdgDataHome, xdgCacheHome)
	case AgentTypeClaude:
		return getClaudeProfile(absProjectPath, home, xdgConfigHome, xdgDataHome, xdgCacheHome)
	case AgentTypeCodex:
		return getCodexProfile(absProjectPath, home, xdgConfigHome, xdgDataHome, xdgCacheHome)
	default:
		return nil, fmt.Errorf("unsupported agent type: %s", agentType)
	}
}

// getOpencodeProfile returns the sandbox profile for opencode
// Based on: https://opencode.ai/docs/config/
// - Config: ~/.opencode.json, $XDG_CONFIG_HOME/opencode/.opencode.json, ./.opencode.json
// - Auth: ~/.local/share/opencode/auth.json
// - Plugins: .opencode/plugin/, ~/.config/opencode/plugin/
// - Agents: ~/.config/opencode/agent/, .opencode/agent/
// - Rules: ~/.config/opencode/AGENTS.md
// - SQLite storage for persistence
func getOpencodeProfile(projectPath, home, xdgConfigHome, xdgDataHome, xdgCacheHome string) (*Profile, error) {
	return &Profile{
		ReadPaths: []string{
			// System binaries and libraries (read-only)
			"/usr",
			"/bin",
			"/sbin",
			"/System",
			"/Library",
			"/opt",
			"/nix",
			"/private/var/db", // dyld shared cache

			// Homebrew paths (read for discovering tools)
			"/opt/homebrew",
			"/usr/local/Homebrew",

			// Global opencode config (read-only)
			filepath.Join(xdgConfigHome, "opencode"),
			filepath.Join(home, ".opencode.json"),

			// Git configuration (read-only)
			filepath.Join(home, ".gitconfig"),
			filepath.Join(xdgConfigHome, "git"),

			// SSH public keys only (for git operations, not private keys)
			filepath.Join(home, ".ssh", "known_hosts"),
			filepath.Join(home, ".ssh", "config"),

			// Shell configs (for environment)
			filepath.Join(home, ".zshrc"),
			filepath.Join(home, ".bashrc"),
			filepath.Join(home, ".profile"),
			filepath.Join(home, ".zshenv"),
			filepath.Join(home, ".bash_profile"),
			filepath.Join(home, ".zprofile"),

			// Node version managers (read-only)
			filepath.Join(home, ".nvm"),
			filepath.Join(home, ".fnm"),
			filepath.Join(home, ".nodenv"),
			filepath.Join(home, ".n"),

			// Other version managers
			filepath.Join(home, ".pyenv"),
			filepath.Join(home, ".rbenv"),
			filepath.Join(home, ".rustup"),
			filepath.Join(home, ".cargo"),
			filepath.Join(home, ".goenv"),
			filepath.Join(home, ".sdkman"),

			// Local binaries (read-only for discovery)
			filepath.Join(home, ".local", "bin"),

			// Go paths
			filepath.Join(home, "go"),
		},

		WritePaths: []string{
			// Temp directories
			"/tmp",
			"/private/tmp",
			"/var/tmp",
			"/private/var/tmp",
			"/var/folders", // macOS temp folders
			"/private/var/folders",

			// Cache
			xdgCacheHome,
		},

		ReadWritePaths: []string{
			// Project directory - full access
			projectPath,

			// Opencode data (auth, sessions, SQLite)
			filepath.Join(xdgDataHome, "opencode"),

			// Opencode config that may need writing
			filepath.Join(xdgConfigHome, "opencode"),
		},

		ExecDirs: []string{
			// System binaries
			"/usr/bin",
			"/bin",
			"/usr/sbin",
			"/sbin",
			"/usr/local/bin",
			"/opt/homebrew/bin",
			"/opt/homebrew/sbin",

			// Nix
			"/nix/store",
			"/run/current-system/sw/bin",

			// User binaries
			filepath.Join(home, ".local", "bin"),
			filepath.Join(home, ".cargo", "bin"),
			filepath.Join(home, "go", "bin"),

			// Node managers
			filepath.Join(home, ".nvm"),
			filepath.Join(home, ".fnm"),
			filepath.Join(home, ".nodenv"),
			filepath.Join(home, ".n"),

			// Python
			filepath.Join(home, ".pyenv"),

			// Project node_modules binaries
			filepath.Join(projectPath, "node_modules", ".bin"),
		},

		AllowOutboundNetwork: true,  // Required for API calls
		AllowInboundNetwork:  true,  // Required for server mode (opencode serve)
	}, nil
}

// getClaudeProfile returns the sandbox profile for Claude Code
func getClaudeProfile(projectPath, home, xdgConfigHome, xdgDataHome, xdgCacheHome string) (*Profile, error) {
	return &Profile{
		ReadPaths: []string{
			// System binaries and libraries
			"/usr",
			"/bin",
			"/sbin",
			"/System",
			"/Library",
			"/opt",
			"/nix",
			"/private/var/db",

			// Homebrew
			"/opt/homebrew",
			"/usr/local/Homebrew",

			// Claude config
			filepath.Join(home, ".claude"),
			filepath.Join(home, ".claude.json"),

			// Git configuration
			filepath.Join(home, ".gitconfig"),
			filepath.Join(xdgConfigHome, "git"),

			// SSH (limited)
			filepath.Join(home, ".ssh", "known_hosts"),
			filepath.Join(home, ".ssh", "config"),

			// Shell configs
			filepath.Join(home, ".zshrc"),
			filepath.Join(home, ".bashrc"),
			filepath.Join(home, ".profile"),
			filepath.Join(home, ".zshenv"),
			filepath.Join(home, ".bash_profile"),
			filepath.Join(home, ".zprofile"),

			// Version managers
			filepath.Join(home, ".nvm"),
			filepath.Join(home, ".fnm"),
			filepath.Join(home, ".nodenv"),
			filepath.Join(home, ".pyenv"),
			filepath.Join(home, ".rbenv"),
			filepath.Join(home, ".rustup"),
			filepath.Join(home, ".cargo"),
			filepath.Join(home, ".goenv"),

			// Editor configs (read-only)
			filepath.Join(home, ".vscode"),
			filepath.Join(home, ".cursor"),

			// Local binaries
			filepath.Join(home, ".local", "bin"),
			filepath.Join(home, "go"),
		},

		WritePaths: []string{
			// Temp
			"/tmp",
			"/private/tmp",
			"/var/tmp",
			"/private/var/tmp",
			"/var/folders",
			"/private/var/folders",

			// Cache
			xdgCacheHome,
		},

		ReadWritePaths: []string{
			// Project
			projectPath,

			// Claude data
			filepath.Join(home, ".claude"),
		},

		ExecDirs: []string{
			"/usr/bin",
			"/bin",
			"/usr/sbin",
			"/sbin",
			"/usr/local/bin",
			"/opt/homebrew/bin",
			"/opt/homebrew/sbin",
			"/nix/store",
			"/run/current-system/sw/bin",
			filepath.Join(home, ".local", "bin"),
			filepath.Join(home, ".cargo", "bin"),
			filepath.Join(home, "go", "bin"),
			filepath.Join(home, ".nvm"),
			filepath.Join(home, ".fnm"),
			filepath.Join(home, ".pyenv"),
			filepath.Join(projectPath, "node_modules", ".bin"),
		},

		AllowOutboundNetwork: true,
		AllowInboundNetwork:  false,
	}, nil
}

// getCodexProfile returns the sandbox profile for OpenAI Codex CLI
func getCodexProfile(projectPath, home, xdgConfigHome, xdgDataHome, xdgCacheHome string) (*Profile, error) {
	return &Profile{
		ReadPaths: []string{
			// System
			"/usr",
			"/bin",
			"/sbin",
			"/System",
			"/Library",
			"/opt",
			"/nix",
			"/private/var/db",

			// Homebrew
			"/opt/homebrew",
			"/usr/local/Homebrew",

			// Git
			filepath.Join(home, ".gitconfig"),
			filepath.Join(xdgConfigHome, "git"),

			// SSH (limited)
			filepath.Join(home, ".ssh", "known_hosts"),
			filepath.Join(home, ".ssh", "config"),

			// Shell
			filepath.Join(home, ".zshrc"),
			filepath.Join(home, ".bashrc"),
			filepath.Join(home, ".profile"),

			// Version managers
			filepath.Join(home, ".nvm"),
			filepath.Join(home, ".fnm"),
			filepath.Join(home, ".pyenv"),
			filepath.Join(home, ".cargo"),

			// Local binaries
			filepath.Join(home, ".local", "bin"),
			filepath.Join(home, "go"),
		},

		WritePaths: []string{
			"/tmp",
			"/private/tmp",
			"/var/tmp",
			"/private/var/tmp",
			"/var/folders",
			"/private/var/folders",
			xdgCacheHome,
		},

		ReadWritePaths: []string{
			projectPath,
			// Codex config location (may vary)
			filepath.Join(xdgConfigHome, "codex"),
			filepath.Join(xdgDataHome, "codex"),
		},

		ExecDirs: []string{
			"/usr/bin",
			"/bin",
			"/usr/sbin",
			"/sbin",
			"/usr/local/bin",
			"/opt/homebrew/bin",
			"/opt/homebrew/sbin",
			"/nix/store",
			filepath.Join(home, ".local", "bin"),
			filepath.Join(home, ".cargo", "bin"),
			filepath.Join(home, "go", "bin"),
			filepath.Join(home, ".nvm"),
			filepath.Join(home, ".pyenv"),
			filepath.Join(projectPath, "node_modules", ".bin"),
		},

		AllowOutboundNetwork: true,
		AllowInboundNetwork:  false,
	}, nil
}
