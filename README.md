# GRTBox

**GRTBox** (Gambiarra R36S Toolbox) is a desktop toolbox environment designed to organize, install, and run useful tools for Linux-based retro handhelds, with an initial focus on the R36S family and compatible devices.

It is not a firmware, not an operating system, and not a replacement for the community projects that already power these handhelds. GRTBox is a companion application: a central place where users can find and run utilities that would otherwise be scattered across guides, forums, repositories, scripts, and separate programs.

The goal is simple:

```text
Open GRTBox.
Choose a tool.
Run it.
```

GRTBox tries to make technical maintenance tasks feel more like using a normal desktop application.

---

## Why GRTBox Exists

Retro handhelds such as the R36S and similar devices are popular because they are inexpensive, flexible, and supported by a very active community. At the same time, they can be confusing for new users.

A user may need to deal with:

```text
Different firmware builds
Different hardware revisions
Different clone models
Different SD card layouts
Different DTB files
Different compatibility notes
Different community tools
Different installation methods
Different Windows utilities
```

Many useful tools exist, but they are often disconnected from each other. Some are web pages, some are scripts, some are downloadable archives, some are command-line utilities, and some are buried inside long tutorials.

GRTBox exists to reduce that fragmentation.

It provides a common desktop home for these utilities and presents them as installable tools with names, descriptions, icons, versions, and a consistent launcher interface.

---

## What GRTBox Is

GRTBox is a Windows-focused desktop toolbox launcher and runtime.

It can:

```text
Install tool packages
Display installed tools
Search installed tools
Open tools
Remove tools
Load tool icons and metadata
Run TC-based tools inside the application
Show logs and errors
Access a simple remote Tool Store
```

The application is built around the idea that each utility should be packaged as a `.tl` file.

A `.tl` package is a tool package. The user does not need to manually extract it, inspect it, or place files in special folders. GRTBox handles the installation and launching process.

---

## What GRTBox Is Not

GRTBox is not intended to replace the systems or projects used by retro handhelds.

It is not:

```text
A firmware
An emulator frontend
A ROM manager
A game launcher
A replacement for ArkOS, ROCKNIX, or similar systems
A replacement for EmulationStation
A replacement for community documentation
```

Instead, it works beside those projects.

It is meant to be a practical desktop utility that helps with tasks around the device, not a system that runs on the handheld itself.

---

## Main Idea

The main idea behind GRTBox is modularity.

The core application stays focused on being a toolbox host. Tools are distributed separately as `.tl` packages.

This means GRTBox can grow over time without turning into one large, hard-coded application. New tools can be added, replaced, updated, or removed independently.

In practical terms:

```text
GRTBox = the toolbox
.tl package = one tool inside the toolbox
Tool Store = a simple way to discover and install tools
```

This keeps the project flexible and makes it easier to add new utilities as the community discovers new needs.

---

## User Experience

GRTBox is designed to feel like a simple desktop toolbox.

The interface uses:

```text
A dark theme
A fixed sidebar
A Tools page
A Tool Store page
A search bar
Large tool cards
Clear descriptions
Simple buttons
A technical, clean visual style
```

The main background color is:

```text
#2d2d2d
```

The visual direction is intentionally simple. GRTBox is not trying to look like a gaming dashboard full of effects. It is meant to feel like a focused technical utility.

---

## Tool Cards

Installed tools are shown as large cards.

A typical card may show:

```text
Tool icon
Tool name
Version
Short description
Open button
Details button
Remove button
```

The information comes from the tool package itself.

If a tool does not include an icon, GRTBox uses a default internal icon.

This keeps tool packaging simple while still allowing more polished tools to provide their own visual identity.

---

## Search

GRTBox includes search so users can quickly find tools when the toolbox grows.

Search can be based on information such as:

```text
Tool name
Tool ID
Author
Description
```

This is important because the long-term goal is for GRTBox to support many small utilities, not just a few built-in tools.

---

## Tool Store

GRTBox includes a simple Tool Store concept.

The Tool Store is not a complex marketplace. It is a straightforward remote list of `.tl` package links.

GRTBox reads a remote `tools.json` file. That JSON file contains only links to `.tl` packages.

Example:

```json
{
  "tools": [
    "https://example.com/tools/example_tool.tl",
    "https://example.com/tools/another_tool.tl"
  ]
}
```

The JSON does not duplicate metadata.

The name, version, author, description, icon, and technical details are read from each package's own `manifest.json`.

This keeps the store simple and avoids the common problem of duplicated information becoming outdated.

The package itself is the source of truth.

---

