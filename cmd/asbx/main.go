package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/syumai/sbx/agentprofile"
	"github.com/urfave/cli/v3"
)

const (
	flagAgentHarness = "agent-harness"
	flagProjectPath  = "project-path"
	flagPrintProfile = "print-profile"
	flagExtraRead    = "extra-read"
	flagExtraWrite   = "extra-write"
	flagExtraExec    = "extra-exec"
)

func main() {
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:     flagAgentHarness,
			Aliases:  []string{"a"},
			Usage:    "agent harness type: opencode, claude, codex",
			Required: true,
		},
		&cli.StringFlag{
			Name:    flagProjectPath,
			Aliases: []string{"p"},
			Usage:   "project directory path (defaults to current directory)",
			Value:   ".",
		},
		&cli.BoolFlag{
			Name:    flagPrintProfile,
			Aliases: []string{"P"},
			Usage:   "print the generated sandbox profile and exit",
		},
		&cli.StringSliceFlag{
			Name:  flagExtraRead,
			Usage: "additional read-only paths",
		},
		&cli.StringSliceFlag{
			Name:  flagExtraWrite,
			Usage: "additional write paths",
		},
		&cli.StringSliceFlag{
			Name:  flagExtraExec,
			Usage: "additional executable paths",
		},
	}

	cmd := &cli.Command{
		Name:  "asbx",
		Usage: "run coding agents in a macOS sandbox",
		UsageText: `asbx --agent-harness=<agent> [flags] <command> [command-args...]
asbx --agent-harness=<agent> [flags] -- <command> [command-flags] [command-args...]

Examples:
    # Run opencode in sandbox for current directory
    asbx --agent-harness=opencode opencode

    # Run opencode for a specific project
    asbx --agent-harness=opencode --project-path=/path/to/project opencode

    # Run claude code with extra paths
    asbx --agent-harness=claude --extra-read=/custom/config -- claude

    # Print the generated profile without running
    asbx --agent-harness=opencode --print-profile

Supported agents:
    opencode  - SST OpenCode AI coding agent (https://opencode.ai)
    claude    - Anthropic Claude Code
    codex     - OpenAI Codex CLI`,
		Flags: flags,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			// Parse agent type
			agentType, err := agentprofile.ParseAgentType(cmd.String(flagAgentHarness))
			if err != nil {
				return err
			}

			// Get project path
			projectPath := cmd.String(flagProjectPath)
			absProjectPath, err := filepath.Abs(projectPath)
			if err != nil {
				return fmt.Errorf("failed to resolve project path: %w", err)
			}

			// Verify project path exists
			if info, err := os.Stat(absProjectPath); err != nil {
				return fmt.Errorf("project path does not exist: %s", absProjectPath)
			} else if !info.IsDir() {
				return fmt.Errorf("project path is not a directory: %s", absProjectPath)
			}

			// Get agent profile
			profile, err := agentprofile.GetProfile(agentType, absProjectPath)
			if err != nil {
				return fmt.Errorf("failed to get agent profile: %w", err)
			}

			// Add extra paths from flags
			for _, path := range cmd.StringSlice(flagExtraRead) {
				absPath, err := filepath.Abs(path)
				if err != nil {
					return fmt.Errorf("failed to resolve extra read path %s: %w", path, err)
				}
				profile.AddReadPath(absPath)
			}
			for _, path := range cmd.StringSlice(flagExtraWrite) {
				absPath, err := filepath.Abs(path)
				if err != nil {
					return fmt.Errorf("failed to resolve extra write path %s: %w", path, err)
				}
				profile.AddWritePath(absPath)
			}
			for _, path := range cmd.StringSlice(flagExtraExec) {
				profile.AddExecPath(path)
			}

			// Generate SBPL profile
			sbplProfile := profile.ToSBPL()

			// Print profile and exit if requested
			if cmd.Bool(flagPrintProfile) {
				fmt.Println(sbplProfile)
				return nil
			}

			// Get command to run
			command := cmd.Args().First()
			if command == "" {
				return cli.ShowAppHelp(cmd)
			}

			// Resolve command path
			commandPath, err := exec.LookPath(command)
			if err != nil {
				return fmt.Errorf("command not found: %s", command)
			}

			// Add the command to the profile's exec paths
			profile.AddExecPath(commandPath)
			sbplProfile = profile.ToSBPL()

			// Run the command in sandbox
			return sandboxExec(ctx, sbplProfile, commandPath, cmd.Args().Tail()...)
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func sandboxExec(ctx context.Context, profile string, command string, args ...string) error {
	if command == "" {
		return errors.New("command is required")
	}
	sandboxArgs := []string{"-p", profile, command}
	sandboxArgs = append(sandboxArgs, args...)

	cmd := exec.CommandContext(ctx, "sandbox-exec", sandboxArgs...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to start command: %v\n", err)
		os.Exit(1)
	}

	err := cmd.Wait()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
			} else {
				exitCode = 1
			}
		} else {
			fmt.Fprintf(os.Stderr, "command execution failed: %v\n", err)
			exitCode = 1
		}
	}

	os.Exit(exitCode)
	return nil
}
