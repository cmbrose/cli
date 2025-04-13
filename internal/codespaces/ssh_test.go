package codespaces

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type parseTestCase struct {
	Args       []string
	ParsedArgs []string
	Command    []string
	Error      string
}

func TestParseSSHArgs(t *testing.T) {
	testCases := []parseTestCase{
		{}, // empty test case
		{
			Args:       []string{"-X", "-Y"},
			ParsedArgs: []string{"-X", "-Y"},
			Command:    nil,
		},
		{
			Args:       []string{"-X", "-Y", "-o", "someoption=test"},
			ParsedArgs: []string{"-X", "-Y", "-o", "someoption=test"},
			Command:    nil,
		},
		{
			Args:       []string{"-X", "-Y", "-o", "someoption=test", "somecommand"},
			ParsedArgs: []string{"-X", "-Y", "-o", "someoption=test"},
			Command:    []string{"somecommand"},
		},
		{
			Args:       []string{"-X", "-Y", "-o", "someoption=test", "echo", "test"},
			ParsedArgs: []string{"-X", "-Y", "-o", "someoption=test"},
			Command:    []string{"echo", "test"},
		},
		{
			Args:       []string{"somecommand"},
			ParsedArgs: []string{},
			Command:    []string{"somecommand"},
		},
		{
			Args:       []string{"echo", "test"},
			ParsedArgs: []string{},
			Command:    []string{"echo", "test"},
		},
		{
			Args:       []string{"-v", "echo", "hello", "world"},
			ParsedArgs: []string{"-v"},
			Command:    []string{"echo", "hello", "world"},
		},
		{
			Args:       []string{"-L", "-l"},
			ParsedArgs: []string{"-L", "-l"},
			Command:    nil,
		},
		{
			Args:       []string{"-v", "echo", "-n", "test"},
			ParsedArgs: []string{"-v"},
			Command:    []string{"echo", "-n", "test"},
		},
		{
			Args:       []string{"-v", "echo", "-b", "test"},
			ParsedArgs: []string{"-v"},
			Command:    []string{"echo", "-b", "test"},
		},
		{
			Args:       []string{"-b"},
			ParsedArgs: nil,
			Command:    nil,
			Error:      "flag: -b requires an argument",
		},
	}

	for _, tcase := range testCases {
		args, command, err := ParseSSHArgs(tcase.Args)

		checkParseResult(t, tcase, args, command, err)
	}
}

func TestParseSCPArgs(t *testing.T) {
	testCases := []parseTestCase{
		{}, // empty test case
		{
			Args:       []string{"-X", "-Y"},
			ParsedArgs: []string{"-X", "-Y"},
			Command:    nil,
		},
		{
			Args:       []string{"-X", "-Y", "-o", "someoption=test"},
			ParsedArgs: []string{"-X", "-Y", "-o", "someoption=test"},
			Command:    nil,
		},
		{
			Args:       []string{"-X", "-Y", "-o", "someoption=test", "local/file", "remote:file"},
			ParsedArgs: []string{"-X", "-Y", "-o", "someoption=test"},
			Command:    []string{"local/file", "remote:file"},
		},
		{
			Args:       []string{"-X", "-Y", "-o", "someoption=test", "local/file", "remote:file"},
			ParsedArgs: []string{"-X", "-Y", "-o", "someoption=test"},
			Command:    []string{"local/file", "remote:file"},
		},
		{
			Args:       []string{"local/file", "remote:file"},
			ParsedArgs: []string{},
			Command:    []string{"local/file", "remote:file"},
		},
		{
			Args:       []string{"-c"},
			ParsedArgs: nil,
			Command:    nil,
			Error:      "flag: -c requires an argument",
		},
	}

	for _, tcase := range testCases {
		args, command, err := parseSCPArgs(tcase.Args)

		checkParseResult(t, tcase, args, command, err)
	}
}

