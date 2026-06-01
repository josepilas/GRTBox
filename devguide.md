# GRTBox Developer Guide

This guide explains how to build GRTBox tool packages, how to write `.tc`
modules, how to prepare `manifest.json`, how `icon.png` should be created, and
how to package everything into a portable `.tl` file.

GRTBox is a Windows-first desktop toolbox launcher. It is not a web app. The
host application is built with Go, Wails, and Svelte, and each tool runs inside
the GRTBox desktop runtime through a package format named `.tl`.

The current goal of the platform is practical:

- tools should be easy to install
- tools should be portable across Windows machines running GRTBox
- tools should not depend on hardcoded local paths
- tool UI should be rendered by TC code
- privileged Windows operations should run through the hidden desktop runtime,
  not through visible PowerShell windows
- destructive actions must be explicit, confirmed, logged, and reversible when
  possible

GRTBox now requires administrator privileges at startup on Windows. If the
program is started without administrator access, it shows an administrator
requirement message and exits before loading the desktop UI. This is intentional
because even basic tools such as Internet Bridge, firmware flashing, adapter
inspection, and device repair workflows rely on elevated Windows APIs.

## Core Concepts

### What Is A `.tl` File?

A `.tl` file is a ZIP archive renamed to use the `.tl` extension.

It is not an executable.
It is not a web page.
It is not a JSON-only UI format.

It is a package containing:

```text
manifest.json
main.tc
```

It may also contain:

```text
icon.png
src/*.tc
assets/*
data/*
original/*
```

GRTBox copies installed packages to the user profile, validates them, extracts
them when opened, and loads their TC modules from the extracted copy.

### What Is TC?

TC is the tool scripting layer used by GRTBox packages. In the current runtime,
TC uses JavaScript module syntax:

```js
import { renderApp } from "./src/ui.tc";

export default async function main(runtime) {
  await renderApp(runtime);
}
```

Every package entry file must export a default async function. GRTBox calls that
function and passes a `runtime` object. The runtime provides generic desktop
primitives such as UI rendering, file dialogs, filesystem access, PowerShell,
process execution, logging, timers, environment variables, and path helpers.

TC tools should build their own domain logic from those generic primitives. Do
not expect feature-specific APIs such as `runtime.internetBridge`,
`runtime.dtb`, or `runtime.firmware`.

### Why Portability Matters

The same `.tl` should run on every GRTBox installation. That means a tool must
not depend on absolute paths from the developer machine, such as:

```text
C:\Users\Someone\Downloads\tool\src\data.json
```

Use package-relative imports and runtime dialogs instead. If the tool needs
bundled data, include it inside the package under `data/`, `assets/`, or
`original/`. If the tool needs user files, ask the user with a file dialog.

Installed tools live here:

```text
%APPDATA%\GRTBox\tools
```

Extracted tools live here:

```text
%APPDATA%\GRTBox\extracted_tools\<tool_id>\<version>
```

The loader resolves relative `.tc` imports from the extracted package, so this
works everywhere:

```js
import { identifyDtb } from "./src/identifier.tc";
```

This does not:

```js
import { identifyDtb } from "C:/Users/grand/Downloads/trerte/tools/identifier.tc";
```

## Recommended Package Layout

Use this layout for a normal tool:

```text
my_tool_src/
  manifest.json
  main.tc
  icon.png
  src/
    ui.tc
    actions.tc
    state.tc
    diagnostics.tc
  data/
    database.json
  assets/
    help.txt
```

Use this layout for a tool adapted from an original browser tool:

```text
dtb_general_src/
  manifest.json
  main.tc
  icon.png
  src/
    ui.tc
    identifier.tc
    extractor.tc
    hashes.tc
  data/
    known_dtbs.json
  original/
    DTB Identify.html
    DTB File Data Extractor.htm
    reference-files/
```

Keeping original files under `original/` is useful for attribution and future
maintenance. Do not load old browser code directly if it depends on DOM-only
global state. Port the logic into TC modules instead.

## Minimal Tool Example

