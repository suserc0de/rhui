# RHUI VS Code Extension

Official Visual Studio Code syntax highlighting and language configuration for **RHUI** (`.rhui`) template files.

## Features
- **Semantic Tag Highlighting:** Distinct coloring for `<ui:Component>`, `<component>`, `<props>`, and `<slot:name>`
- **Directives & Controls:** Highlighted `<if>`, `<else>`, `<each>`
- **Dynamic Binding:** Highlighting for `:prop="expression"` and `:attrs="extra"`
- **Interpolation:** Bracket-pair matching and highlighting for `{{ expression | filter }}`
- **HTML Autocompletion:** Native auto-closing tags, quotes, and HTML block comments (`<!-- ... -->`)

## Installation

### Option 1: Symlink / Copy to Extensions Folder (Recommended for local use)
You can directly link or copy this folder to your VS Code extensions directory:

**Linux / macOS:**
```bash
ln -s "$(pwd)/tooling/vscode" ~/.vscode/extensions/rhui-vscode
```

**Windows (PowerShell):**
```powershell
New-Item -ItemType SymbolicLink -Path "$HOME\.vscode\extensions\rhui-vscode" -Target "$PWD\tooling\vscode"
```

### Option 2: Build into `.vsix` package
If you have `npm` installed:
```bash
cd tooling/vscode
npx @vscode/vsce package
```
Then in VS Code:
1. Open Extensions (`Ctrl+Shift+X` or `Cmd+Shift+X`).
2. Click the `...` (Views and More Actions) menu in the top right.
3. Select **Install from VSIX...** and choose the generated `.vsix` file.
