# README

## About

This is refactor Web App to Dekstop App `go-web-code-interactive`

## Live Development

To run in live development mode, run `wails dev` in the project directory. This will run a Vite development
server that will provide very fast hot reload of your frontend changes. If you want to develop in a browser
and have access to your Go methods, there is also a dev server that runs on http://localhost:34115. Connect
to this in your browser, and you can call your Go code from devtools.

## Building

To build a redistributable, production mode package, use `wails build`.

- Windows
  - Need Webview Microsoft Edge
- UI
  - [x] PHP Tested
  - [x] Width CodeMirror Not Fixed

## Bundled Runtime Execution

Place any language runtime you want to ship with the app under `internal/runtimebundles/runtimes/<runtime-name>/`. Each folder should contain the executable for that language (for example `php` on Linux or `php.exe` on Windows). During startup the app extracts these folders into a temporary directory, lists them with `ListBundledRuntimes`, and removes them again on shutdown.

Set the `execMode` field when calling `RunFileExecutable` to pick how code is run:

- `bundled`: always use the packaged interpreter from `bundledRuntime`.
- `custom`: use the binary pointed to by `customExecutable` and (optionally) `customWorkingDir`.
- `default` (or empty): run the interpreter found on the system, but fall back to a bundled runtime when `preferBundled` is `true`.

### Example

To run PHP scripts with a bundled runtime called `php-8.2`, create this structure before building:

```
internal/runtimebundles/runtimes/php-8.2/php
```

Then send the following payload from the frontend:

```json
{
  "language": "php",
  "txt": "<?php echo 'Hello from bundled PHP!';",
  "tipe": "repl",
  "execMode": "bundled",
  "bundledRuntime": "php-8.2"
}
```

To run a local PHP build instead, switch to custom mode:

```json
{
  "language": "php",
  "txt": "<?php echo 'Hello from custom PHP!';",
  "tipe": "repl",
  "execMode": "custom",
  "customExecutable": "/opt/php/bin/php",
  "customWorkingDir": "."
}
```

These options make it clear to the backend whether it should use the packaged runtime, a user-specified binary, or the system’s default interpreter with an optional bundled fallback.