`manifest.json`:

```json
{
  "id": "hello_tool",
  "name": "Hello Tool",
  "version": "1.0.0",
  "author": "Your Name",
  "description": "A minimal GRTBox TC tool.",
  "entry": "main.tc",
  "runtime": "tc",
  "target_platforms": ["windows"],
  "package_format_version": "1.0.0"
}
```

`main.tc`:

```js
export default async function main(runtime) {
  const h = runtime.ui.createElement;

  runtime.ui.render(
    h("section", { className: "tc-panel" }, [
      h("h1", {}, "Hello Tool"),
      h("p", {}, "This UI was rendered by TC."),
      h("button", {
        onClick: async () => {
          await runtime.logs.write("Button clicked.");
        }
      }, "Write Log")
    ])
  );
}
```

Package it:

```powershell
Push-Location .\hello_tool_src
tar --format zip -cf ..\hello_tool.tl manifest.json main.tc
Pop-Location
```

Install `hello_tool.tl` through the GRTBox UI.

## `manifest.json`

The manifest describes the package. It is read before any TC code runs.

### Required Fields

`id`

Stable package identifier. Use lowercase letters, digits, and underscores.
The ID should never contain spaces. Good IDs:

```text
internet_bridge
dtb_general
easy_firmware
```

`name`

Human-readable tool name shown in the launcher.

`version`

Tool version. Use semantic versioning when possible:

```text
1.0.0
1.1.0
2.0.0
```

When bundled files or behavior changes, bump the version. GRTBox extracts tools
under `<tool_id>/<version>`, so version bumps prevent stale extracted copies.

`author`

Tool author or maintainer.

`description`

Short launcher-card description. Keep it useful and direct.

`entry`

Entry TC module. Usually:

```json
"entry": "main.tc"
```

The path must be relative and must end in `.tc`.

`runtime`

Must currently be:

```json
"runtime": "tc"
```

### Recommended Optional Fields

`icon`

Relative path to a PNG icon:

```json
"icon": "icon.png"
```

If omitted or invalid, GRTBox uses the internal default tool icon.

`requires_admin`

Use this when the tool needs elevated privileges:

```json
"requires_admin": true
```

GRTBox itself already requires administrator privileges on Windows, but keeping
this field is still useful because the launcher can show the requirement and the
tool can document why it needs elevation.

`permissions`

List the runtime features the tool expects:

```json
"permissions": [
  "dialogs.openFile",
  "filesystem.readFile",
  "powershell.exec",
  "process.exec",
  "logs.write"
]
```

Current permissions are descriptive and validation-oriented. They make reviews
easier and help users understand what the tool does.

`target_platforms`

Use `windows` for Windows-only tools:

```json
"target_platforms": ["windows"]
```

Use multiple platforms only when the tool was actually tested across them:

```json
"target_platforms": ["windows", "linux", "macos"]
```

`min_grtbox_version`

Minimum GRTBox version needed:

```json
"min_grtbox_version": "0.1.0"
```

`package_format_version`

Current package format:

```json
"package_format_version": "1.0.0"
```

### Complete Manifest Example

```json
{
  "id": "internet_bridge",
  "name": "Internet Bridge",
  "version": "1.0.0",
  "author": "GRTBox",
  "description": "Share your PC internet connection with a handheld device over USB OTG.",
  "entry": "main.tc",
  "runtime": "tc",
  "icon": "icon.png",
  "requires_admin": true,
  "permissions": [
    "os.platform",
    "os.isAdmin",
    "powershell.exec",
    "process.exec",
    "logs.write"
  ],
  "target_platforms": ["windows"],
  "min_grtbox_version": "0.1.0",
  "package_format_version": "1.0.0"
}
```

## `icon.png`

Every production-quality package should include `icon.png`.

Rules:

- file name should be `icon.png`
- format must be PNG
- use a square canvas
- recommended size: 256x256 or 512x512
- transparent background is allowed
- keep the shape readable at 64x64
- avoid tiny text
- avoid using screenshots as icons
- avoid neon or overly busy artwork for utility tools

