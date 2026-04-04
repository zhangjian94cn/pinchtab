package actions

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/pinchtab/pinchtab/internal/cli/apiclient"
	"github.com/spf13/cobra"
)

// StorageGet retrieves localStorage and/or sessionStorage items for the active tab.
func StorageGet(client *http.Client, base, token string, cmd *cobra.Command) {
	params := url.Values{}
	if t, _ := cmd.Flags().GetString("type"); t != "" {
		params.Set("type", t)
	}
	if k, _ := cmd.Flags().GetString("key"); k != "" {
		params.Set("key", k)
	}
	if tab, _ := cmd.Flags().GetString("tab"); tab != "" {
		params.Set("tabId", tab)
	}

	result := apiclient.DoGetRaw(client, base, token, "/storage", params)
	if result == nil {
		fmt.Fprintln(os.Stderr, "Failed to get storage")
		os.Exit(1)
	}

	var buf map[string]any
	if err := json.Unmarshal(result, &buf); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse response: %v\n", err)
		os.Exit(1)
	}

	out, _ := json.MarshalIndent(buf, "", "  ")
	fmt.Println(string(out))
}

// StorageSet sets a single localStorage or sessionStorage item.
func StorageSet(client *http.Client, base, token string, cmd *cobra.Command, key, value string) {
	storageType, _ := cmd.Flags().GetString("type")
	if storageType == "" {
		storageType = "local"
	}
	tabID, _ := cmd.Flags().GetString("tab")

	body := map[string]any{
		"key":   key,
		"value": value,
		"type":  storageType,
	}
	if tabID != "" {
		body["tabId"] = tabID
	}

	result := apiclient.DoPost(client, base, token, "/storage", body)
	if result == nil {
		fmt.Fprintln(os.Stderr, "Failed to set storage item")
		os.Exit(1)
	}
}

// StorageDelete removes a storage item, clears a store, or clears both (--all).
// It calls DELETE /storage so the server-side delete/clear handler is used.
func StorageDelete(client *http.Client, base, token string, cmd *cobra.Command) {
	storageType, _ := cmd.Flags().GetString("type")
	key, _ := cmd.Flags().GetString("key")
	all, _ := cmd.Flags().GetBool("all")
	tabID, _ := cmd.Flags().GetString("tab")

	if all {
		storageType = "all"
	}
	if storageType == "" {
		storageType = "local"
	}

	body := map[string]any{
		"type": storageType,
	}
	if key != "" {
		body["key"] = key
	}
	if tabID != "" {
		body["tabId"] = tabID
	}

	result := apiclient.DoDeleteJSON(client, base, token, "/storage", body)
	if result == nil {
		fmt.Fprintln(os.Stderr, "Failed to delete storage")
		os.Exit(1)
	}
}

// StorageClear clears storage (alias: passes type=all or the given type).
func StorageClear(client *http.Client, base, token string, cmd *cobra.Command) {
	StorageDelete(client, base, token, cmd)
}
