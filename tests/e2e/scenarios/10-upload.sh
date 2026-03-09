#!/bin/bash
# 10-upload.sh — File upload

source "$(dirname "$0")/common.sh"

# Navigate to upload test page
pt_post /navigate -d "{\"url\":\"${FIXTURES_URL}/upload.html\"}"
sleep 1

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab upload (base64 file)"

# Create a simple base64-encoded test file
# "Hello from E2E test!" in base64
FILE_CONTENT="SGVsbG8gZnJvbSBFMkUgdGVzdCE="

pt_post /upload -d "{\"selector\":\"#single-file\",\"files\":[{\"name\":\"test.txt\",\"content\":\"${FILE_CONTENT}\"}]}"
assert_ok "upload base64"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab upload (multiple files)"

pt_post /upload -d "{\"selector\":\"#multi-file\",\"files\":[{\"name\":\"file1.txt\",\"content\":\"${FILE_CONTENT}\"},{\"name\":\"file2.txt\",\"content\":\"${FILE_CONTENT}\"}]}"
assert_ok "upload multiple"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab upload --tab <id>"

pt_get /tabs
TAB_ID=$(get_first_tab)

pt_post "/tabs/${TAB_ID}/upload" -d "{\"selector\":\"#image-upload\",\"files\":[{\"name\":\"test.png\",\"content\":\"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==\"}]}"
assert_ok "tab upload"

end_test
