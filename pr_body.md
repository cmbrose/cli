# Add environment variables for SSH and SCP commands in Codespaces

Fixes #9244

This PR adds support for two new environment variables that allow users to customize the SSH and SCP commands used by GitHub CLI when working with codespaces:

- `GH_CS_SSH_COMMAND`: Overrides the default `ssh` command used by `gh cs ssh`
- `GH_CS_SCP_COMMAND`: Overrides the default `scp` command used by `gh cs cp`

## Problem

Currently, the Codespaces SSH and CP commands are hardcoded to use the first `ssh` and `scp` executables on your PATH. This makes it difficult for users who need to use alternative SSH/SCP implementations or want to pass custom parameters.

## Solution

This PR implements environment variables similar to Git's `GIT_SSH_COMMAND` that allow users to override these commands. The implementation:

- Allows specifying an alternative executable path
- Supports including command-line arguments in the variable value
- Handles paths with spaces when quoted properly
- Updates command documentation to explain the feature

Example usage:

```bash
# Use a different SSH implementation
GH_CS_SSH_COMMAND="/path/to/custom/ssh" gh cs ssh

# Use SSH with specific options 
GH_CS_SSH_COMMAND="ssh -v -o ProxyCommand=none" gh cs ssh

# Use a path with spaces
GH_CS_SSH_COMMAND="/Applications/Custom SSH.app/ssh" gh cs ssh 

# Similar usage for SCP command
GH_CS_SCP_COMMAND="/path/to/custom/scp -l 8192" gh cs cp file.txt remote:~/
```

## Implementation details

- Added command lookup via environment variables in `internal/codespaces/ssh.go`
- Used the existing `github.com/google/shlex` library for robust command parsing
- Refactored common code into a helper function to reduce duplication
- Added comprehensive test coverage for all new functionality

## Testing

Added several test types to verify the implementation:
- Unit tests for command parsing
- Tests for the helper functions
- Integration tests for both SSH and SCP commands

Manual testing was also performed with real codespaces to ensure proper functionality.