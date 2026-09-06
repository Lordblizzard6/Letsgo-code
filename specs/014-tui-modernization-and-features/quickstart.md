# Quickstart: Validating TUI Modernization & Features

## Prerequisites
- Working Go environment (`go build ./...`).
- A local git repository or the current repository.

## Interactive Verification Scenarios

### Scenario 1: Header & Status Bar Verification
1. Run `./letsgo`.
2. Inspect the top header:
   - Should show repository name, active branch, and session ID.
3. Inspect bottom status bar:
   - Should show `[BUILD MODE]` badge, active model, token counter, and cost.
4. Press `Tab`:
   - Mode badge should toggle to `[PLAN MODE]`, then `[RESEARCH MODE]`, then `[BUILD MODE]`.

### Scenario 2: Dynamic Model Selector (`Ctrl+S` / `/model`)
1. In the TUI, press `Ctrl+S`.
2. Model picker overlay opens.
3. Use arrows and type query to filter models.
4. Press `Enter` to select a model.
5. Verify chat shows confirmation message and status bar reflects the chosen model.

### Scenario 3: Slash Command Autocomplete
1. Type `/` in the prompt textarea.
2. Verify interactive suggestions popup appears with descriptions.
3. Type `di` and see suggestions filter to `/diff`.
4. Press `Tab` or `Enter` to complete the command.

### Scenario 4: Git Review & Staging (`Ctrl+D` / `/diff`)
1. Make a small edit in any file.
2. In TUI, press `Ctrl+D` or run `/diff`.
3. Verify the file is listed under Unstaged with `+` additions and `-` deletions.
4. Press `+` or `s` to stage the file. File moves to Staged.
5. Press `Enter` to preview the colored diff.
6. Press `Esc` to return to chat.

### Scenario 5: Activity Inspector (`Ctrl+B`)
1. Send a prompt that triggers a tool call (e.g. `list files`).
2. Press `Ctrl+B`.
3. Inspect executed tools with duration, parameters, and results.
4. Press `Esc` to return to chat.
