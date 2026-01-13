package agentprofile

import (
	"fmt"
	"strings"
)

// ToSBPL generates a macOS sandbox profile (SBPL) from an agent Profile
func (p *Profile) ToSBPL() string {
	var b strings.Builder

	// Header
	b.WriteString("(version 1)\n")
	b.WriteString("(import \"system.sb\")\n")
	b.WriteString("(deny default)\n")
	b.WriteString("\n")

	// Allow dylib access (required for any process execution)
	b.WriteString("; Allow dynamic library access\n")
	b.WriteString("(allow file-read*\n")
	b.WriteString("  (subpath \"/opt/local/lib\")\n")
	b.WriteString("  (subpath \"/usr/lib\")\n")
	b.WriteString("  (subpath \"/usr/local/lib\")\n")
	b.WriteString("  (subpath \"/opt/homebrew/lib\")\n")
	b.WriteString("  (subpath \"/Library/Developer\")\n")
	b.WriteString(")\n")
	b.WriteString("\n")

	// Allow unix sockets if network is enabled
	if p.AllowOutboundNetwork || p.AllowInboundNetwork {
		b.WriteString("; Allow unix sockets for local IPC\n")
		b.WriteString("(allow network* (local unix-socket))\n")
		b.WriteString("\n")
	}

	// Network rules
	if p.AllowOutboundNetwork {
		b.WriteString("; Allow all outbound network\n")
		b.WriteString("(allow network-outbound)\n")
		b.WriteString("\n")
	}

	if p.AllowInboundNetwork {
		b.WriteString("; Allow inbound network (for server mode)\n")
		b.WriteString("(allow network-inbound)\n")
		b.WriteString("\n")
	}

	// Read-only paths
	if len(p.ReadPaths) > 0 {
		b.WriteString("; Read-only paths\n")
		b.WriteString("(allow file-read*\n")
		for _, path := range p.ReadPaths {
			b.WriteString(fmt.Sprintf("  (subpath %q)\n", path))
		}
		b.WriteString(")\n")
		b.WriteString("\n")
	}

	// Write-only paths
	if len(p.WritePaths) > 0 {
		b.WriteString("; Write paths (includes read)\n")
		b.WriteString("(allow file-write*\n")
		for _, path := range p.WritePaths {
			b.WriteString(fmt.Sprintf("  (subpath %q)\n", path))
		}
		b.WriteString(")\n")
		// Also need read access
		b.WriteString("(allow file-read*\n")
		for _, path := range p.WritePaths {
			b.WriteString(fmt.Sprintf("  (subpath %q)\n", path))
		}
		b.WriteString(")\n")
		b.WriteString("\n")
	}

	// Read-write paths (project directory etc)
	if len(p.ReadWritePaths) > 0 {
		b.WriteString("; Full access paths (project directory, config)\n")
		b.WriteString("(allow file*\n")
		for _, path := range p.ReadWritePaths {
			b.WriteString(fmt.Sprintf("  (subpath %q)\n", path))
		}
		b.WriteString(")\n")
		b.WriteString("\n")
	}

	// Process execution - directories
	if len(p.ExecDirs) > 0 {
		b.WriteString("; Executable directories\n")
		b.WriteString("(allow process-exec\n")
		for _, path := range p.ExecDirs {
			b.WriteString(fmt.Sprintf("  (subpath %q)\n", path))
		}
		b.WriteString(")\n")
		b.WriteString("\n")
	}

	// Process execution - specific paths
	if len(p.ExecPaths) > 0 {
		b.WriteString("; Specific executables\n")
		b.WriteString("(allow process-exec\n")
		for _, path := range p.ExecPaths {
			b.WriteString(fmt.Sprintf("  (literal %q)\n", path))
		}
		b.WriteString(")\n")
		b.WriteString("\n")
	}

	// Allow process-fork for spawning child processes
	b.WriteString("; Allow process operations\n")
	b.WriteString("(allow process-fork)\n")
	b.WriteString("\n")

	// Allow signal handling
	b.WriteString("; Allow signal handling\n")
	b.WriteString("(allow signal (target self))\n")
	b.WriteString("\n")

	// Allow sysctl read for system info (required by many tools)
	b.WriteString("; Allow reading system info\n")
	b.WriteString("(allow sysctl-read)\n")
	b.WriteString("\n")

	// Allow mach operations (required for IPC)
	b.WriteString("; Allow mach IPC\n")
	b.WriteString("(allow mach-lookup)\n")
	b.WriteString("(allow mach-register)\n")
	b.WriteString("\n")

	// Allow IOKit for device info (some tools need this)
	b.WriteString("; Allow IOKit for hardware info\n")
	b.WriteString("(allow iokit-open)\n")
	b.WriteString("\n")

	// Allow pseudo-terminals (required for TUI)
	b.WriteString("; Allow pseudo-terminals\n")
	b.WriteString("(allow file-read* file-write* file-ioctl (regex #\"^/dev/tty\"))\n")
	b.WriteString("(allow file-read* file-write* file-ioctl (regex #\"^/dev/pty\"))\n")
	b.WriteString("(allow file-read* file-write* (literal \"/dev/null\"))\n")
	b.WriteString("(allow file-read* (literal \"/dev/random\"))\n")
	b.WriteString("(allow file-read* (literal \"/dev/urandom\"))\n")

	return b.String()
}

// AddExecPath adds a specific executable path to the profile
func (p *Profile) AddExecPath(path string) {
	p.ExecPaths = append(p.ExecPaths, path)
}

// AddReadPath adds a read-only path to the profile
func (p *Profile) AddReadPath(path string) {
	p.ReadPaths = append(p.ReadPaths, path)
}

// AddWritePath adds a write path to the profile
func (p *Profile) AddWritePath(path string) {
	p.WritePaths = append(p.WritePaths, path)
}

// AddReadWritePath adds a full access path to the profile
func (p *Profile) AddReadWritePath(path string) {
	p.ReadWritePaths = append(p.ReadWritePaths, path)
}