If `icon.png` is missing, invalid, or not referenced, GRTBox falls back to the
internal default icon. That is acceptable for prototypes, but real tools should
ship their own icon.

## Writing TC Modules

### Entry Function

Every entry module must export a default async function:

```js
export default async function main(runtime) {
  await runtime.logs.write("Tool started.");
}
```

If startup fails, throw an error with a useful message:

```js
export default async function main(runtime) {
  const platform = await runtime.os.platform();
  if (platform !== "windows") {
    throw new Error("This tool currently supports Windows only.");
  }
}
```

GRTBox shows tool startup errors in the runner with:

- `Tool crashed`
- the error message
- `Reload Tool`
- `Back to Tools`

The error is also logged.

### Imports

Use relative imports:

```js
import { renderApp } from "./src/ui.tc";
import { runDiagnostics } from "./src/diagnostics.tc";
```

Do not import from local absolute paths.

Do not import `.js` files. Use `.tc` modules for package code.

### Exports

Named exports are useful for organizing code:

```js
export function normalizeText(text) {
  return String(text || "").trim().replace(/\s+/g, " ");
}
```

Default exports should be reserved for the entry module.

### State

Use a plain object for state:

```js
const state = {
  busy: false,
  message: "",
  selectedFile: "",
  result: null
};
```

Then call `render()` after changes:

```js
state.message = "Done.";
render();
```

This simple model is predictable and works well for desktop utility tools.

### UI Rendering

The runtime provides:

```js
const h = runtime.ui.createElement;
```

Create elements like this:

```js
h("button", { onClick: runTask }, "Run")
```

Children can be strings, elements, or arrays:

```js
h("section", { className: "tc-panel" }, [
  h("h2", {}, "Diagnostics"),
  h("p", {}, state.message)
])
```

Render the root:

```js
runtime.ui.render(view());
```

Update the root after state changes:

```js
runtime.ui.update(view());
```

Clear the root:

```js
runtime.ui.clear();
```

### Common UI Elements

Button:

```js
h("button", { onClick: handleClick }, "Run")
```

Input:

```js
h("input", {
  value: state.query,
  placeholder: "Search...",
  onInput: (event) => {
    state.query = event.target.value;
    render();
  }
})
```

Textarea:

```js
h("textarea", {
  value: state.notes,
  onInput: (event) => {
    state.notes = event.target.value;
  }
})
```

Select:

```js
h("select", {
  value: state.mode,
  onChange: (event) => {
    state.mode = event.target.value;
    render();
  }
}, [
  h("option", { value: "safe" }, "Safe"),
  h("option", { value: "advanced" }, "Advanced")
])
```

Status row:

```js
function statusRow(label, value, className = "") {
  return h("div", { className: "tc-status-row " + className }, [
    h("span", {}, label),
    h("strong", {}, value)
  ]);
}
```

### Event Handlers

Event handlers may be async:

```js
h("button", {
  onClick: async () => {
    state.busy = true;
    render();
    try {
      await runWork();
      state.message = "Finished.";
    } catch (error) {
      state.message = error.message || String(error);
    } finally {
      state.busy = false;
      render();
    }
  }
}, state.busy ? "Running..." : "Run")
```

Always reset `busy` flags in `finally`.

### Styling

Tools inherit the GRTBox dark desktop environment. Keep tool UI practical:

- dark surfaces
- light text
- restrained borders
- clear spacing
- compact controls
- no neon effects
- no emojis
- no glassmorphism
- no excessive animation

Use simple class names:

```js
h("section", { className: "tc-panel" }, [...])
```

Avoid inline style except for small dynamic values.

## Runtime API

The runtime object contains these groups:

```text
ui
os
process
shell
powershell
filesystem
path
env
dialogs
logs
timers
http
json
events
```

### `runtime.ui`

```js
runtime.ui.createElement(type, props, children)
runtime.ui.render(node)
runtime.ui.update(node)
runtime.ui.clear()
```

