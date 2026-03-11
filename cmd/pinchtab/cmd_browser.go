package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// --- navigate ---

func cliNavigate(client *http.Client, base, token string, args []string) {
	if len(args) < 1 {
		fatal("Usage: pinchtab nav <url> [--new-tab] [--block-images] [--block-ads]")
	}
	body := map[string]any{"url": args[0]}
	for _, a := range args[1:] {
		switch a {
		case "--new-tab":
			body["newTab"] = true
		case "--block-images":
			body["blockImages"] = true
		case "--block-ads":
			body["blockAds"] = true
		}
	}
	result := doPost(client, base, token, "/navigate", body)
	suggestNextAction("navigate", result)
}

// --- snapshot ---

func cliSnapshot(client *http.Client, base, token string, args []string) {
	params := url.Values{}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--interactive", "-i":
			params.Set("filter", "interactive")
		case "--compact", "-c":
			params.Set("format", "compact")
		case "--text":
			params.Set("format", "text")
		case "--diff", "-d":
			params.Set("diff", "true")
		case "--selector", "-s":
			if i+1 < len(args) {
				i++
				params.Set("selector", args[i])
			}
		case "--max-tokens":
			if i+1 < len(args) {
				i++
				params.Set("maxTokens", args[i])
			}
		case "--depth":
			if i+1 < len(args) {
				i++
				params.Set("depth", args[i])
			}
		case "--tab":
			if i+1 < len(args) {
				i++
				params.Set("tabId", args[i])
			}
		}
	}
	result := doGet(client, base, token, "/snapshot", params)
	suggestNextAction("snapshot", result)
}

// --- element actions ---

func cliAction(client *http.Client, base, token, kind string, args []string) {
	body := map[string]any{"kind": kind}

	switch kind {
	case "click", "hover", "focus":
		var cssSelector string
		var refArg string
		for i := 0; i < len(args); i++ {
			switch args[i] {
			case "--css":
				if i+1 < len(args) {
					i++
					cssSelector = args[i]
				}
			case "--wait-nav":
				body["waitNav"] = true
			default:
				if refArg == "" {
					refArg = args[i]
				}
			}
		}
		if cssSelector != "" {
			body["selector"] = cssSelector
		} else if refArg != "" {
			body["ref"] = refArg
		} else {
			fatal("Usage: pinchtab %s <ref> [--css <selector>] [--wait-nav]", kind)
		}
	case "type":
		if len(args) < 2 {
			fatal("Usage: pinchtab type <ref> <text>")
		}
		body["ref"] = args[0]
		body["text"] = strings.Join(args[1:], " ")
	case "fill":
		if len(args) < 2 {
			fatal("Usage: pinchtab fill <ref|selector> <text>")
		}
		if strings.HasPrefix(args[0], "e") {
			body["ref"] = args[0]
		} else {
			body["selector"] = args[0]
		}
		body["text"] = strings.Join(args[1:], " ")
	case "press":
		if len(args) < 1 {
			fatal("Usage: pinchtab press <key>  (e.g. Enter, Tab, Escape)")
		}
		body["key"] = args[0]
	case "scroll":
		if len(args) < 1 {
			fatal("Usage: pinchtab scroll <ref|pixels|direction>  (e.g. e5, 800, or down)")
		}
		if strings.HasPrefix(args[0], "e") {
			body["ref"] = args[0]
		} else if px, err := strconv.Atoi(args[0]); err == nil {
			body["scrollY"] = px
		} else {
			switch strings.ToLower(args[0]) {
			case "down":
				body["scrollY"] = 800
			case "up":
				body["scrollY"] = -800
			case "right":
				body["scrollX"] = 800
			case "left":
				body["scrollX"] = -800
			default:
				fatal("Usage: pinchtab scroll <ref|pixels|direction>  (e.g. e5, 800, or down)")
			}
		}
	case "select":
		if len(args) < 2 {
			fatal("Usage: pinchtab select <ref> <value>")
		}
		body["ref"] = args[0]
		body["value"] = args[1]
	}

	doPost(client, base, token, "/action", body)
}

// --- text ---

func cliText(client *http.Client, base, token string, args []string) {
	params := url.Values{}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--raw":
			params.Set("mode", "raw")
		case "--tab":
			if i+1 < len(args) {
				i++
				params.Set("tabId", args[i])
			}
		}
	}
	doGet(client, base, token, "/text", params)
}

