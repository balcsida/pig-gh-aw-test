---
on:
  workflow_dispatch:
permissions:
  contents: read
  copilot-requests: write
engine:
  id: pi
  model: copilot/gpt-4.1
tools:
  bash: true
safe-outputs:
  create-issue:
    max: 1
---

# Pi control

Read `README.md` in this repository. Run `uname -a` with the bash tool.

Then create one issue titled "Pi control" whose body contains a one-sentence summary of the README and the `uname -a` output.