Use this for all TC-rendered UI.

### `runtime.os`

```js
const platform = await runtime.os.platform();
const admin = await runtime.os.isAdmin();
```

`platform` returns values such as `windows`, `linux`, or `darwin` depending on
the host runtime. On GRTBox Windows builds, tools should normally expect
`windows`.

`isAdmin` checks whether the app is elevated. Since the Windows app now exits
without elevation, this should normally return true on Windows, but tools may
still check it for diagnostics.

### `runtime.process`

Run a specific executable with explicit arguments:

```js
const result = await runtime.process.exec("ipconfig.exe", ["/all"], {
  timeoutSeconds: 20
});
```

Prefer `process.exec` when you know the executable and arguments.

### `runtime.shell`

Run a command through the platform shell:

```js
const result = await runtime.shell.exec("where powershell", {
  timeoutSeconds: 10
});
```

Use this sparingly. Shell commands are easier to quote incorrectly.

### `runtime.powershell`

Run Windows PowerShell scripts:

```js
const result = await runtime.powershell.exec(
  "Get-NetAdapter | ConvertTo-Json -Depth 4",
  { timeoutSeconds: 20 }
);
```

GRTBox configures child PowerShell processes to run hidden on Windows. Tools
should not call `Start-Process powershell` just to run another script, because
that can create visible windows and detached processes. Run the script directly
through `runtime.powershell.exec`.

Use `ConvertTo-Json` when returning structured data:

```js
const script = `
$adapters = Get-NetAdapter | Select-Object Name, InterfaceDescription, Status
$adapters | ConvertTo-Json -Depth 4
`;
const result = await runtime.powershell.exec(script, { timeoutSeconds: 20 });
const adapters = JSON.parse(result.stdout || "[]");
```

### `runtime.filesystem`

```js
const text = await runtime.filesystem.readFile(path);
const base64 = await runtime.filesystem.readFileBase64(path);
await runtime.filesystem.writeFile(path, text);
const exists = await runtime.filesystem.exists(path);
```

Use `readFileBase64` for binary files when you need byte-accurate handling.

### `runtime.path`

```js
const full = runtime.path.join(base, "data", "firmware.json");
const name = runtime.path.basename(full);
const dir = runtime.path.dirname(full);
```

Use `runtime.path.join` instead of manual string concatenation.

### `runtime.env`

```js
const appData = await runtime.env.get("APPDATA");
```

Environment variables are useful for user-specific storage, but package code
should still avoid hardcoding developer paths.

### `runtime.dialogs`

Open file:

```js
const file = await runtime.dialogs.openFile({
  title: "Select DTB File",
  filters: [
    { displayName: "DTB Files (*.dtb)", pattern: "*.dtb" },
    { displayName: "All Files", pattern: "*.*" }
  ]
});
```

Save file:

```js
const path = await runtime.dialogs.saveFile({
  title: "Save Extracted Data",
  defaultFilename: "dtb-data.txt",
  filters: [
    { displayName: "Text File (*.txt)", pattern: "*.txt" }
  ]
});
```

Confirm:

```js
const ok = await runtime.dialogs.confirm("Continue?");
if (!ok) return;
```

Use confirmation before destructive or global actions, such as changing
Internet Connection Sharing, overwriting generated files, or flashing disks.

### `runtime.logs`

```js
await runtime.logs.write("Diagnostics started.");
```

Write clear logs for:

- startup
- selected files
- detected devices
- external commands
- warnings
- destructive actions
- failures

Do not log secrets, tokens, personal files, or full contents of private files.

### `runtime.timers`

```js
const id = runtime.timers.setInterval(refresh, 5000);
runtime.timers.clearInterval(id);
```

Use timers only when the UI truly needs polling. For tools like Internet Bridge,
avoid constant background polling that runs PowerShell forever while the user is
not doing anything.

### `runtime.http`

```js
const text = await runtime.http.get("https://example.com/status.json");
```

