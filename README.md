# GRTBox

GRTBox is a Windows-focused desktop toolbox launcher built with Go, Wails, and Svelte. It installs `.tl` packages, extracts them into the user profile, and runs their TC modules through a small desktop runtime.

The UI is intentionally simple: a fixed dark sidebar, a Tools page, search, large launcher cards, and a technical desktop-toolbox look. The main app background is `#2d2d2d`.

On Windows, GRTBox must be started as administrator. If it is launched without elevated privileges, it shows an administrator requirement message and exits before opening the desktop UI. This is intentional because the core toolbox workflows depend on elevated Windows APIs.

## Package Model

A `.tl` file is a ZIP archive renamed to `.tl`. It is not an external executable and it is not a declarative JSON UI document.

Every package must contain:

```text
manifest.json
main.tc
```

Packages may also contain:

```text
icon.png
src/*.tc
assets/*
```

`main.tc` is real TC code. TC uses JavaScript module syntax for now, including relative imports:

```js
import { renderApp } from "./src/ui.tc";

export default async function main(runtime) {
  await renderApp(runtime);
}
```

## Manifest

Minimum `manifest.json`:

```json
{
  "id": "my_tool",
  "name": "My Tool",
  "version": "1.0.0",
  "author": "Tool Author",
  "description": "Short launcher-card description.",
  "entry": "main.tc",
  "runtime": "tc"
}
```

Common optional fields:

```json
{
  "icon": "icon.png",
  "requires_admin": true,
  "permissions": ["powershell.exec", "logs.write"],
  "target_platforms": ["windows"],
  "min_grtbox_version": "0.1.0",
  "package_format_version": "1.0.0"
}
```

Validation checks:

- Package extension is `.tl`.
- Package is a valid ZIP archive.
- `manifest.json` exists and is valid JSON.
- `main.tc` exists.
- Manifest includes `id`, `name`, `version`, `entry`, and `runtime`.
- `runtime` must be exactly `tc`.
- `entry` must be a portable relative path ending in `.tc`.
- The entry file must exist inside the package.
- ZIP paths must use forward slashes and must not escape the package root.
- Case-insensitive duplicate entries are rejected.
- `icon.png`, when present, must be a valid PNG.
- Tool IDs must be unique in the installed tools directory.

## Install And Extraction

Installed packages are copied to:

```text
%APPDATA%\GRTBox\tools
```

When a tool is opened, GRTBox extracts it to:

```text
%APPDATA%\GRTBox\extracted_tools\<tool_id>\<version>
```

The TC loader reads modules from the extracted copy. Relative imports are resolved inside the extracted package so the same `.tl` can run on every GRTBox installation.

If a selected `.tl` package has the same tool ID as a package already in the library, GRTBox asks whether the user wants to update that tool. Confirming replaces the installed package and clears the extracted copy so the next launch uses the new `.tl`.

## TC Runtime

TC tools receive a generic runtime object:

```js
export default async function main(runtime) {
  const h = runtime.ui.createElement;
  runtime.ui.render(h("div", { className: "tc-panel" }, [
    h("h1", {}, "Hello"),
    h("button", { onClick: () => runtime.logs.write("Clicked") }, "Run")
  ]));
}
```

Runtime primitives:

- `ui`
- `os`
- `process`
- `shell`
- `powershell`
- `filesystem`
- `path`
- `env`
- `dialogs`
- `logs`
- `timers`
- `http`
- `json`
- `events`

The runtime exposes generic desktop primitives only. It does not include Internet Bridge-specific APIs, adapter-specific APIs, diagnostics APIs, or other feature APIs. Tools build their own behavior using the generic primitives.

## Crash Handling

If a TC tool throws during startup or runtime loading, GRTBox shows:

- `Tool crashed`
- the error message
- `Reload Tool`
- `Back to Tools`

The crash is also written to the app log.

## Example Packages

`examples/sample_tool_src` is a minimal TC package.

`examples/internet_bridge_src` is the first real TC tool. It is Windows-first and uses generic PowerShell/process primitives to:

- detect network adapters with `Get-NetAdapter` and `Get-NetIPConfiguration`
- check admin status
- run network diagnostics
- inspect and control Windows Internet Connection Sharing through `HNetCfg.HNetShare`
- render its UI from TC code
- write runtime logs from real state

It does not rely on feature-specific GRTBox APIs.

`examples/dtb_general_src` combines the original DTB Identify and DTB File Data Extractor browser tools by Aeolusux into one TC package named DTB General. It uses generic file dialogs and filesystem primitives to identify known R36S panel DTBs and export DTB text data. The original HTML/HTM files and reference files are included inside the package for attribution and traceability.

`examples/easy_firmware_src` is a Windows-first firmware helper for compatible R36S, R36H and related RK3326 handhelds. It uses official firmware release links from the bundled `data/firmware.json`, supports checksum verification when upstream hashes are available, extracts supported archive formats, and keeps flashing behind explicit administrator and destructive-action safety checks.

Build the example packages with PowerShell:

```powershell
Push-Location examples\sample_tool_src
tar --format zip -cf ..\sample_tool.tl manifest.json main.tc src
Pop-Location

Push-Location examples\internet_bridge_src
tar --format zip -cf ..\internet_bridge.tl manifest.json main.tc icon.png src
Pop-Location

Push-Location examples\dtb_general_src
tar --format zip -cf ..\dtb_general.tl manifest.json main.tc icon.png src data original
Pop-Location

Push-Location examples\easy_firmware_src
tar --format zip -cf ..\easy_firmware.tl manifest.json main.tc src data
Pop-Location
```

## Development

Run backend tests:

```powershell
go test ./...
```

Run frontend checks and build:

```powershell
cd frontend
npm install
npm run check
npm run build
```

Build the Windows desktop app:

```powershell
wails build -clean -platform windows/amd64
```

The executable is generated at:

```text
build\bin\GRTBox.exe
```
