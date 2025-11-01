#!/usr/bin/env bash
# Simple installer wrapper that runs `go install` and optionally appends GOPATH/bin to the user's zshrc

set -euo pipefail

GOBIN_DIR=$(go env GOPATH)/bin

echo "Installing gitfit to ${GOBIN_DIR} (using 'go install')..."
go install ./cmd/gitfit

echo
echo "Installed gitfit to ${GOBIN_DIR}."
echo "If ${GOBIN_DIR} is not on your PATH you can add it to your ~/.zshrc with:"
echo
echo "  export PATH=\"${GOBIN_DIR}:\$PATH\""
echo
echo "To append automatically, re-run this script with --add-path"

if [ "${1:-}" = "--add-path" ]; then
  SHELL_RC=${ZDOTDIR:-$HOME}/.zshrc
  if grep -q "${GOBIN_DIR}" "$SHELL_RC" 2>/dev/null; then
    echo "${GOBIN_DIR} already appears in $SHELL_RC"
  else
    echo "Adding export to $SHELL_RC"
    printf "\n# added by git-fit installer\nexport PATH=\"%s:\$PATH\"\n" "${GOBIN_DIR}" >> "$SHELL_RC"
    echo "Appended PATH entry to $SHELL_RC. Reload your shell or run: source $SHELL_RC"
  fi
fi

echo "Done. Run 'gitfit --help' to see usage." 