Use HTTP for small metadata only. Large firmware downloads should normally be
performed with PowerShell or a controlled process so the tool can write to a
file, verify hashes, and recover from failures.

### `runtime.json`

The runtime exposes JSON helpers through the standard JSON object:

```js
const data = runtime.json.parse(text);
const pretty = runtime.json.stringify(data, null, 2);
```

### `runtime.events`

Use events for internal communication between modules if needed:

```js
runtime.events.emit("diagnostics:done", result);
runtime.events.on("diagnostics:done", (event) => {
  console.log(event.detail);
});
```

For most small tools, direct function calls are simpler.

## PowerShell And Background Execution

Windows desktop tools often need PowerShell. The important rule is:

Run PowerShell through the GRTBox runtime, not by opening visible windows.

Good:

```js
await runtime.powershell.exec("Get-NetAdapter | ConvertTo-Json -Depth 4", {
  timeoutSeconds: 20
});
```

Avoid:

```js
await runtime.shell.exec("start powershell -NoExit -Command Get-NetAdapter");
```

GRTBox configures child commands with hidden windows on Windows. That prevents
the user from seeing random PowerShell tabs while a tool runs diagnostics or
network operations.

Use bounded timeouts:

```js
await runtime.powershell.exec(script, { timeoutSeconds: 30 });
```

Avoid infinite loops in PowerShell. If polling is required, use a TC timer and
run short commands at controlled intervals.

Return JSON when possible:

```powershell
Get-NetIPConfiguration | Select-Object InterfaceAlias, IPv4Address | ConvertTo-Json -Depth 5
```

Then parse in TC:

```js
const data = JSON.parse(result.stdout || "[]");
```

## Administrator Requirements

GRTBox checks administrator access before opening the UI on Windows. This keeps
tools from failing halfway through workflows that require elevation.

Tools should still do their own checks when admin access affects safety:

```js
const admin = await runtime.os.isAdmin();
if (!admin) {
  throw new Error("Administrator privileges are required for this tool.");
}
```

Use `requires_admin: true` in `manifest.json` for tools that:

- change network adapter settings
- enable or disable Internet Connection Sharing
- write raw disk images
- format partitions
- inspect protected devices
- modify system files
- run repair commands

## Files, Assets, And Data

### Bundled Text Data

Put JSON and text data under `data/`:

```text
data/firmware.json
data/known_dtbs.json
```

Read the data from a module using package-relative logic when the runtime
exposes the extracted path:

```js
const dataPath = runtime.path.join(runtime.tool.extractedPath, "data", "firmware.json");
const text = await runtime.filesystem.readFile(dataPath);
const database = JSON.parse(text);
```

### Bundled Original Files

For adapted tools, keep upstream originals under `original/`:

```text
original/DTB Identify.html
original/DTB File Data Extractor.htm
```

This makes attribution clear and allows future maintainers to compare the port
against the original behavior.

### Binary Files

Use `readFileBase64` when reading binary files:

```js
const encoded = await runtime.filesystem.readFileBase64(filePath);
```

For hashing, use PowerShell:

```js
const script = `Get-FileHash -Algorithm SHA256 -LiteralPath ${quotePowerShell(path)} | Select-Object -ExpandProperty Hash`;
const result = await runtime.powershell.exec(script, { timeoutSeconds: 60 });
const sha256 = result.stdout.trim().toLowerCase();
```

Quote PowerShell paths safely:

```js
export function quotePowerShell(value) {
  return "'" + String(value).replace(/'/g, "''") + "'";
}
```

## DTB General Guidance

DTB General combines the original DTB Identify and DTB File Data Extractor
browser tools into one GRTBox TC package.

Credits must remain visible:

```text
Original DTB tools by Aeolusux.
```

Recommended behavior:

- let the user select a `.dtb` file
- read it as binary when possible
- compute a stable hash
- compare against known DTB hashes
- show the detected panel/core result
- if hash matching fails, fall back to normalized text comparison only when
  that preserves compatibility with the original tool
- let the user export extracted text data
- never modify the user's original DTB file