## How Tool Store Metadata Works

When the Tool Store is opened, GRTBox follows this process:

```text
Download tools.json
Read the list of .tl links
Load each .tl package
Read its manifest.json
Load icon.png if available
Show the tool as a store card
```

This allows the remote store to stay extremely simple.

The store only needs to know where each package is. Each package describes itself.

---

## First Boot Experience

GRTBox can include a small set of default tools on first launch.

The idea is that a new user should not open an empty application. A fresh installation should already have useful tools available or ready to install.

After that, users can install more tools through the Tool Store or by importing `.tl` packages manually.

---

## The .tl Package Format

A `.tl` file is a ZIP archive renamed with the `.tl` extension.

It is the standard package format for GRTBox tools.

A basic `.tl` package contains:

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
```

The exact internal structure can vary depending on the tool.

The important rule is that each package must include a manifest and an entry file.

---

## manifest.json

The `manifest.json` file describes the tool.

It usually contains information such as:

```text
ID
Name
Version
Author
Description
Entry file
Runtime type
Optional icon
Optional permissions
Optional minimum GRTBox version
```

Example:

```json
{
  "id": "example_tool",
  "name": "Example Tool",
  "version": "1.0.0",
  "author": "GRTBox",
  "description": "A simple example tool.",
  "entry": "main.tc",
  "runtime": "tc",
  "icon": "icon.png"
}
```

For normal users, this file is invisible. GRTBox reads it automatically.

For tool creators, it is how the package identifies itself.

---

## main.tc

The `main.tc` file is the tool's entry point.

It is not a static JSON interface and it is not a separate executable file.

It is code that runs inside the GRTBox TC runtime.

A tool can also contain additional `.tc` files and import them from `main.tc`.

For example:

```text
main.tc
src/ui.tc
src/logs.tc
src/actions.tc
```

This allows tools to stay organized without needing to be compiled into external `.exe` files.

---

## TC Runtime

GRTBox includes a small runtime for TC tools.

A TC tool runs inside the GRTBox environment and can use generic runtime primitives.

These primitives are not specific to one tool. They are general desktop capabilities.

Examples include:

```text
UI rendering
File access
Dialogs
Logs
Timers
HTTP requests
Environment information
Shell or process execution
PowerShell access on Windows
```

This design keeps GRTBox flexible.

The core application does not need to know how every tool works. Each tool contains its own logic, while GRTBox provides the environment needed to run it.

---

## Generic Runtime Philosophy

GRTBox avoids hard-coding tool-specific APIs into the core application.

The core does not need to provide a special API for every possible feature.

Instead, it provides generic building blocks.

A tool can use those building blocks to implement its own behavior.

In simple terms:

```text
GRTBox provides the runtime.
The tool provides the feature.
```

---

## Installation Model

When a `.tl` package is installed, GRTBox copies it into the user's application data folder.

On Windows, installed packages are stored in a location similar to:

```text
%APPDATA%\GRTBox\tools
```

When a tool is opened, GRTBox extracts it into a managed folder, then runs its `main.tc`.

This allows tools to keep their internal files together while still being easy to install and remove.

---

## Updating Tools

If a user installs a package with the same tool ID as an existing installed tool, GRTBox can treat it as an update.

The usual process is:

```text
Detect existing tool ID
Ask whether to update
Replace the installed package
Clear the extracted copy
Run the new version next time
```

The tool ID comes from `manifest.json`.

This avoids relying on file names, which can change from one download to another.

---

## Removing Tools

Tools can be removed from inside GRTBox.

Removing a tool deletes the installed package from the local tools folder.

GRTBox may also clear extracted files associated with that tool.

This keeps the user's toolbox clean and avoids old extracted versions accumulating unnecessarily.

---

## Error Handling

GRTBox is designed to show clear errors when something goes wrong.

Examples:

```text
Invalid .tl package
Missing manifest.json
Missing main.tc
Invalid manifest
Tool crashed
Tool failed to load
Tool requires administrator privileges
Tool Store could not be reached
```

If a tool crashes, GRTBox should not crash with it. The user should be able to go back to the Tools page and continue using the application.

---

## Administrator Privileges

Some tools may require administrator privileges, especially tools that interact with Windows system settings, disks, network configuration, or low-level device operations.

GRTBox can identify this through the tool manifest.

When a tool requires elevated privileges, the application should clearly inform the user.

The goal is to avoid silent failures. If a tool cannot perform an operation because Windows requires elevation, the user should see a direct and understandable message.

---

## Project Focus

GRTBox is currently focused on Windows.

That focus makes sense because many users manage their SD cards, downloads, firmware files, and utilities from a Windows PC.

The internal architecture, however, is designed in a way that may allow future expansion.

The project can evolve over time, but the initial goal is to provide a strong Windows experience first.

---

## Who GRTBox Is For

GRTBox is intended for:

```text
R36S users
R36H users
Users of compatible Linux handhelds
People who manage firmware and SD cards on Windows
People who want a simpler desktop workflow
Community tool creators
Users who prefer visual tools over command-line steps
```

It is especially useful for people who like these devices but do not want to manually follow long technical procedures every time they need to perform maintenance.

---

## For Regular Users

If you are a regular user, GRTBox is meant to be simple.

You should be able to:

```text
Install GRTBox
Open it
Choose a tool
Read what the tool does
Click Open
Follow the on-screen instructions
```

You do not need to understand the `.tl` format to use GRTBox.

You do not need to manually edit package files.

You do not need to know how the TC runtime works.

Those details exist so the project can grow, but the user experience should remain straightforward.

---

## For Tool Creators

If you want to create tools for GRTBox, the `.tl` format provides a simple packaging model.

A tool can include:

```text
A manifest
A main TC file
Additional TC modules
Images
Data files
Assets
```

This allows community utilities to be distributed in a format that GRTBox can install and present cleanly.

Technical details for creating tools should live in the Dev Guide, not in this README.

This README is meant to introduce the project.

---

## Local Tools and Remote Tools

GRTBox supports two basic ways to get tools:

```text
Local installation
Tool Store installation
```

Local installation means selecting a `.tl` file manually from the computer.

Tool Store installation means browsing tools from a remote list and installing them directly.

Both methods use the same package format.

This means a tool can be tested locally first and later published through the store without changing its internal structure.

---

## Design Principles

GRTBox follows a few simple design principles.

### Keep the core simple

The main application should not become overloaded with every possible feature.

It should focus on loading, organizing, and running tools.

### Keep tools modular

Each tool should be packaged separately.

This makes updates and experimentation easier.

### Keep the interface practical

The UI should be clean, readable, and useful.

It should avoid unnecessary effects that make the application harder to use.

### Avoid duplicated metadata

Tool information should come from the tool package itself.

If a package says its name is "Example Tool", the store and installed tools page should read that from the package manifest.

### Make advanced tasks feel approachable

Many handheld maintenance tasks are technical. GRTBox should make them easier to understand without hiding important warnings.

---

## Current Status

GRTBox is in active development.

The project already has the foundation for:

```text
A desktop interface
Tool package loading
TC runtime execution
Installed tool management
Example tools
Tool metadata handling
A modular expansion model
```

Some tools may be more mature than others.

The project is evolving quickly, and the structure is being refined as real tools are tested.

---

## Roadmap

Future improvements may include:

```text
Improved Tool Store interface
More tool categories
Better update detection
More package validation
Better crash recovery
Improved logs
Optional online databases for tool data
More polished default icons
More complete user documentation
Better support for non-technical users
```

The long-term goal is to make GRTBox a reliable hub for useful retro handheld utilities.

---

## Repository Structure

A typical project structure may include:

```text
docs/
examples/
frontend/
internal/
tools/
```

The exact structure may change as the project evolves.

For technical implementation details, refer to the development guide.

---

## Building From Source

This README is primarily for users and visitors who want to understand the project.

Developers who want to build from source, modify the runtime, or create tools should read the Dev Guide.

The Dev Guide should contain:

```text
Development setup
Build commands
TC runtime details
Package validation rules
Tool creation examples
Frontend architecture
Backend architecture
```

Keeping those details separate makes the main README easier to understand.

---

## Safety Notes

Some tools may perform operations that affect devices, disks, network settings, or system configuration.

Before using any tool that modifies a disk or system setting, read the tool's instructions carefully.

GRTBox should always try to show warnings before destructive actions.

Users should still verify that they selected the correct device, disk, or file before continuing.

---

## Community

GRTBox is designed around the idea that community tools should be easier to discover and use.

The project can grow through:

```text
Bug reports
Feature suggestions
New tool packages
Improved documentation
Compatibility information
Testing on different devices
```

The more organized the ecosystem becomes, the easier it is for new users to enter it.

---

## Vision

The long-term vision for GRTBox is to become a central toolbox for retro handheld maintenance and utility workflows.

Instead of searching across many different places for small utilities, users should be able to open one application and find the tools they need.

GRTBox aims to become that place:

```text
A clean toolbox.
A simple launcher.
A modular runtime.
A home for community utilities.
```
