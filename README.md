# pig-gh-aw-test

An experiment: running [PiG](https://github.com/MichaelKinsy/PiG) as the `pi` engine of [GitHub Agentic Workflows](https://github.github.com/gh-aw/) (gh-aw).

PiG is a Go implementation of [Pi](https://github.com/earendil-works/pi). gh-aw has a built-in Pi engine and lets you replace its executable with `engine.command`, so the question was whether a PiG binary works there without changes to gh-aw. This is not an official integration of either project.

## Result

It works. Tested with gh-aw v1.0.1 and PiG 0.7.0 (which ports Pi 1.5.0), on `ubuntu-latest`, inside the gh-aw firewall sandbox.

| Workflow | Engine | Model | Result |
|---|---|---|---|
| [`pig-smoke`](.github/workflows/pig-smoke.md) | `pig` | `copilot/gpt-4.1` | [#11](https://github.com/balcsida/pig-gh-aw-test/issues/11) |
| [`pig-anthropic`](.github/workflows/pig-anthropic.md) | `pig` | `anthropic/claude-haiku-5-5` | [#16](https://github.com/balcsida/pig-gh-aw-test/issues/16) |
| [`pig-piglet`](.github/workflows/pig-piglet.md) | a Piglet Binary built in the workflow | `copilot/gpt-4.1` | [#13](https://github.com/balcsida/pig-gh-aw-test/issues/13) |
| [`pi-control`](.github/workflows/pi-control.md) | stock Pi from npm | `copilot/gpt-4.1` | [#10](https://github.com/balcsida/pig-gh-aw-test/issues/10) |

Each workflow asks the agent to read this README, run `uname -a` and open an issue through gh-aw's safe outputs. PiG and stock Pi made the same three tool calls: `read`, `bash` and `mcp__safeoutputs__create_issue`.

The Piglet in [`piglet/`](piglet/) keeps only the `read` and `bash` tools and has its own system prompt. In its run the agent reported that it had no write tool, and the issue ends with the marker line from that prompt.

## What PiG needs

```yaml
engine:
  id: pi
  command: pig
  model: anthropic/claude-haiku-5-5
  env:
    PIG_VERSION_FORMAT: pi
    PIG_USE_PI_DIRS: "1"
    XDG_RUNTIME_DIR: /tmp/gh-aw/pig-runtime
  config: |
    {"model": {"api": "anthropic-messages", "compat": {"forceAdaptiveThinking": true}, "reasoning": true, "input": ["text", "image"], "contextWindow": 1000000, "maxTokens": 128000}}
steps:
  - name: Install PiG
    run: |
      # download the release, check its SHA-256, install it to /usr/local/bin
```

Why each setting is there:

- `PIG_VERSION_FORMAT: pi`. gh-aw runs `<command> --version` and requires 1.0.0 or newer. PiG prints `0.7.0+1.5.0` by default. With this it prints the Pi version it ports.
- `PIG_USE_PI_DIRS: "1"`. gh-aw writes `models.json` and `settings.json` to `PI_CODING_AGENT_DIR`. PiG reads `PIG_CODING_AGENT_DIR` unless this is set, and then fails with `Model "aw-gateway/..." not found`.
- `XDG_RUNTIME_DIR`. PiG runs extensions as subprocesses and puts their sockets under `$XDG_RUNTIME_DIR`. The runner's `/run/user/<uid>` is not writable inside the sandbox, so every extension fails to load with `mkdir /run/user: permission denied`.
- `engine.config.model`. With stock Pi, gh-aw reads model metadata from the npm package. A custom executable has no such catalog, so the API, the thinking mode and the limits have to be written out per model. Without `forceAdaptiveThinking`, Anthropic rejects the request with `"thinking.type.enabled" is not supported for this model`.

## Is it worth it?

For most gh-aw users, no. Stock Pi is the supported path and needs no extra configuration. PiG gives the same agent behaviour with more setup. It pays off in two cases, and both are about control over what runs, not about speed.

### Where PiG is better

- One pinned binary. The workflow downloads one file and checks its SHA-256, where stock Pi is an npm install with a dependency tree. PiG's releases also ship Sigstore bundles and SBOMs.
- Install time: 1s against 15s. On a job of about two minutes that is a small gain. The agent step was not faster (57s against 53s, one run each).

### Where PiG is worse

- gh-aw does not support it. Its docs say a custom `engine.command` has to satisfy the engine's contract itself, and the compiler's version pinning does not cover it. A gh-aw update can break it. I tested v1.0.1 only.
- Model metadata is written by hand for every model (see `engine.config.model` above). A mistake there fails at the provider.
- Three environment variables to make PiG look like Pi. One of them works around a PiG problem in the sandbox.
- The SDK driver can't work with a Go binary. From reading gh-aw's scripts, sub-agents need the Pi npm package in some configurations too.
- Node is still required, because gh-aw's own extensions are JavaScript.

### What a Piglet adds

A Piglet Binary fixes the agent's definition into one file: its tools, system prompt, extensions and skills.

- The tool scope removes tools instead of denying calls. In the test the agent had no write tool, and `tool_search` did not find one. The scope goes with the binary into every workflow that uses it.
- The agent is the same on a laptop and in CI, if you already run that Piglet locally.
- A Piglet Binary can be signed, so a workflow can check who built it before it runs. I did not test signing.

The costs:

- It needs a build and release pipeline. Building in the workflow took about two minutes per run, and the binary is 69 MB.
- Much of it overlaps with what gh-aw already has per workflow: tool restrictions, bare mode and `engine.extensions`. See [Without a Piglet](#without-a-piglet).
- Two traps, see [Piglet traps](#piglet-traps). A Piglet built with `build.closed` does not run under gh-aw at all. A Piglet with an `extensions:` section and no `pig-mcp-adapter` entry loses all MCP tools. Safe outputs then work only through gh-aw's CLI, and the agent step took 11 times as long.

### When I would use it

With a Piglet, when one reviewed and signed agent definition should be shared across many workflows, or when the same Piglet already runs locally. Otherwise stock Pi.

This rests on ten workflows with one simple prompt. It is not a benchmark.

## Piglet traps

PiG's docs describe two ways a Piglet can go wrong under gh-aw. I tested both, each with the Piglet of `pig-piglet` and one change to it, and then the fix for the second one.

| Workflow | Change to the Piglet | Result |
|---|---|---|
| [`pig-piglet-closed`](.github/workflows/pig-piglet-closed.md) | [`build.closed: true`](piglet/gh-aw-closed.yaml) | the engine exits, [#17](https://github.com/balcsida/pig-gh-aw-test/issues/17) |
| [`pig-piglet-ext`](.github/workflows/pig-piglet-ext.md) | an [`extensions:` section](piglet/gh-aw-ext.yaml) with no `pig-mcp-adapter` entry | no MCP tools, but [#18](https://github.com/balcsida/pig-gh-aw-test/issues/18) was created |
| [`pig-piglet-mcp`](.github/workflows/pig-piglet-mcp.md) | the same section [with the entry](piglet/gh-aw-mcp.yaml) | MCP tools are back, [#20](https://github.com/balcsida/pig-gh-aw-test/issues/20) |

### `build.closed`

A closed Binary does not work as the engine. gh-aw passes seven `--extension` options, four for its own extensions and three for built-in ones, and the Binary refuses them. It exits with code 1 before the agent starts:

```text
pig: this Piglet Binary (gh-aw-closed) was built with build.closed: true, so it runs only what its signed closure lists and does not accept -e/--extension; rebuild its Piglet without build.closed to accept it, or use stock pig
```

gh-aw reports this as "engine driver crashed". Run locally, the Binary refuses `--extension builtin:mcp` in the same way as a file path.

A Piglet for gh-aw therefore has to be an open Binary. PiG's security docs give their guarantees for closed Binaries only.

### `extensions:` without `pig-mcp-adapter`

The Piglet lists one extension of its own, [`marker`](piglet/extensions/marker/extension.go). Its `marker_ping` tool worked, and all MCP tools were gone, `mcp__github__*` as well as `mcp__safeoutputs__*`. `pig piglet show` gives the reason:

```text
MCP tools: hidden, because the Piglet lists `extensions:` but not `pig-mcp-adapter`
```

Safe outputs still worked. gh-aw also mounts its MCP servers as CLIs on `PATH`, and its prompt says so. The agent made 162 `tool_search` calls that found nothing, then ran `safeoutputs create_issue` with the bash tool. The run is green.

The cost shows in the numbers. The agent step took 10m41s against 56s for `pig-piglet`, and used 69 AI credits against 4. Nothing in the run's log or in the agent artifact says that the Piglet hid the tools.

This is one run with `gpt-4.1`. A Piglet without the bash tool has no way to the CLI, so I'd expect that one to fail.

A `pig-mcp-adapter` entry in `extensions:` fixes it. With the entry, `pig piglet show` no longer prints the line above, and in `pig-piglet-mcp` the agent called `mcp__safeoutputs__create_issue` directly. The agent step took 56s and 4 AI credits again.

## Without a Piglet

Stock `pig` can do what the Piglet above does, with gh-aw's own settings. [`pig-ext`](.github/workflows/pig-ext.md) has no Piglet:

| The Piglet has | `pig-ext` has |
|---|---|
| `tools: [read, bash]` | `tools.edit: false`. gh-aw then passes `--exclude-tools edit,write`. |
| `systemPrompt` | the workflow body |
| `extensions:` | `engine.extensions`. gh-aw runs `pig install` for each entry. |

MCP needs nothing. gh-aw writes `mcp.json` into the agent dir and loads PiG's built-in MCP support with `--extension builtin:mcp`, as in every other run here.

In the run the agent called `marker_ping`, found no write tool and created [#19](https://github.com/balcsida/pig-gh-aw-test/issues/19) with `mcp__safeoutputs__create_issue`. The agent step took 77s. It still wrote `scratch.txt` with bash, and nothing stops that in the Piglet either.

It needs one more setting. gh-aw's `pig install` step gets `PI_CODING_AGENT_DIR` but not `engine.env`, so `PIG_USE_PI_DIRS` has to be in the workflow's `env:` as well. Without it PiG ignores `PI_CODING_AGENT_DIR`, which I checked locally only, and installs the extension where the agent does not look.

## PiG as an engine of its own

gh-aw can also take an engine from a definition file. [`shared/pig.md`](.github/workflows/shared/pig.md) defines a `pig` engine with `engine.behaviors`, and [`pig-engine`](.github/workflows/pig-engine.md) imports it:

```yaml
imports:
  - shared/pig.md
engine:
  id: pig
  model: copilot/gpt-4.1
```

It works. The agent made the usual three tool calls and created [#23](https://github.com/balcsida/pig-gh-aw-test/issues/23) with `mcp__safeoutputs__create_issue`. The agent step took 44s.

The definition is about 100 lines. Its harness script asks AWF for the inference endpoint, writes `models.json` and `mcp.json` into an agent dir of its own and starts `pig --print --mode json`. For the MCP config and the log summary it calls gh-aw's Pi scripts, because PiG reads and writes Pi's formats.

What goes away, compared to `engine.command`:

- `PIG_VERSION_FORMAT` and `PIG_USE_PI_DIRS`. There is no Pi contract to imitate.
- The install step and the model metadata in every workflow. Both are in the definition.
- gh-aw's seven `--extension` options. The harness passes none, so a closed Binary has nothing to refuse. [`pig-engine-closed`](.github/workflows/pig-engine-closed.md) runs the closed Piglet from the traps above as the engine, and it created [#24](https://github.com/balcsida/pig-gh-aw-test/issues/24).

What it costs:

- gh-aw's Pi extensions go with those options: its tool policy, steering and sub-agents. The compiler warns that it ignores the workflow's `tools:` section for this engine, so `tools.edit: false` does nothing. The tool scope has to come from a Piglet.
- The definition is mine to maintain, and it leans on three of gh-aw's internal scripts. gh-aw calls even its own sample definitions unsupported.
- It has the limits of `gpt-4.1` only, and only the Copilot route.
- Threat detection no longer follows the workflow's model. It ran Copilot with gh-aw's `detection` alias, which is `claude-haiku-5.5`, and this account can't use that model. So the first run's issue, #22, starts with the detection warning. With `safe-outputs.threat-detection.engine.model` set in the workflow, detection ran on `gpt-4.1` again.
- A closed Binary still has no MCP tools, because it ignores the servers in `mcp.json` as well. In `pig-engine-closed` PiG never connected to the MCP gateway, and the agent was left with the `safeoutputs` CLI again. Seven calls failed, six of them on shell quoting, and the eighth went through after the agent cut the issue body down. #24 has no `uname -a` output. The agent step took 2m35s and 7 AI credits, against 44s and 4.

## Not tested

Sub-agents, `engine.extensions` with an npm or git source, bare mode, persistent sessions, a restricted bash tool policy, and the SDK and RPC drivers (`engine.driver`). The SDK driver imports the Pi npm package, so I don't expect that one to work with a Go binary.

The Piglet Binary is built on every run here, which takes about two minutes. A real setup would build it once and download it from a release.

## Notes

- `copilot/claude-haiku-4.5` and the default `copilot/auto` were rejected with `model_not_supported` for stock Pi as well, so that is a limit of this account and not of PiG.
- Issues #2 to #9, #14 and #15 were opened by gh-aw for the failed attempts on the way here.
- #21 was opened by gh-aw when `pig-engine` ran into Copilot's rate limit (`429`) in three runs.
- #19 and #20 start with a "threat detection engine error" warning. gh-aw's detection step runs the Copilot CLI, and that got `429 Too Many Requests` in both runs.
