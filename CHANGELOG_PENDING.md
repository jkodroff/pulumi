### Improvements

- [provider/python]: Improved exception display. The traceback is now shorter and it always starts with user code.  
  [#10336](https://github.com/pulumi/pulumi/pull/10336)

- [codegen/schema] Support options.version on pulumi convert

### Bug Fixes

- [cli] Paginate template options
  [#10130](https://github.com/pulumi/pulumi/issues/10130)

- [sdk/dotnet] Fix serialization of non-generic list types.
  [#10277](https://github.com/pulumi/pulumi/pull/10277)

- [sdk/python] Update PyYAML to 6.0

- [cli/watch] `pulumi watch` now uses relies on a program built on [`watchexec`](https://github.com/watchexec/watchexec)
  to implement recursive file watching, improving performance and cross-platform compatibility.
  This `pulumi-watch` program is now included in releases.
  [#10213](https://github.com/pulumi/pulumi/issues/10213)


### Bug Fixes

- [engine/backends]: Fix bug where File state backend failed to apply validation to stack names, resulting in a panic.
  [#10417](https://github.com/pulumi/pulumi/pull/10417)