// --- tabs ---

func cliTabs(client *http.Client, base, token string, args []string) {
	if len(args) == 0 {
		// List all tabs
		doGet(client, base, token, "/tabs", nil)
		return
	}

	cmd := args[0]
	subArgs := args[1:]

	// Check if this is a tab operation (navigate, snapshot, click, etc.)
	// Pattern: pinchtab tab <operation> <tabId> [args...]
	if isTabOperation(cmd) {
		cliTabOperation(client, base, token, cmd, subArgs)
		return
	}

	// Legacy: pinchtab tab new/close
	switch cmd {
	case "new":
		url := ""
		if len(subArgs) > 0 {
			url = subArgs[0]
		}

		// Check if any instances are running
		instances := getInstances(client, base, token)
		if len(instances) == 0 {
			fmt.Fprintln(os.Stderr, styleStderr(cliWarningStyle, "No instances running, launching default..."))
			launchInstance(client, base, token, "default")
			fmt.Fprintln(os.Stderr, styleStderr(cliSuccessStyle, "Instance launched"))
		}

		body := map[string]any{"action": "new"}
		if url != "" {
			body["url"] = url
		}
		doPost(client, base, token, "/tab", body)

	case "close":
		if len(subArgs) < 1 {
			fatal("Usage: pinchtab tab close <tabId>")
		}
		doPost(client, base, token, "/tab", map[string]any{
			"action": "close",
			"tabId":  subArgs[0],
		})

	default:
		cliTabOperation(client, base, token, cmd, subArgs)
	}
}

func isTabOperation(op string) bool {
	ops := map[string]bool{
		"navigate": true, "snapshot": true, "screenshot": true,
		"click": true, "type": true, "press": true, "fill": true,
		"hover": true, "scroll": true, "select": true, "focus": true,
		"text": true, "eval": true, "evaluate": true, "pdf": true,
		"cookies": true, "lock": true, "unlock": true, "locks": true,
		"fingerprint": true, "info": true,
	}
	return ops[op]
}

