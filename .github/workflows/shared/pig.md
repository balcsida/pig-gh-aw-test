---
engine:
  id: pig
  detection-engine: copilot
  version: "0.7.0"
  display-name: PiG
  description: PiG as an engine of its own, not behind the Pi engine's contract
  experimental: true
  provider:
    name: github
  behaviors:
    secret-strategy: universal-llm-consumer
    network:
      defaults:
        - host.docker.internal
        - github.com
        - raw.githubusercontent.com
        - api.github.com
        - objects.githubusercontent.com
      provider-domains:
        copilot: api.githubcopilot.com
    execution:
      command-name: pig
      step-name: Execute PiG
      model-env-var: GH_AW_PIG_MODEL
      mcp-config-env-var: GH_AW_MCP_CONFIG
      provider-env-mode: universal-llm-consumer
      env:
        XDG_RUNTIME_DIR: /tmp/gh-aw/pig-runtime  # the runner's /run/user/<uid> is not writable inside the AWF chroot
    mcp:
      # gh-aw's Pi converter writes the gateway servers in the mcp.json format that PiG reads.
      config-adapter: |
        process.env.GH_AW_PI_NATIVE_MCP = "1"; // keep the servers that are also mounted as CLIs
        require("./convert_gateway_config_pi.cjs").main();
    harness-script: |
      const { spawnSync } = require("child_process");
      const { mkdirSync, readFileSync, writeFileSync } = require("fs");
      const { join } = require("path");
      const { fetchAWFReflect, resolveOpenAICompatibleEndpointFromReflect } = require("./awf_reflect.cjs");

      const [command, ...commandArgs] = process.argv.slice(2);
      const log = message => process.stderr.write(`[pig-harness] ${message}\n`);
      // shortcut: these are the limits of gpt-4.1, add a table per model when a workflow uses another one
      const modelMetadata = { reasoning: false, input: ["text", "image"], contextWindow: 128000, maxTokens: 16384 };

      const main = async () => {
        const model = (process.env.GH_AW_PIG_MODEL || "").replace(/^[^/]+\//, "");
        if (!model) throw new Error("engine.model is required");

        // AWF holds the credential. PiG gets the URL of its proxy and a placeholder key.
        const reflect = await fetchAWFReflect({ logger: log });
        if (!reflect.ok || !reflect.reflectData) throw new Error(`no /reflect data: ${reflect.reason || "empty response"}`);
        const endpoint = resolveOpenAICompatibleEndpointFromReflect({ provider: process.env.GH_AW_LLM_PROVIDER, reflectData: reflect.reflectData, logger: log });
        if (!endpoint) throw new Error(`no /reflect endpoint for provider ${process.env.GH_AW_LLM_PROVIDER || "(missing)"}`);
        const baseUrl = `${endpoint.host}/${endpoint.basePath}`.replace(/\/chat\/completions$/, "");

        const agentDir = "/tmp/gh-aw/pig-agent-dir";
        mkdirSync(agentDir, { recursive: true, mode: 0o700 });
        const provider = { baseUrl, api: "openai-completions", apiKey: "awf-proxy", models: [{ id: model, ...modelMetadata }] };
        writeFileSync(join(agentDir, "models.json"), JSON.stringify({ providers: { "aw-gateway": provider } }), { mode: 0o600 });

        if (process.env.GH_AW_MCP_CONFIG) {
          const mcp = JSON.parse(readFileSync(process.env.GH_AW_MCP_CONFIG, "utf8"));
          // "direct" declares the tools to the model, so it needs no tool search to find safe outputs.
          for (const server of Object.values(mcp.mcpServers || {})) server.exposure = "direct";
          writeFileSync(join(agentDir, "mcp.json"), JSON.stringify(mcp), { mode: 0o600 });
        }

        const args = [...commandArgs, "--print", "--mode", "json", "--no-session", "--no-approve", "--model", `aw-gateway/${model}`];
        log(`running ${command} ${args.join(" ")}`);
        const result = spawnSync(command, args, {
          input: readFileSync(process.env.GH_AW_PROMPT),
          stdio: ["pipe", "inherit", "inherit"],
          env: { ...process.env, PIG_CODING_AGENT_DIR: agentDir },
        });
        if (result.error) throw result.error;
        process.exitCode = result.status ?? 1;
      };

      main().catch(error => {
        log(error instanceof Error ? error.message : String(error));
        process.exitCode = 1;
      });
    # PiG's --mode json stream is Pi's, so gh-aw's Pi parser reads it.
    log-parser: |
      function parseLog(logContent) {
        return require("./parse_pi_log.cjs").parsePiLog(logContent);
      }
steps:
  - name: Install PiG
    run: |
      set -euo pipefail
      cd "$RUNNER_TEMP"
      curl -fsSL -o pig.tar.gz https://github.com/MichaelKinsy/PiG/releases/download/v0.7.0/pig-0.7.0-linux-amd64.tar.gz
      echo "0b5ab7cb5f680e7d3353bdaa3a1018863c3c3552f88bec730370668e5819995a  pig.tar.gz" | sha256sum -c -
      tar xzf pig.tar.gz
      sudo install -m 0755 pig-0.7.0-linux-amd64/pig /usr/local/bin/pig
      pig --version
---

<!--
# PiG

A gh-aw engine definition for [PiG](https://github.com/MichaelKinsy/PiG). Import this file and set `engine: id: pig`.

It is an experiment and not an official definition of either project.
-->
