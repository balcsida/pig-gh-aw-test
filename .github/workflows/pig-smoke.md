---
on:
  workflow_dispatch:
permissions:
  contents: read
  copilot-requests: write
engine:
  id: pi
  command: pig
  model: copilot/claude-haiku-4.5
  env:
    PIG_VERSION_FORMAT: pi   # gh-aw requires `--version` >= 1.0.0; PiG prints 0.7.0+1.5.0 by default
    PIG_USE_PI_DIRS: "1"      # make PiG read PI_CODING_AGENT_DIR, where gh-aw writes models.json and settings.json
    XDG_RUNTIME_DIR: /tmp/gh-aw/pig-runtime  # the runner's /run/user/<uid> is not writable inside the AWF chroot
  # A custom executable has no npm SDK catalog, so gh-aw needs the model metadata here.
  config: |
    {"model": {"reasoning": true, "input": ["text", "image"], "contextWindow": 200000, "maxTokens": 64000}}
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
tools:
  bash: true
safe-outputs:
  create-issue:
    max: 1
---

# PiG smoke test

Read `README.md` in this repository. Run `uname -a` with the bash tool.

Then create one issue titled "PiG smoke test" whose body contains a one-sentence summary of the README and the `uname -a` output.