func cliTabOperation(client *http.Client, base, token string, op string, args []string) {
	if len(args) < 1 {
		fatal("Usage: pinchtab tab %s <tabId> [args...]", op)
	}

	tabID := args[0]
	restArgs := args[1:]

	switch op {
	case "navigate":
		if len(restArgs) < 1 {
			fatal("Usage: pinchtab tab navigate <tabId> <url> [--timeout N] [--block-images]")
		}
		body := map[string]any{"url": restArgs[0]}
		for i := 1; i < len(restArgs); i++ {
			switch restArgs[i] {
			case "--timeout":
				if i+1 < len(restArgs) {
					body["timeout"] = restArgs[i+1]
					i++
				}
			case "--block-images":
				body["blockImages"] = true
			case "--block-ads":
				body["blockAds"] = true
			}
		}
		doPost(client, base, token, fmt.Sprintf("/tabs/%s/navigate", tabID), body)

	case "snapshot":
		params := url.Values{}
		for _, arg := range restArgs {
			switch arg {
			case "-i", "--interactive":
				params.Set("interactive", "true")
			case "-c", "--compact":
				params.Set("compact", "true")
			case "-d", "--diff":
				params.Set("diff", "true")
			}
		}
		doGet(client, base, token, fmt.Sprintf("/tabs/%s/snapshot", tabID), params)

	case "screenshot", "ss":
		params := url.Values{}
		outFile := ""
		for i := 0; i < len(restArgs); i++ {
			switch restArgs[i] {
			case "-o", "--output":
				if i+1 < len(restArgs) {
					outFile = restArgs[i+1]
					i++
				}
			case "-q", "--quality":
				if i+1 < len(restArgs) {
					params.Set("quality", restArgs[i+1])
					i++
				}
			}
		}
		params.Set("raw", "true")
		data := doGetRaw(client, base, token, fmt.Sprintf("/tabs/%s/screenshot", tabID), params)
		if outFile == "" {
			outFile = fmt.Sprintf("screenshot-%s.png", time.Now().Format("20060102-150405"))
		}
		if data != nil {
			if err := os.WriteFile(outFile, data, 0600); err == nil {
				fmt.Println(styleStdout(cliSuccessStyle, fmt.Sprintf("Saved %s (%d bytes)", outFile, len(data))))
			}
		}

	case "click", "hover", "focus":
		if len(restArgs) < 1 {
			fatal("Usage: pinchtab tab %s <tabId> <ref>", op)
		}
		body := map[string]any{"kind": op, "ref": restArgs[0]}
		doPost(client, base, token, fmt.Sprintf("/tabs/%s/action", tabID), body)

	case "type":
		if len(restArgs) < 2 {
			fatal("Usage: pinchtab tab type <tabId> <ref> <text>")
		}
		body := map[string]any{"kind": "type", "ref": restArgs[0], "text": strings.Join(restArgs[1:], " ")}
		doPost(client, base, token, fmt.Sprintf("/tabs/%s/action", tabID), body)

	case "fill":
		if len(restArgs) < 2 {
			fatal("Usage: pinchtab tab fill <tabId> <ref> <text>")
		}
		body := map[string]any{"kind": "fill", "ref": restArgs[0], "text": strings.Join(restArgs[1:], " ")}
		doPost(client, base, token, fmt.Sprintf("/tabs/%s/action", tabID), body)

	case "press":
		if len(restArgs) < 1 {
			fatal("Usage: pinchtab tab press <tabId> <key>")
		}
		body := map[string]any{"kind": "press", "key": restArgs[0]}
		doPost(client, base, token, fmt.Sprintf("/tabs/%s/action", tabID), body)

	case "scroll":
		if len(restArgs) < 1 {
			fatal("Usage: pinchtab tab scroll <tabId> <direction|pixels>")
		}
		body := map[string]any{}
		if v, err := strconv.Atoi(restArgs[0]); err == nil {
			body["kind"] = "scroll"
			body["scrollY"] = v
		} else {
			body["kind"] = "scroll"
			body["direction"] = restArgs[0]
		}
		doPost(client, base, token, fmt.Sprintf("/tabs/%s/action", tabID), body)

	case "select":
		if len(restArgs) < 2 {
			fatal("Usage: pinchtab tab select <tabId> <ref> <value>")
		}
		body := map[string]any{"kind": "select", "ref": restArgs[0], "value": restArgs[1]}
		doPost(client, base, token, fmt.Sprintf("/tabs/%s/action", tabID), body)

	case "text":
		params := url.Values{}
		for _, arg := range restArgs {
			if arg == "--raw" {
				params.Set("raw", "true")
			}
		}
		doGet(client, base, token, fmt.Sprintf("/tabs/%s/text", tabID), params)

	case "eval", "evaluate":
		if len(restArgs) < 1 {
			fatal("Usage: pinchtab tab eval <tabId> <expression>")
		}
		body := map[string]any{"expression": strings.Join(restArgs, " ")}
		doPost(client, base, token, fmt.Sprintf("/tabs/%s/evaluate", tabID), body)

	case "pdf":
		params := url.Values{}
		outFile := ""
		for i := 0; i < len(restArgs); i++ {
			switch restArgs[i] {
			case "-o", "--output":
				if i+1 < len(restArgs) {
					outFile = restArgs[i+1]
					i++
				}
			case "--landscape":
				params.Set("landscape", "true")
			case "--scale":
				if i+1 < len(restArgs) {
					params.Set("scale", restArgs[i+1])
					i++
				}
			}
		}
		params.Set("raw", "true")
		data := doGetRaw(client, base, token, fmt.Sprintf("/tabs/%s/pdf", tabID), params)
		if outFile == "" {
			outFile = fmt.Sprintf("page-%s.pdf", time.Now().Format("20060102-150405"))
		}
		if data != nil {
			if err := os.WriteFile(outFile, data, 0600); err != nil {
				fmt.Printf("Saved %s (%d bytes)\n", outFile, len(data))
			}
		}

	case "cookies":
		doGet(client, base, token, fmt.Sprintf("/tabs/%s/cookies", tabID), nil)

	case "lock":
		body := map[string]any{}
		for i := 0; i < len(restArgs); i++ {
			switch restArgs[i] {
			case "--owner":
				if i+1 < len(restArgs) {
					body["owner"] = restArgs[i+1]
					i++
				}
			case "--ttl":
				if i+1 < len(restArgs) {
					if ttl, err := strconv.Atoi(restArgs[i+1]); err == nil {
						body["ttl"] = ttl
					}
					i++
				}
			}
		}
		doPost(client, base, token, fmt.Sprintf("/tabs/%s/lock", tabID), body)

	case "unlock":
		body := map[string]any{}
		for i := 0; i < len(restArgs); i++ {
			switch restArgs[i] {
			case "--owner":
				if i+1 < len(restArgs) {
					body["owner"] = restArgs[i+1]
					i++
				}
			}
		}
		doPost(client, base, token, fmt.Sprintf("/tabs/%s/unlock", tabID), body)

	case "locks":
		doGet(client, base, token, fmt.Sprintf("/tabs/%s/locks", tabID), nil)

	case "info":
		doGet(client, base, token, fmt.Sprintf("/tabs/%s", tabID), nil)

	default:
		fatal("Unknown tab operation: %s", op)
	}
}

