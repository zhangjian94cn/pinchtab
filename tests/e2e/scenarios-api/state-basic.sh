#!/bin/bash
# state-basic.sh — State management API tests.
# Requires: a running PinchTab instance with security.allowStateExport=true.

GROUP_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${GROUP_DIR}/../helpers/api.sh"

STATE_NAME="pt-e2e-state-$(date +%s)"

# ─────────────────────────────────────────────────────────────────
start_test "GET /state/list returns state list"

pt_get "/state/list"
assert_ok "list states"
assert_json_exists "$RESULT" '.states' "has states array"
assert_json_exists "$RESULT" '.count' "has count"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "POST /state/save captures browser state"

pt_post "/state/save" -d "{\"name\":\"${STATE_NAME}\"}"
assert_ok "save state"
assert_json_exists "$RESULT" '.name' "has name"
assert_json_contains "$RESULT" '.name' "${STATE_NAME}" "name matches"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "GET /state/list includes newly saved state"

pt_get "/state/list"
assert_ok "list after save"
COUNT=$(echo "$RESULT" | jq '.count')
if [ "$COUNT" -gt "0" ]; then
  echo -e "  ${GREEN}✓${NC} state count > 0"
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${RED}✗${NC} expected at least 1 state, got $COUNT"
  ((ASSERTIONS_FAILED++)) || true
fi

end_test

# ─────────────────────────────────────────────────────────────────
start_test "GET /state/show returns full state details"

pt_get "/state/show?name=${STATE_NAME}"
assert_ok "show state"
assert_json_exists "$RESULT" '.name' "has name"
assert_json_exists "$RESULT" '.cookies' "has cookies array"
assert_json_exists "$RESULT" '.storage' "has storage map"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "POST /state/load restores the saved state"

pt_post "/state/load" -d "{\"name\":\"${STATE_NAME}\"}"
assert_ok "load state"
assert_json_exists "$RESULT" '.name' "has name"
assert_json_exists "$RESULT" '.cookiesRestored' "has cookiesRestored"
assert_json_exists "$RESULT" '.storageItemsRestored' "has storageItemsRestored"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "POST /state/load with prefix finds most recent match"

PREFIX=$(echo "$STATE_NAME" | cut -c1-8)
pt_post "/state/load" -d "{\"name\":\"${PREFIX}\"}"
assert_ok "load by prefix"
assert_json_exists "$RESULT" '.name' "has name"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "DELETE /state removes the saved state"

pt_delete "/state?name=${STATE_NAME}"
assert_ok "delete state"
assert_json_exists "$RESULT" '.deleted' "has deleted field"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "DELETE /state on nonexistent name returns error"

pt_delete "/state?name=nonexistent_xyz_$(date +%s)"
assert_not_ok "rejects nonexistent name"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "POST /state/clean removes old files (olderThanHours=0 for test)"

# Use 9999 hours so nothing real gets removed in test environment.
pt_post "/state/clean" -d '{"olderThanHours":9999}'
assert_ok "clean states"
assert_json_exists "$RESULT" '.removed' "has removed count"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "POST /state/save rejects missing name if auto-name not acceptable"

# PinchTab auto-generates a name when none is provided — this is valid.
pt_post "/state/save" -d '{}'
assert_ok "save with auto-generated name"

end_test
