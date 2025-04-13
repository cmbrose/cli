# Add support for custom SSH and SCP commands via environment variables

## Description

This PR adds support for customizing the SSH and SCP commands used by the GitHub CLI's codespace functionality through environment variables. Users can now specify alternative SSH/SCP commands with additional options, providing greater flexibility and control over their codespace connections.

## Features Added

- Added support for two new environment variables:
  - `GH_CS_SSH_COMMAND`: Allows customization of the SSH command used for `gh cs ssh`
  - `GH_CS_SCP_COMMAND`: Allows customization of the SCP command used for `gh cs cp`

- Implemented robust command-line parsing using the `github.com/google/shlex` library to handle quoted arguments and complex command structures

- Modified the newSSHCommand and newSCPCommand functions to incorporate arguments from the environment variables

## Implementation Details

- The implementation uses the `shlex` library to parse command lines with quoted arguments, ensuring proper handling of complex command structures
- Custom arguments from environment variables are inserted at the beginning of the command arguments list
- Comprehensive unit tests were added to verify the functionality with various command formats

## Example Usage

```bash
# Use SSH with verbose output
GH_CS_SSH_COMMAND="ssh -v" gh cs ssh -c my-codespace

# Use SSH with custom options
GH_CS_SSH_COMMAND="ssh -o LogLevel=DEBUG" gh cs ssh -c my-codespace

# Use a custom SSH binary with options
GH_CS_SSH_COMMAND="/path/to/custom/ssh -v -o IdentityFile=/custom/key" gh cs ssh -c my-codespace

# Similar usage for SCP commands
GH_CS_SCP_COMMAND="scp -v" gh cs cp -e README.md remote:/tmp/
```

## Testing

The implementation has been thoroughly tested with:

1. Unit tests that verify the parsing of environment variables with various command formats
2. Manual testing with real codespaces, confirming that both SSH and SCP commands work as expected

## Dependencies

- Added `github.com/google/shlex` for parsing command line arguments

## Security Considerations

The implementation ensures that only the command path and arguments are extracted from the environment variables, preventing the injection of unwanted or harmful arguments.
