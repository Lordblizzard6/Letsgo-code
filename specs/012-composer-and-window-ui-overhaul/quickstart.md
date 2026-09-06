# Quickstart Validation Guide: Composer Layout and Window Sizing

## Verification Scenarios

### Scenario 1: Initial Window Size
1. Run `letsgo gui` or `go run . gui`.
2. Verify window opens with 1280x850 resolution.
3. Verify window can be resized down to 960x600 but not smaller.

### Scenario 2: Sidebar "Nueva conversación" Button
1. Look at the lateral sidebar.
2. Verify a prominent `[ + Nueva conversación ]` button is visible at the top of the conversations section.
3. Click the button: verify a new conversation tab is created and focused immediately.

### Scenario 3: Work Mode Cycling with Tab
1. Focus the composer textarea.
2. Notice the top corner shows `[ Plan ] [ Código ] [ Arch ] [ Research ]` with `Tab ⇥`.
3. Press `Tab`: mode advances from Código to Arch.
4. Press `Tab` again: mode advances to Plan.
5. Press `Shift+Tab`: mode returns to Arch.
6. Verify focus remains uninterrupted in the textarea.

### Scenario 4: Autocomplete Tab Override
1. Type `/` in the textarea.
2. Verify the command viewer popup appears with command icons and descriptions.
3. Press `Tab`: the highlighted command is autocompleted; mode DOES NOT switch.

### Scenario 5: Bottom Toolbar Disposition
1. Verify the bottom-left of the composer container has `[📎]` (Attach) and `[🤖 Model Selector ▾]`.
2. Verify the bottom-right has `[↑ Enviar]`.
3. Click the model selector dropdown: verify model list opens and selecting a model updates the active model cleanly.