func checkParseResult(t *testing.T, tcase parseTestCase, gotArgs, gotCmd []string, gotErr error) {
	if tcase.Error != "" {
		if gotErr == nil {
			t.Errorf("expected error and got nil: %#v", tcase)
		}

		if gotErr.Error() != tcase.Error {
			t.Errorf("error does not match expected error, got: '%s', expected: '%s'", gotErr.Error(), tcase.Error)
		}

		return
	}

	if gotErr != nil {
		t.Errorf("unexpected error: %v on test case: %#v", gotErr, tcase)
		return
	}

	argsStr, parsedArgsStr := fmt.Sprintf("%s", gotArgs), fmt.Sprintf("%s", tcase.ParsedArgs)
	if argsStr != parsedArgsStr {
		t.Errorf("args do not match parsed args. got: '%s', expected: '%s'", argsStr, parsedArgsStr)
	}

	commandStr, parsedCommandStr := fmt.Sprintf("%s", gotCmd), fmt.Sprintf("%s", tcase.Command)
	if commandStr != parsedCommandStr {
		t.Errorf("command does not match parsed command. got: '%s', expected: '%s'", commandStr, parsedCommandStr)
	}
}

func TestNewCommandFromEnv(t *testing.T) {
	// Save original env and restore after test
	originalEnv := os.Getenv("GH_CS_TEST_COMMAND")
	defer os.Setenv("GH_CS_TEST_COMMAND", originalEnv)

	tests := []struct {
		name        string
		envValue    string
		defaultCmd  string
		cmdArgs     []string
		wantCmd     string
		wantArgs    []string
		wantErr     bool
		errContains string
	}{
		{
			name:       "uses default command when env not set",
			envValue:   "",
			defaultCmd: "ssh",
			cmdArgs:    []string{"-p", "22", "user@host"},
			wantCmd:    "ssh",
			wantArgs:   []string{"-p", "22", "user@host"},
		},
		{
			name:       "uses env command when set",
			envValue:   "/usr/local/bin/ssh",
			defaultCmd: "ssh",
			cmdArgs:    []string{"-p", "22", "user@host"},
			wantCmd:    "/usr/local/bin/ssh",
			wantArgs:   []string{"-p", "22", "user@host"},
		},
		{
			name:       "handles command with arguments in env",
			envValue:   "/usr/local/bin/ssh -o StrictHostKeyChecking=no",
			defaultCmd: "ssh",
			cmdArgs:    []string{"-p", "22", "user@host"},
			wantCmd:    "/usr/local/bin/ssh",
			wantArgs:   []string{"-o", "StrictHostKeyChecking=no", "-p", "22", "user@host"},
		},
		{
			name:       "handles quoted paths with spaces",
			envValue:   `"/path/with a/space/ssh" -o "key=value with spaces"`,
			defaultCmd: "ssh",
			cmdArgs:    []string{"-p", "22", "user@host"},
			wantCmd:    "/path/with a/space/ssh",
			wantArgs:   []string{"-o", "key=value with spaces", "-p", "22", "user@host"},
		},
		{
			name:       "handles escaped spaces",
			envValue:   `/path/with\ a/space/ssh -o key=value`,
			defaultCmd: "ssh",
			cmdArgs:    []string{"-p", "22", "user@host"},
			wantCmd:    "/path/with a/space/ssh",
			wantArgs:   []string{"-o", "key=value", "-p", "22", "user@host"},
		},
		{
			name:        "returns error for invalid env value",
			envValue:    `/path/with a/space/ssh" -o key=value`, // Unclosed quote
			defaultCmd:  "ssh",
			cmdArgs:     []string{"-p", "22", "user@host"},
			wantErr:     true,
			errContains: "invalid GH_CS_TEST_COMMAND",
		},
		{
			name:        "returns error for empty env value",
			envValue:    "",
			defaultCmd:  "nonexistent",
			cmdArgs:     []string{"-p", "22", "user@host"},
			wantErr:     true,
			errContains: "failed to execute nonexistent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable for test
			if tt.envValue != "" {
				os.Setenv("GH_CS_TEST_COMMAND", tt.envValue)
			} else {
				os.Unsetenv("GH_CS_TEST_COMMAND")
			}

			// Create a temporary directory for the default command
			if tt.envValue == "" && !tt.wantErr {
				tmpDir := t.TempDir()
				cmdPath := filepath.Join(tmpDir, tt.defaultCmd)
				err := os.WriteFile(cmdPath, []byte("#!/bin/sh\nexit 0"), 0755)
				assert.NoError(t, err)
				
				// Add the temp directory to PATH
				path := os.Getenv("PATH")
				os.Setenv("PATH", tmpDir+":"+path)
				defer os.Setenv("PATH", path)
			}

			cmd, err := newCommandFromEnv(context.Background(), "GH_CS_TEST_COMMAND", tt.defaultCmd, tt.cmdArgs)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			assert.NoError(t, err)
			assert.True(t, strings.HasSuffix(cmd.Path, tt.wantCmd), 
				"Expected command path to end with %q, got %q", tt.wantCmd, cmd.Path)
			assert.Equal(t, tt.wantArgs, cmd.Args[1:]) // Args[0] is the command path
		})
	}
}

