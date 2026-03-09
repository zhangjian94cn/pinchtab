#!/bin/bash
# 09-download.sh — File download

source "$(dirname "$0")/common.sh"

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab download (text file)"

# Download the sample.txt file from fixtures server
pt_get "/download?url=${FIXTURES_URL}/sample.txt"
assert_ok "download text"

# Verify we got content back (base64 encoded by default)
if ! echo "$LAST_BODY" | jq -e '.content' >/dev/null 2>&1; then
  fail "expected content in response"
fi

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab download (raw mode)"

pt_get "/download?url=${FIXTURES_URL}/sample.txt&raw=true"
assert_ok "download raw"

# In raw mode, body should be the file content directly
if ! echo "$LAST_BODY" | grep -q "PinchTab E2E Test Suite"; then
  fail "expected file content in raw response"
fi

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab download (HTML page)"

pt_get "/download?url=${FIXTURES_URL}/index.html"
assert_ok "download html"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab download --tab <id>"

pt_get /tabs
TAB_ID=$(get_first_tab)

pt_get "/tabs/${TAB_ID}/download?url=${FIXTURES_URL}/sample.txt"
assert_ok "tab download"

end_test
