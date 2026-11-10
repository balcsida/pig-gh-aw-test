---
on:
  workflow_dispatch:
permissions:
  contents: read
  copilot-requests: write
engine:
  id: pi
tools:
  bash: true
safe-outputs:
  create-issue:
    max: 1
---

# Pi control (auto)

Read `README.md` in this repository. Run `uname -a` with the bash tool.

Then create one issue titled "Pi control (auto)" whose body contains a one-sentence summary of the README and the `uname -a` output.
