<div align="center">
	<p><img src="./Glance/Assets.xcassets/AppIcon.appiconset/app-icon-256pt@1x.png" alt="" height="140"></p>
	<h1>Glance</h1>
	<p><strong>All-in-one Quick Look plugin for macOS</strong></p>
	<p>Glance provides Quick Look previews for files that macOS doesn't support out of the box.</p>
</div>

> **Note:** This is a maintained fork of [samuelmeuli/glance](https://github.com/samuelmeuli/glance), which was archived by its original author. This fork updates the project to build and run on modern macOS versions (Ventura, Sonoma, Sequoia) with current Xcode, Swift, and Go toolchains. See [Changes from upstream](#changes-from-upstream) below for details.

## Supported file types

- **Source code** (with [Chroma](https://github.com/alecthomas/chroma) syntax highlighting): `.cpp`, `.js`, `.json`, `.py`, `.swift`, `.yml` and many more

  <p><img src="./AppStore/Assets/Screenshots/ScreenshotSourceCode.png" alt="" width="600"></p>

- **Markdown** (rendered using [goldmark](https://github.com/yuin/goldmark)): `.md`, `.markdown`, `.mdown`, `.mkdn`, `.mkd`, `.Rmd`

  <p><img src="./AppStore/Assets/Screenshots/ScreenshotMarkdown.png" alt="" width="600"></p>

- **Archive**: `.tar`, `.tar.gz`, `.zip`

  <p><img src="./AppStore/Assets/Screenshots/ScreenshotArchive.png" alt="" width="600"></p>

- **Jupyter Notebook**: `.ipynb`

  <p><img src="./AppStore/Assets/Screenshots/ScreenshotJupyterNotebook.png" alt="" width="600"></p>

- **Tab-separated values** (parsed using [SwiftCSV](https://github.com/swiftcsv/SwiftCSV)): `.tab`, `.tsv`

  <p><img src="./AppStore/Assets/Screenshots/ScreenshotTSV.png" alt="" width="600"></p>

## Changes from upstream

This fork was created because the [original project](https://github.com/samuelmeuli/glance) by Samuel Meuli was archived and no longer maintained, while the Quick Look extension remained useful and had no drop-in replacement. The following changes have been made:

- **macOS compatibility** -- Updated deployment target, build settings, and APIs to support macOS 13+ (Ventura, Sonoma, Sequoia) on both Intel and Apple Silicon
- **Go toolchain** -- Upgraded from Go 1.14 to a current release; updated all Go dependencies including the Chroma v0 to v2 migration
- **Jupyter Notebook support** -- Replaced the archived `nbtohtml` dependency with a built-in converter
- **Swift dependencies** -- Updated SwiftCSV, replaced `swift-exec`, and modernized Swift APIs
- **Deprecated APIs** -- Replaced hardcoded system icon paths, WKWebView private API usage, and other deprecated patterns
- **Rebranded identifiers** -- Bundle IDs moved to `io.2075.Glance` namespace for independent signing and distribution

## Building

Xcode, Swift, and Go need to be installed to build the app locally.

```bash
# Clone the repo
git clone https://github.com/2075/glance.git
cd glance

# Open in Xcode
open Glance.xcodeproj
```

Build and run the **Glance** scheme. The Quick Look extension is embedded automatically.

## FAQ

**Why does Glance require network permissions?**

Glance renders some previews in a `WKWebView`. All assets are stored locally and network access is disabled, but web views unfortunately still need the `com.apple.security.network.client` entitlement to function.

**Why are images in my Markdown files not loading?**

Glance blocks remote assets. Furthermore, the app only has access to the file that's being previewed. Local image files referenced from Markdown are therefore not loaded.

**Why isn't [file type] supported?**

Feel free to [open an issue](https://github.com/2075/glance/issues/new) or [contribute](#contributing)! When opening an issue, please describe what kind of preview you'd expect for your file.

Please note that macOS doesn't allow the handling of some file types (e.g. `.plist`, `.ts` and `.xml`).

**You claim to support [file type], but previews aren't showing up.**

Please note that Glance skips previews for large files to avoid slowing down your Mac.

It's possible that your file's extension or [UTI](https://en.wikipedia.org/wiki/Uniform_Type_Identifier) isn't associated with Glance. You can easily verify this:

1. Check whether the file extension is matched to the correct class in [`PreviewVCFactory.swift`](./QLPlugin/Views/PreviewVCFactory.swift).
2. Find your file's UTI by running `mdls -name kMDItemContentType /path/to/your/file`. Check whether the UTI is listed under `QLSupportedContentTypes` in [`Info.plist`](./QLPlugin/Info.plist).
3. If an association is missing, please feel free to add it and submit a PR.

## Contributing

Suggestions and contributions are always welcome! Please discuss larger changes (e.g. adding support for a new file type) via issue before submitting a pull request.

To add previews for a new file extension, please follow these steps:

1. Create a new class for your file type in [this directory](./QLPlugin/Views/Previews/). It should implement the `Preview` protocol. See the other files in the directory for examples.
2. Match the file extension to your class in [`PreviewVCFactory.swift`](./QLPlugin/Views/PreviewVCFactory.swift).
3. Find your file's UTI by running `mdls -name kMDItemContentType /path/to/your/file`. Add it to `QLSupportedContentTypes` in [`Info.plist`](./QLPlugin/Info.plist).
4. Update this README and [`SupportedFilesWC.xib`](Glance/SupportedFilesWC.xib) accordingly.

## Credits

Originally created by [Samuel Meuli](https://github.com/samuelmeuli). This fork is maintained by [2075](https://github.com/2075).

Licensed under the [MIT License](./LICENSE.md).
