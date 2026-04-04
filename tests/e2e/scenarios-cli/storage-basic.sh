#!/bin/bash
# storage-basic.sh — CLI tests for `pinchtab storage` commands.

GROUP_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${GROUP_DIR}/../helpers/cli.sh"

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab storage set writes a localStorage item"

pt_cli storage set pt_cli_key pt_cli_value --type local
assert_cli_ok "set local item"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab storage get reads back the item"

pt_cli storage get --type local --key pt_cli_key
assert_cli_ok "get local item"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab storage set writes a sessionStorage item"

pt_cli storage set pt_sess_key pt_sess_value --type session
assert_cli_ok "set session item"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab storage delete removes a key"

pt_cli storage delete --key pt_cli_key --type local
assert_cli_ok "delete local key"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab storage clear --all clears both stores"

pt_cli storage clear --all
assert_cli_ok "clear --all"

end_test
