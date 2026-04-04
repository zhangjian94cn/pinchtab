#!/bin/bash
# state-basic.sh — CLI tests for `pinchtab state` commands.

GROUP_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${GROUP_DIR}/../helpers/cli.sh"

STATE_NAME="cli-e2e-state-$(date +%s)"

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab state list shows saved states"

pt_cli state list
assert_cli_ok "list states"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab state save captures current browser state"

pt_cli state save --name "$STATE_NAME"
assert_cli_ok "save state"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab state show displays state details"

pt_cli state show --name "$STATE_NAME"
assert_cli_ok "show state"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab state load restores saved state (exact name)"

pt_cli state load --name "$STATE_NAME"
assert_cli_ok "load exact name"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab state load restores saved state (prefix)"

PREFIX=$(echo "$STATE_NAME" | cut -c1-8)
pt_cli state load --name "$PREFIX"
assert_cli_ok "load by prefix"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab state delete removes the saved state"

pt_cli state delete --name "$STATE_NAME"
assert_cli_ok "delete state"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab state clean runs without error"

pt_cli state clean --older-than 9999
assert_cli_ok "clean old states"

end_test