// getInstances fetches the list of running instances
func getInstances(client *http.Client, base, token string) []map[string]any {
	resp, err := http.NewRequest("GET", base+"/instances", nil)
	if err != nil {
		return nil
	}
	if token != "" {
		resp.Header.Set("Authorization", "Bearer "+token)
	}

	result, err := client.Do(resp)
	if err != nil || result.StatusCode >= 400 {
		return nil
	}
	defer func() { _ = result.Body.Close() }()

	var data map[string]any
	if err := json.NewDecoder(result.Body).Decode(&data); err != nil {
		log.Printf("warning: error decoding instances response: %v", err)
	}

	if instances, ok := data["instances"].([]interface{}); ok {
		converted := make([]map[string]any, len(instances))
		for i, inst := range instances {
			if m, ok := inst.(map[string]any); ok {
				converted[i] = m
			}
		}
		return converted
	}
	return nil
}

// launchInstance launches a default instance
func launchInstance(client *http.Client, base, token string, profile string) {
	body := map[string]any{"profile": profile}
	doPost(client, base, token, "/instances/launch", body)
}

// --- screenshot ---

func cliScreenshot(client *http.Client, base, token string, args []string) {
	params := url.Values{}
	params.Set("raw", "true")
	outFile := ""
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-o", "--output":
			if i+1 < len(args) {
				i++
				outFile = args[i]
			}
		case "--quality", "-q":
			if i+1 < len(args) {
				i++
				params.Set("quality", args[i])
			}
		case "--tab":
			if i+1 < len(args) {
				i++
				params.Set("tabId", args[i])
			}
		}
	}

	if outFile == "" {
		outFile = fmt.Sprintf("screenshot-%s.jpg", time.Now().Format("20060102-150405"))
	}

	data := doGetRaw(client, base, token, "/screenshot", params)
	if data == nil {
		return
	}
	if err := os.WriteFile(outFile, data, 0600); err != nil {
		fatal("Write failed: %v", err)
	}
	fmt.Println(styleStdout(cliSuccessStyle, fmt.Sprintf("Saved %s (%d bytes)", outFile, len(data))))
}

// --- evaluate ---

func cliEvaluate(client *http.Client, base, token string, args []string) {
	if len(args) < 1 {
		fatal("Usage: pinchtab eval <expression>")
	}
	expr := strings.Join(args, " ")
	doPost(client, base, token, "/evaluate", map[string]any{
		"expression": expr,
	})
}

// --- pdf ---