Identifier logic should be deterministic. If two signatures match, show a
warning and list both candidates instead of pretending the result is certain.

Do not depend on browser-only APIs from the old HTML files. Port the logic into
TC modules and keep the old files under `original/` for attribution.

## Internet Bridge Guidance

Internet Bridge is Windows-first and depends on administrator privileges.

Recommended behavior:

- detect source adapters with `Get-NetAdapter` and `Get-NetIPConfiguration`
- detect the target USB/RNDIS adapter
- verify external internet reachability only when the user runs diagnostics
- enable or disable Internet Connection Sharing with `HNetCfg.HNetShare`
- log every system change
- avoid background PowerShell loops
- avoid visible PowerShell windows
- keep status polling short and user-triggered

The tool should not open PowerShell windows by itself. Use:

```js
await runtime.powershell.exec(script, { timeoutSeconds: 30 });
```

Do not use:

```js
Start-Process powershell
```

unless the user explicitly needs an interactive console, which normal tools
should not require.

## Easy Firmware Guidance

Easy Firmware uses a firmware database file:

```text
data/firmware.json
```

The database should store real source links, release pages, checksums when
available, and compatibility notes.

Recommended download fields:

```json
{
  "id": "rocknix_rk3326_20250517_a_r36s",
  "firmware_id": "rocknix",
  "device_id": "r36s",
  "version": "20250517-a",
  "url": "https://github.com/ROCKNIX/distribution/releases/download/20250517/ROCKNIX-RK3326.aarch64-20250517-a.img.gz",
  "download_method": "direct",
  "file_name": "ROCKNIX-RK3326.aarch64-20250517-a.img.gz",
  "archive_type": "gz",
  "image_format": "img",
  "sha256": "a299021442cdc6c1a134202c12c3bdd2742cacfb0d30e9a8be57a7045d0f7ebd",
  "release_page": "https://github.com/ROCKNIX/distribution/releases/tag/20250517"
}
```

Supported archive handling should be explicit:

- `zip` through built-in PowerShell extraction
- `gz` through a controlled decompression path
- `xz` through Windows `tar.exe` when available
- `7z` through external 7-Zip when available

Do not claim that a firmware image was verified unless a checksum was provided
and the computed hash matches it.

Firmware flashing is destructive. A flashing workflow must:

- block system disks
- block fixed internal disks
- require administrator privileges
- show the selected disk clearly
- require a typed confirmation such as `ERASE`
- show a final confirmation dialog
- log the selected disk and image path
- fail closed if disk safety cannot be determined

## Packaging `.tl` Files

The `.tl` archive must contain files at the package root. Do not compress the
folder itself.

Good archive contents:

```text
manifest.json
main.tc
icon.png
src/ui.tc
data/database.json
```

Bad archive contents:

```text
my_tool_src/manifest.json
my_tool_src/main.tc
```

PowerShell packaging with Windows `tar`:

```powershell
Push-Location .\examples\my_tool_src
tar --format zip -cf ..\my_tool.tl manifest.json main.tc icon.png src data
Pop-Location
```

If a package has no `icon.png`, remove it from the command.

Prefer `tar --format zip` because it writes portable ZIP entry paths with
forward slashes. Some PowerShell `Compress-Archive` flows can create backslash
paths inside the archive, and GRTBox rejects those packages.

If a package has an `original/` directory, include it:

```powershell
tar --format zip -cf ..\dtb_general.tl manifest.json main.tc icon.png src data original
```

## Validation Rules

GRTBox validation checks:

- extension must be `.tl`
- package must be a valid ZIP archive
- `manifest.json` must exist
- `manifest.json` must be valid JSON
- `main.tc` or the configured entry must exist
- required manifest fields must be present
- runtime must be `tc`
- entry must be a safe relative path
- entry must end in `.tc`
- ZIP paths must not escape the package root
- duplicate case-insensitive paths are rejected
- `icon.png`, when present, must be valid PNG
- tool IDs must be unique inside the installed tools directory

Run syntax checks on TC files before packaging:

