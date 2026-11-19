---
on:
  workflow_dispatch:
permissions:
  contents: read
  copilot-requests: write
imports:
  - shared/pig.md
engine:
  id: pig
  command: pig-gh-aw-closed
  model: copilot/gpt-4.1
steps:
  - name: Build the Piglet Binary
    run: |
      set -euo pipefail
      pig piglet build ./piglet/gh-aw-closed.yaml --format binary --out "$RUNNER_TEMP/pig-gh-aw-closed" --no-input --verbose
      sudo install -m 0755 "$RUNNER_TEMP/pig-gh-aw-closed" /usr/local/bin/pig-gh-aw-closed
      pig-gh-aw-closed --version
tools:
  bash: true
safe-outputs:
  create-issue:
    max: 1
  # With the PiG engine definition, detection runs Copilot with gh-aw's `detection` model alias, and this account can't use that model.
  threat-detection:
    engine:
      id: copilot
      model: gpt-4.1
---

# PiG engine closed test

Read `README.md` in this repository. Run `uname -a` with the bash tool.

Then create one issue titled "PiG engine closed test" whose body contains a one-sentence summary of the README and the `uname -a` output.

In the issue body, also name the exact tool or command you used to create the issue.