func cliPDF(client *http.Client, base, token string, args []string) {
	params := url.Values{}
	params.Set("raw", "true")
	outFile := ""
	tabID := ""
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-o", "--output":
			if i+1 < len(args) {
				i++
				outFile = args[i]
			}
		case "--landscape":
			params.Set("landscape", "true")
		case "--scale":
			if i+1 < len(args) {
				i++
				params.Set("scale", args[i])
			}
		case "--tab":
			if i+1 < len(args) {
				i++
				tabID = args[i]
			}
		// Paper dimensions
		case "--paper-width":
			if i+1 < len(args) {
				i++
				params.Set("paperWidth", args[i])
			}
		case "--paper-height":
			if i+1 < len(args) {
				i++
				params.Set("paperHeight", args[i])
			}
		// Margins
		case "--margin-top":
			if i+1 < len(args) {
				i++
				params.Set("marginTop", args[i])
			}
		case "--margin-bottom":
			if i+1 < len(args) {
				i++
				params.Set("marginBottom", args[i])
			}
		case "--margin-left":
			if i+1 < len(args) {
				i++
				params.Set("marginLeft", args[i])
			}
		case "--margin-right":
			if i+1 < len(args) {
				i++
				params.Set("marginRight", args[i])
			}
		// Content options
		case "--page-ranges":
			if i+1 < len(args) {
				i++
				params.Set("pageRanges", args[i])
			}
		case "--prefer-css-page-size":
			params.Set("preferCSSPageSize", "true")
		// Header/Footer
		case "--display-header-footer":
			params.Set("displayHeaderFooter", "true")
		case "--header-template":
			if i+1 < len(args) {
				i++
				params.Set("headerTemplate", args[i])
			}
		case "--footer-template":
			if i+1 < len(args) {
				i++
				params.Set("footerTemplate", args[i])
			}
		// Accessibility
		case "--generate-tagged-pdf":
			params.Set("generateTaggedPDF", "true")
		case "--generate-document-outline":
			params.Set("generateDocumentOutline", "true")
		// Output options
		case "--file-output":
			params.Del("raw")
			params.Set("output", "file")
		case "--path":
			if i+1 < len(args) {
				i++
				params.Set("path", args[i])
			}
		case "--raw":
			params.Set("raw", "true")
		}
	}

	if outFile == "" {
		outFile = fmt.Sprintf("page-%s.pdf", time.Now().Format("20060102-150405"))
	}

	var data []byte
	if tabID != "" {
		data = doGetRaw(client, base, token, fmt.Sprintf("/tabs/%s/pdf", tabID), params)
	} else {
		data = doGetRaw(client, base, token, "/pdf", params)
	}
	if data == nil {
		return
	}
	if err := os.WriteFile(outFile, data, 0600); err != nil {
		fatal("Write failed: %v", err)
	}
	fmt.Println(styleStdout(cliSuccessStyle, fmt.Sprintf("Saved %s (%d bytes)", outFile, len(data))))
}

// --- quick command ---

func cliQuick(client *http.Client, base, token string, args []string) {
	if len(args) < 1 {
		fatal("Usage: pinchtab quick <url>")
	}

	fmt.Println(styleStdout(cliHeadingStyle, fmt.Sprintf("Navigating to %s...", args[0])))

	// Navigate
	navBody := map[string]any{"url": args[0]}
	navResult := doPost(client, base, token, "/navigate", navBody)

	// Small delay for page to stabilize
	time.Sleep(1 * time.Second)

	fmt.Println()
	fmt.Println(styleStdout(cliHeadingStyle, "Page structure"))

	// Snapshot with interactive filter
	snapParams := url.Values{}
	snapParams.Set("filter", "interactive")
	snapParams.Set("compact", "true")
	doGet(client, base, token, "/snapshot", snapParams)

	// Extract info from navigation result
	if title, ok := navResult["title"].(string); ok {
		fmt.Println()
		fmt.Printf("%s %s\n", styleStdout(cliMutedStyle, "Title:"), styleStdout(cliValueStyle, title))
	}
	if urlStr, ok := navResult["url"].(string); ok {
		fmt.Printf("%s %s\n", styleStdout(cliMutedStyle, "URL:"), styleStdout(cliValueStyle, urlStr))
	}

	fmt.Println()
	fmt.Println(styleStdout(cliHeadingStyle, "Quick actions"))
	fmt.Printf("  %s %s\n", styleStdout(cliCommandStyle, "pinchtab click <ref>"), styleStdout(cliMutedStyle, "# Click an element (use refs from above)"))
	fmt.Printf("  %s %s\n", styleStdout(cliCommandStyle, "pinchtab type <ref> <text>"), styleStdout(cliMutedStyle, "# Type into input field"))
	fmt.Printf("  %s %s\n", styleStdout(cliCommandStyle, "pinchtab screenshot"), styleStdout(cliMutedStyle, "# Take a screenshot"))
	fmt.Printf("  %s %s\n", styleStdout(cliCommandStyle, "pinchtab pdf --tab <id> -o output.pdf"), styleStdout(cliMutedStyle, "# Save tab as PDF"))
}

// --- health ---

func cliHealth(client *http.Client, base, token string) {
	doGet(client, base, token, "/health", nil)
}

// --- instance ---