```powershell
Get-ChildItem examples -Recurse -Filter *.tc | ForEach-Object {
  node --check $_.FullName
}
```

Validate JSON:

```powershell
Get-Content .\examples\easy_firmware_src\data\firmware.json -Raw | ConvertFrom-Json | Out-Null
```

Run backend tests:

```powershell
go test ./...
```

Run frontend checks:

```powershell
Push-Location frontend
npm install
npm run check
npm run build
Pop-Location
```

Build the desktop app:

```powershell
wails build -clean -platform windows/amd64
```

## Debugging

Use runtime logs first:

```js
await runtime.logs.write("Selected file: " + filePath);
```

Keep error messages specific:

```js
throw new Error("No removable disk was selected.");
```

When wrapping errors:

```js
try {
  await runTask();
} catch (error) {
  throw new Error("Firmware extraction failed: " + (error.message || String(error)));
}
```

For PowerShell diagnostics, include stderr in the message:

```js
if (result.exitCode !== 0) {
  throw new Error((result.stderr || result.stdout || "PowerShell command failed").trim());
}
```

## Security And Safety

Do:

- use file dialogs instead of hardcoded paths
- quote all paths passed to PowerShell
- set timeouts for every external command
- parse JSON instead of scraping formatted tables
- verify checksums when provided
- ask confirmation before destructive actions
- log high-level actions
- keep bundled upstream files for attribution
- keep tools offline-capable when the task can work offline

Do not:

- open random PowerShell windows
- run hidden infinite loops
- silently modify network sharing
- silently flash disks
- assume a device is safe because its name looks familiar
- download firmware from unverified reposts when an official release page exists
- store absolute developer paths in package code
- claim hash verification when no hash exists
- bundle private files or local caches into `.tl` packages

## Portability Checklist

Before shipping a `.tl`, confirm:

- all imports are relative
- no source file contains your local username or download path
- all bundled data is inside the package
- `manifest.json` has the correct version
- `icon.png` is valid or intentionally omitted
- the package root contains `manifest.json`
- the package root contains the configured entry file
- Windows-only tools declare `target_platforms: ["windows"]`
- administrator tools declare `requires_admin: true`
- PowerShell scripts run through `runtime.powershell.exec`
- long commands have timeouts
- destructive flows have confirmation
- source credits are preserved
- the tool was repackaged after source changes
- stale extracted copies were cleared or the package version was bumped

## Release Checklist

1. Update source files under `examples/<tool>_src`.
2. Bump the tool version in `manifest.json` when behavior or bundled data changes.
3. Validate JSON files.
4. Syntax-check all `.tc` modules.
5. Rebuild the `.tl` archive.
6. Copy the `.tl` to `%APPDATA%\GRTBox\tools`.
7. Remove stale extracted copies when testing locally.
8. Run `go test ./...`.
9. Run frontend checks and build.
10. Build the Windows executable with Wails.
11. Start the built app as administrator.
12. Open each changed tool and test the real workflow.

## Example Commands For Current Tools

Build Sample Tool:

```powershell
Push-Location examples\sample_tool_src
tar --format zip -cf ..\sample_tool.tl manifest.json main.tc src
Pop-Location
```

Build Internet Bridge:

```powershell
Push-Location examples\internet_bridge_src
tar --format zip -cf ..\internet_bridge.tl manifest.json main.tc icon.png src
Pop-Location
```

Build DTB General:

```powershell
Push-Location examples\dtb_general_src
tar --format zip -cf ..\dtb_general.tl manifest.json main.tc icon.png src data original
Pop-Location
```

Build Easy Firmware:

```powershell
Push-Location examples\easy_firmware_src
tar --format zip -cf ..\easy_firmware.tl manifest.json main.tc src data
Pop-Location
```

## Final Notes

Good GRTBox tools should feel like practical desktop utilities. Keep them
technical, direct, reliable, and transparent. The best packages are boring in
the right way: predictable UI, clear logs, safe defaults, real source links,
portable code, and no surprise windows.