func TestNewSSHCommand(t *testing.T) {
	// Save original env and restore after test
	originalEnv := os.Getenv("GH_CS_SSH_COMMAND")
	defer os.Setenv("GH_CS_SSH_COMMAND", originalEnv)

	tests := []struct {
		name        string
		envValue    string
		port        int
		dst         string
		cmdArgs     []string
		command     []string
		wantCmd     string
		wantArgs    []string
		wantErr     bool
		errContains string
	}{
		{
			name:     "uses default ssh command",
			envValue: "",
			port:     1234,
			dst:      "user@host",
			cmdArgs:  []string{"-v"},
			command:  []string{"echo", "hello"},
			wantCmd:  "ssh",
			wantArgs: []string{"-v", "-p", "1234", "-o", "NoHostAuthenticationForLocalhost=yes", "-o", "PasswordAuthentication=no", "-C", "user@host", "echo", "hello"},
		},
		{
			name:     "uses custom ssh command from env",
			envValue: "/usr/local/bin/ssh",
			port:     1234,
			dst:      "user@host",
			cmdArgs:  []string{"-v"},
			command:  []string{"echo", "hello"},
			wantCmd:  "/usr/local/bin/ssh",
			wantArgs: []string{"-v", "-p", "1234", "-o", "NoHostAuthenticationForLocalhost=yes", "-o", "PasswordAuthentication=no", "-C", "user@host", "echo", "hello"},
		},
		{
			name:     "uses custom ssh command with args from env",
			envValue: "/usr/local/bin/ssh -o StrictHostKeyChecking=no",
			port:     1234,
			dst:      "user@host",
			cmdArgs:  []string{"-v"},
			command:  []string{"echo", "hello"},
			wantCmd:  "/usr/local/bin/ssh",
			wantArgs: []string{"-o", "StrictHostKeyChecking=no", "-v", "-p", "1234", "-o", "NoHostAuthenticationForLocalhost=yes", "-o", "PasswordAuthentication=no", "-C", "user@host", "echo", "hello"},
		},
		{
			name:     "handles quoted paths with spaces in env",
			envValue: `"/path/with a/space/ssh" -o "key=value with spaces"`,
			port:     1234,
			dst:      "user@host",
			cmdArgs:  []string{"-v"},
			command:  []string{"echo", "hello"},
			wantCmd:  "/path/with a/space/ssh",
			wantArgs: []string{"-o", "key=value with spaces", "-v", "-p", "1234", "-o", "NoHostAuthenticationForLocalhost=yes", "-o", "PasswordAuthentication=no", "-C", "user@host", "echo", "hello"},
		},
		{
			name:        "returns error for invalid env value",
			envValue:    `/path/with a/space/ssh" -o key=value`, // Unclosed quote
			port:        1234,
			dst:         "user@host",
			cmdArgs:     []string{"-v"},
			command:     []string{"echo", "hello"},
			wantErr:     true,
			errContains: "invalid GH_CS_SSH_COMMAND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable for test
			if tt.envValue != "" {
				os.Setenv("GH_CS_SSH_COMMAND", tt.envValue)
			} else {
				os.Unsetenv("GH_CS_SSH_COMMAND")
			}

			// Create a temporary directory for the default command
			if tt.envValue == "" && !tt.wantErr {
				tmpDir := t.TempDir()
				cmdPath := filepath.Join(tmpDir, "ssh")
				err := os.WriteFile(cmdPath, []byte("#!/bin/sh\nexit 0"), 0755)
				assert.NoError(t, err)
				
				// Add the temp directory to PATH
				path := os.Getenv("PATH")
				os.Setenv("PATH", tmpDir+":"+path)
				defer os.Setenv("PATH", path)
			}

			cmd, connArgs, err := newSSHCommand(context.Background(), tt.port, tt.dst, tt.cmdArgs, tt.command)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			assert.NoError(t, err)
			assert.True(t, strings.HasSuffix(cmd.Path, tt.wantCmd), 
				"Expected command path to end with %q, got %q", tt.wantCmd, cmd.Path)
			assert.Equal(t, tt.wantArgs, cmd.Args[1:]) // Args[0] is the command path
			assert.Equal(t, []string{"-p", "1234", "-o", "NoHostAuthenticationForLocalhost=yes", "-o", "PasswordAuthentication=no"}, connArgs)
		})
	}
}