func cliInstance(client *http.Client, base, token string, args []string) {
	if len(args) < 1 {
		fatal("Usage: pinchtab instance <subcommand> [options]\nSubcommands: start, launch (alias), navigate, logs, stop")
	}

	subCmd := args[0]
	subArgs := args[1:]

	switch subCmd {
	case "start", "launch": // "start" is new Phase 2 API, "launch" is legacy
		cliInstanceStart(client, base, token, subArgs)
	case "navigate":
		cliInstanceNavigate(client, base, token, subArgs)
	case "logs":
		cliInstanceLogs(client, base, token, subArgs)
	case "stop":
		cliInstanceStop(client, base, token, subArgs)
	default:
		fatal("Unknown subcommand: %s", subCmd)
	}
}

func cliInstanceStart(client *http.Client, base, token string, args []string) {
	body := map[string]any{}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--profileId":
			if i+1 < len(args) {
				body["profileId"] = args[i+1]
				i++
			}
		case "--mode":
			if i+1 < len(args) {
				body["mode"] = args[i+1]
				i++
			}
		case "--port":
			if i+1 < len(args) {
				body["port"] = args[i+1]
				i++
			}
		}
	}

	// Use new /instances/start endpoint if available, fall back to /instances/launch for backward compat
	endpoint := "/instances/start"
	doPost(client, base, token, endpoint, body)
}

func cliInstanceNavigate(client *http.Client, base, token string, args []string) {
	if len(args) < 2 {
		fatal("Usage: pinchtab instance navigate <instance-id> <url>")
	}

	instID := args[0]
	targetURL := args[1]

	// Instance navigate now works via tab-scoped navigation:
	// open a tab on the instance, then navigate that tab.
	openResp := doPost(client, base, token, fmt.Sprintf("/instances/%s/tabs/open", instID), map[string]any{
		"url": "about:blank",
	})
	tabID, _ := openResp["tabId"].(string)
	if tabID == "" {
		fatal("failed to open tab for instance %s", instID)
	}

	// doPost auto-prints JSON response.
	doPost(client, base, token, fmt.Sprintf("/tabs/%s/navigate", tabID), map[string]any{
		"url": targetURL,
	})
}

func cliInstanceLogs(client *http.Client, base, token string, args []string) {
	var instID string

	// Support both positional argument and --id flag
	if len(args) == 0 {
		fatal("Usage: pinchtab instance logs <instance-id> OR pinchtab instance logs --id <instance-id>")
	}

	// Check if first arg is --id flag
	if args[0] == "--id" {
		if len(args) < 2 {
			fatal("Usage: --id requires instance ID")
		}
		instID = args[1]
	} else {
		// Positional argument (backward compat)
		instID = args[0]
	}

	logs := doGetRaw(client, base, token, fmt.Sprintf("/instances/%s/logs", instID), nil)
	fmt.Println(string(logs))
}

func cliInstanceStop(client *http.Client, base, token string, args []string) {
	var instID string

	// Support both positional argument and --id flag
	if len(args) == 0 {
		fatal("Usage: pinchtab instance stop <instance-id> OR pinchtab instance stop --id <instance-id>")
	}

	// Check if first arg is --id flag
	if args[0] == "--id" {
		if len(args) < 2 {
			fatal("Usage: --id requires instance ID")
		}
		instID = args[1]
	} else {
		// Positional argument (backward compat)
		instID = args[0]
	}

	// doPost auto-prints JSON response
	doPost(client, base, token, fmt.Sprintf("/instances/%s/stop", instID), nil)
}

// --- instances ---

func cliInstances(client *http.Client, base, token string) {
	body := doGetRaw(client, base, token, "/instances", nil)

	// Parse and format as JSON
	var instances []map[string]any
	if err := json.Unmarshal(body, &instances); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse instances: %v\n", err)
		os.Exit(1)
	}

	// Transform to cleaner output format
	output := make([]map[string]any, len(instances))
	for i, inst := range instances {
		id, _ := inst["id"].(string)
		port, _ := inst["port"].(string)
		headless, _ := inst["headless"].(bool)
		status, _ := inst["status"].(string)

		mode := "headless"
		if !headless {
			mode = "headed"
		}

		output[i] = map[string]any{
			"id":     id,
			"port":   port,
			"mode":   mode,
			"status": status,
		}
	}

	// Output as JSON
	data, _ := json.MarshalIndent(output, "", "  ")
	fmt.Println(string(data))
}

// --- profiles ---

func cliProfiles(client *http.Client, base, token string) {
	result := doGet(client, base, token, "/profiles", nil)

	// Display profiles in a friendly format
	if profiles, ok := result["profiles"].([]interface{}); ok && len(profiles) > 0 {
		fmt.Println()
		for _, prof := range profiles {
			if m, ok := prof.(map[string]any); ok {
				name, _ := m["name"].(string)

				fmt.Printf("👤 %s\n", name)
			}
		}
		fmt.Println()
	} else {
		fmt.Println("No profiles available")
	}
}

