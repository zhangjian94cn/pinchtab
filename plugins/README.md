# PinchTab SMCP Plugins

This directory contains MCP (SMCP) plugins for use with [sanctumos/smcp](https://github.com/sanctumos/smcp) or compatible MCP servers.

## What is SMCP?

**SMCP** is an MCP (Model Context Protocol) server used in the Animus/Letta/Sanctum ecosystem. It discovers plugins in a directory (e.g. `plugins/`), runs `python cli.py --describe` to get tool schemas, and exposes each command as an MCP tool (e.g. `pinchtab__navigate`). AI agents can then call PinchTab through the MCP layer without using HTTP directly.

## SMCP setup (env and paths)

- **Plugin directory:** Point your MCP server at this repo’s `plugins/` directory.
  - **SMCP:** Set `MCP_PLUGINS_DIR` to the absolute path of `plugins/` (e.g. `/path/to/pinchtab/plugins`), or copy the `pinchtab` folder into your existing SMCP `plugins/` directory.
- **PinchTab URL:** The plugin defaults to `http://localhost:9867` (orchestrator). Agents pass `base_url` (and optionally `token`, `instance_id`) per tool call; no extra env vars are required for the plugin itself.
- **Optional:** If PinchTab is protected with `BRIDGE_TOKEN`, agents must pass `token` in the tool args (or you can set it in your MCP server config if it supports per-plugin env).

**Quick check:** From the repo root:
```bash
cd plugins/pinchtab && python3 cli.py --describe | head -20
```

## pinchtab

Full SMCP plugin for the PinchTab HTTP API: navigate, snapshot, action, text, screenshot, PDF, instances, tabs, and more.

- **Location:** `pinchtab/`
- **Discovery:** `python cli.py --describe` (JSON with plugin + commands)
- **Tests:** `cd pinchtab && python -m venv .venv && .venv/bin/pip install pytest && .venv/bin/pytest tests/ -v`

See [pinchtab/README.md](pinchtab/README.md) for usage and SMCP compatibility notes.
