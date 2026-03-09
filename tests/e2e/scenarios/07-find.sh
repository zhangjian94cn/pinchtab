#!/bin/bash
# 07-find.sh — Element finding with semantic search

source "$(dirname "$0")/common.sh"

# Navigate to find test page
pt_post /navigate -d "{\"url\":\"${FIXTURES_URL}/find.html\"}"
sleep 1

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab find (login button)"

pt_post /find -d '{"query":"login button"}'
assert_ok "find login"

# Verify we got a ref back
if ! echo "$LAST_BODY" | jq -e '.best_ref' >/dev/null 2>&1; then
  fail "expected best_ref in response"
fi

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab find (email input)"

pt_post /find -d '{"query":"email input field"}'
assert_ok "find email"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab find (delete button)"

pt_post /find -d '{"query":"delete account button","threshold":0.2}'
assert_ok "find delete"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab find (search)"

pt_post /find -d '{"query":"search input","topK":5}'
assert_ok "find search"

# Verify candidates array
if ! echo "$LAST_BODY" | jq -e '.candidates | length > 0' >/dev/null 2>&1; then
  fail "expected candidates in response"
fi

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab find --tab <id>"

# Open find page in new tab to get known tab ID
pt_post /navigate -d "{\"url\":\"${FIXTURES_URL}/find.html\",\"newTab\":true}"
sleep 1

pt_get /tabs
TAB_ID=$(get_last_tab)

pt_post "/tabs/${TAB_ID}/find" -d '{"query":"sign up link"}'
assert_ok "tab find"

end_test
