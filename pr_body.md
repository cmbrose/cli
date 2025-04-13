# Add support for custom SSH and SCP commands via environment variables

This PR adds support for customizing the SSH and SCP commands used by the GitHub CLI when connecting to codespaces. This is similar to Git's `GIT_SSH_COMMAND` environment variable, but specifically for GitHub Codespaces.

## Changes

- Added support for `GH_CS_SSH_COMMAND` environment variable to specify an alternative SSH command
- Added support for `GH_CS_SCP_COMMAND` environment variable to specify an alternative SCP command
- Added comprehensive test coverage for both environment variables
- Added documentation in help text for both commands

## Use Cases

This change allows users to:
- Use custom SSH implementations (e.g., `mosh`, `openssh-client`)
- Add default SSH options (e.g., `ssh -o StrictHostKeyChecking=no`)
- Use SSH wrappers or proxies
- Customize SCP behavior for file transfers

## Examples

```bash
# Use a custom SSH command with options
export GH_CS_SSH_COMMAND="/usr/local/bin/ssh -v"
gh codespace ssh

# Use a custom SCP command with options
export GH_CS_SCP_COMMAND="/usr/local/bin/scp -C"
gh codespace cp local/file.txt remote:/workspaces/repo/

# Use paths with spaces
export GH_CS_SSH_COMMAND='"/Applications/My SSH/ssh" -o StrictHostKeyChecking=no'
gh codespace ssh
```

## Testing

- Added unit tests for both environment variables
- Tested with various command formats and options
- Tested with paths containing spaces
- Tested with quoted paths and arguments

## Documentation

Added help text to both commands explaining the new environment variables:

```bash
$ gh codespace ssh --help
...
You can use the GH_CS_SSH_COMMAND environment variable to specify an alternative SSH command
to use instead of the default ssh command.
...

$ gh codespace cp --help
...
You can use the GH_CS_SCP_COMMAND environment variable to specify an alternative SCP command
to use instead of the default scp command.
...
```

## Related Issues

Fixes #XXXX (if there's a related issue)

## Notes

- The implementation uses the `github.com/google/shlex` package for robust command-line parsing
- Both environment variables support full command lines with arguments
- Paths with spaces are handled correctly through proper quoting