func TestNewSCPCommand(t *testing.T) {
	// Save original env and restore after test
	originalEnv := os.Getenv("GH_CS_SCP_COMMAND")
	defer os.Setenv("GH_CS_SCP_COMMAND", originalEnv)

	tests := []struct {
		name        string
		envValue    string
		port        int
		dst         string
		cmdArgs     []string
		wantCmd     string
		wantArgs    []string
		wantErr     bool
		errContains string
	}{
		{
			name:     "uses default scp command",
			envValue: "",
			port:     1234,
			dst:      "user@host",
			cmdArgs:  []string{"-v", "local/file", "remote:file"},
			wantCmd:  "scp",
			wantArgs: []string{"-v", "-P", "1234", "-o", "NoHostAuthenticationForLocalhost=yes", "-o", "PasswordAuthentication=no", "-C", "local/file", "user@host:file"},
		},
		{
			name:     "uses custom scp command from env",
			envValue: "/usr/local/bin/scp",
			port:     1234,
			dst:      "user@host",
			cmdArgs:  []string{"-v", "local/file", "remote:file"},
			wantCmd:  "/usr/local/bin/scp",
			wantArgs: []string{"-v", "-P", "1234", "-o", "NoHostAuthenticationForLocalhost=yes", "-o", "PasswordAuthentication=no", "-C", "local/file", "user@host:file"},
		},
		{
			name:     "uses custom scp command with args from env",
			envValue: "/usr/local/bin/scp -o StrictHostKeyChecking=no",
			port:     1234,
			dst:      "user@host",
			cmdArgs:  []string{"-v", "local/file", "remote:file"},
			wantCmd:  "/usr/local/bin/scp",
			wantArgs: []string{"-o", "StrictHostKeyChecking=no", "-v", "-P", "1234", "-o", "NoHostAuthenticationForLocalhost=yes", "-o", "PasswordAuthentication=no", "-C", "local/file", "user@host:file"},
		},
		{
			name:     "handles quoted paths with spaces in env",
			envValue: `"/path/with a/space/scp" -o "key=value with spaces"`,
			port:     1234,
			dst:      "user@host",
			cmdArgs:  []string{"-v", "local/file", "remote:file"},
			wantCmd:  "/path/with a/space/scp",
			wantArgs: []string{"-o", "key=value with spaces", "-v", "-P", "1234", "-o", "NoHostAuthenticationForLocalhost=yes", "-o", "PasswordAuthentication=no", "-C", "local/file", "user@host:file"},
		},
		{
			name:        "returns error for invalid env value",
			envValue:    `/path/with a/space/scp" -o key=value`, // Unclosed quote
			port:        1234,
			dst:         "user@host",
			cmdArgs:     []string{"-v", "local/file", "remote:file"},
			wantErr:     true,
			errContains: "invalid GH_CS_SCP_COMMAND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable for test
			if tt.envValue != "" {
				os.Setenv("GH_CS_SCP_COMMAND", tt.envValue)
			} else {
				os.Unsetenv("GH_CS_SCP_COMMAND")
			}

			// Create a temporary directory for the default command
			if tt.envValue == "" && !tt.wantErr {
				tmpDir := t.TempDir()
				cmdPath := filepath.Join(tmpDir, "scp")
				err := os.WriteFile(cmdPath, []byte("#!/bin/sh\nexit 0"), 0755)
				assert.NoError(t, err)
				
				// Add the temp directory to PATH
				path := os.Getenv("PATH")
				os.Setenv("PATH", tmpDir+":"+path)
				defer os.Setenv("PATH", path)
			}

			cmd, err := newSCPCommand(context.Background(), tt.port, tt.dst, tt.cmdArgs)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			assert.NoError(t, err)
			assert.True(t, strings.HasSuffix(cmd.Path, tt.wantCmd), 
				"Expected command path to end with %q, got %q", tt.wantCmd, cmd.Path)
			assert.Equal(t, tt.wantArgs, cmd.Args[1:]) // Args[0] is the command path
		})
	}
}
