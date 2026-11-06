# Command Metadata

Read this before adding or changing commands in `bin/`.

Commands in `bin/` can declare CLI metadata in comments near the top of the
file. `bin/lunaos` scans the first 80 lines, and tests expect command metadata
to remain valid.

Supported metadata keys:

- `# lunaos:group=...` - override the command group inferred from the filename
- `# lunaos:name=...` - override the command name inferred from the filename
- `# lunaos:summary=...` - short help text
- `# lunaos:args=...` - usage arguments
- `# lunaos:examples=...` - examples separated with ` | `
- `# lunaos:alias=...` / `# lunaos:aliases=...` - alternate routes
- `# lunaos:hidden=true` - hide from default command listings
- `# lunaos:requires-sudo=true` - mark commands that require sudo

Only use `lunaos:examples` where there are args that need explaining.

Prefer explicit metadata for user-facing commands. Keep routes consistent with
the filename unless there is a deliberate alias or compatibility route.

Example:

```bash
# lunaos:summary=Take a screenshot
# lunaos:args=[smart|region|windows|fullscreen] [slurp|copy]
# lunaos:examples=lunaos screenshot | lunaos capture screenshot region
```
