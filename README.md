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
- Much of it overlaps with what gh-aw already has per workflow: tool restrictions, bare mode and `engine.extensions`.
- Two traps, from PiG's docs and not tested here. A Piglet built with `build.closed` refuses `-e`, which gh-aw uses to load its own extensions. A Piglet with an `extensions:` section and no `pig-mcp-adapter` entry removes all MCP tools, and that would break safe outputs.

### When I would use it

With a Piglet, when one reviewed and signed agent definition should be shared across many workflows, or when the same Piglet already runs locally. Otherwise stock Pi.

This rests on four runs with one simple prompt. It is not a benchmark.

## Not tested

Sub-agents, `engine.extensions`, bare mode, persistent sessions, a restricted bash tool policy, and the SDK and RPC drivers (`engine.driver`). The SDK driver imports the Pi npm package, so I don't expect that one to work with a Go binary.

The Piglet Binary is built on every run here, which takes about two minutes. A real setup would build it once and download it from a release.

## Notes

- `copilot/claude-haiku-4.5` and the default `copilot/auto` were rejected with `model_not_supported` for stock Pi as well, so that is a limit of this account and not of PiG.
- Issues #2 to #9, #14 and #15 were opened by gh-aw for the failed attempts on the way here.