// --- helpers ---

func doGet(client *http.Client, base, token, path string, params url.Values) map[string]any {
	u := base + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	req, _ := http.NewRequest("GET", u, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := client.Do(req)
	if err != nil {
		fatal("Request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		fmt.Fprintln(os.Stderr, styleStderr(cliErrorStyle, fmt.Sprintf("Error %d: %s", resp.StatusCode, string(body))))
		os.Exit(1)
	}

	// Pretty-print JSON if possible
	var buf bytes.Buffer
	if json.Indent(&buf, body, "", "  ") == nil {
		fmt.Println(buf.String())
	} else {
		fmt.Println(string(body))
	}

	// Parse and return result
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("warning: error unmarshaling response: %v", err)
	}
	return result
}

func doGetRaw(client *http.Client, base, token, path string, params url.Values) []byte {
	u := base + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	req, _ := http.NewRequest("GET", u, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := client.Do(req)
	if err != nil {
		fatal("Request failed: %v", err)
		return nil
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		fmt.Fprintln(os.Stderr, styleStderr(cliErrorStyle, fmt.Sprintf("Error %d: %s", resp.StatusCode, string(body))))
		os.Exit(1)
	}
	return body
}

func doPost(client *http.Client, base, token, path string, body map[string]any) map[string]any {
	data, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", base+path, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := client.Do(req)
	if err != nil {
		fatal("Request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		fmt.Fprintln(os.Stderr, styleStderr(cliErrorStyle, fmt.Sprintf("Error %d: %s", resp.StatusCode, string(respBody))))
		os.Exit(1)
	}

	var buf bytes.Buffer
	if json.Indent(&buf, respBody, "", "  ") == nil {
		fmt.Println(buf.String())
	} else {
		fmt.Println(string(respBody))
	}

	// Parse and return result for suggestions
	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		log.Printf("warning: error unmarshaling response: %v", err)
	}
	return result
}

// checkServerAndGuide checks if pinchtab server is running and provides guidance
func checkServerAndGuide(client *http.Client, base, token string) bool {
	req, _ := http.NewRequest("GET", base+"/health", nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "dial tcp") {
			fmt.Fprintln(os.Stderr, styleStderr(cliErrorStyle, fmt.Sprintf("Pinchtab server is not running on %s", base)))
			fmt.Fprintln(os.Stderr)
			fmt.Fprintln(os.Stderr, styleStderr(cliHeadingStyle, "To start the server"))
			fmt.Fprintf(os.Stderr, "  %s %s\n", styleStderr(cliCommandStyle, "pinchtab"), styleStderr(cliMutedStyle, "# Run in foreground (recommended for beginners)"))
			fmt.Fprintf(os.Stderr, "  %s %s\n", styleStderr(cliCommandStyle, "pinchtab &"), styleStderr(cliMutedStyle, "# Run in background"))
			fmt.Fprintf(os.Stderr, "  %s %s\n", styleStderr(cliCommandStyle, "PINCHTAB_PORT=9868 pinchtab"), styleStderr(cliMutedStyle, "# Use different port"))
			fmt.Fprintln(os.Stderr)
			fmt.Fprintln(os.Stderr, styleStderr(cliHeadingStyle, "Then try your command again"))
			fmt.Fprintf(os.Stderr, "  %s\n", styleStderr(cliCommandStyle, strings.Join(os.Args, " ")))
			fmt.Fprintln(os.Stderr)
			fmt.Fprintf(os.Stderr, "%s %s\n", styleStderr(cliMutedStyle, "Learn more:"), styleStderr(cliCommandStyle, "https://github.com/pinchtab/pinchtab#quick-start"))
			return false
		}
		// Other connection errors
		fmt.Fprintln(os.Stderr, styleStderr(cliErrorStyle, fmt.Sprintf("Cannot connect to Pinchtab server: %v", err)))
		return false
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == 401 {
		fmt.Fprintln(os.Stderr, styleStderr(cliErrorStyle, "Authentication required. Set PINCHTAB_TOKEN."))
		return false
	}

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Fprintln(os.Stderr, styleStderr(cliErrorStyle, fmt.Sprintf("Server error %d: %s", resp.StatusCode, string(body))))
		return false
	}

	return true
}

// suggestNextAction provides helpful suggestions based on the current command and state
func suggestNextAction(cmd string, result map[string]any) {
	switch cmd {
	case "nav", "navigate":
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, styleStderr(cliHeadingStyle, "Next steps"))
		fmt.Fprintf(os.Stderr, "  %s %s\n", styleStderr(cliCommandStyle, "pinchtab snap"), styleStderr(cliMutedStyle, "# See page structure"))
		fmt.Fprintf(os.Stderr, "  %s %s\n", styleStderr(cliCommandStyle, "pinchtab screenshot"), styleStderr(cliMutedStyle, "# Capture visual"))
		fmt.Fprintf(os.Stderr, "  %s %s\n", styleStderr(cliCommandStyle, "pinchtab click <ref>"), styleStderr(cliMutedStyle, "# Click an element"))
		fmt.Fprintf(os.Stderr, "  %s %s\n", styleStderr(cliCommandStyle, "pinchtab pdf --tab <id> -o output.pdf"), styleStderr(cliMutedStyle, "# Save tab as PDF"))

	case "snap", "snapshot":
		refs := extractRefs(result)
		if len(refs) > 0 {
			fmt.Fprintln(os.Stderr)
			fmt.Fprintln(os.Stderr, styleStderr(cliHeadingStyle, fmt.Sprintf("Found %d interactive elements", len(refs))))
			for i, ref := range refs[:min(3, len(refs))] {
				fmt.Fprintf(os.Stderr, "  %s %s\n", styleStderr(cliCommandStyle, fmt.Sprintf("pinchtab click %s", ref.id)), styleStderr(cliMutedStyle, "# "+ref.desc))
				if i >= 2 {
					break
				}
			}
			if len(refs) > 3 {
				fmt.Fprintf(os.Stderr, "  %s\n", styleStderr(cliMutedStyle, fmt.Sprintf("... and %d more", len(refs)-3)))
			}
		}

	case "click", "type", "fill":
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, styleStderr(cliHeadingStyle, "Action completed"))
		fmt.Fprintf(os.Stderr, "  %s %s\n", styleStderr(cliCommandStyle, "pinchtab snap"), styleStderr(cliMutedStyle, "# See updated page"))
		fmt.Fprintf(os.Stderr, "  %s %s\n", styleStderr(cliCommandStyle, "pinchtab screenshot"), styleStderr(cliMutedStyle, "# Visual confirmation"))
	}
}

type refInfo struct {
	id   string
	desc string
}

func extractRefs(data map[string]any) []refInfo {
	var refs []refInfo

	// Handle different snapshot formats
	if elements, ok := data["elements"].([]any); ok {
		for _, elem := range elements {
			if m, ok := elem.(map[string]any); ok {
				if ref, ok := m["ref"].(string); ok && ref != "" {
					desc := ""
					if role, ok := m["role"].(string); ok {
						desc = role
					}
					if name, ok := m["name"].(string); ok && name != "" {
						desc += ": " + name
					}
					// Only include interactive elements
					if role, ok := m["role"].(string); ok {
						if role == "button" || role == "link" || role == "textbox" ||
							role == "checkbox" || role == "radio" || role == "combobox" {
							refs = append(refs, refInfo{id: ref, desc: desc})
						}
					}
				}
			}
		}
	}

	return refs
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// resolveInstanceBase fetches the named instance from the orchestrator and returns
// a base URL pointing directly at that instance's API port.
func resolveInstanceBase(orchBase, token, instanceID, bind string) string {
	c := &http.Client{Timeout: 10 * time.Second}
	body := doGetRaw(c, orchBase, token, fmt.Sprintf("/instances/%s", instanceID), nil)

	var inst struct {
		Port string `json:"port"`
	}
	if err := json.Unmarshal(body, &inst); err != nil {
		fatal("failed to parse instance %q: %v", instanceID, err)
	}
	if inst.Port == "" {
		fatal("instance %q has no port assigned (is it still starting?)", instanceID)
	}
	return fmt.Sprintf("http://%s:%s", bind, inst.Port)
}

func fatal(format string, args ...any) {
	fmt.Fprintln(os.Stderr, styleStderr(cliErrorStyle, fmt.Sprintf(format, args...)))
	os.Exit(1)
}
