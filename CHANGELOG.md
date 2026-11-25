# Changelog

Notable changes are documented below. Earlier releases are available in the repository tags and GitHub Releases.

## [Unreleased]

## [4.0.8] - 2026-11-26

### Changed

- Update the indirect Fiber adapter dependency `github.com/gofiber/schema`
  from v1.10.0 to v1.10.2.
- Refresh GitHub Actions dependencies used by CI and release validation.

The module path, public API and Go 1.27.4 requirement are unchanged.
Fiber remains in the optional adapter; this patch does not add Fiber
to the root package imports.


## [4.0.7] - 2026-11-24


### Changed

- Update `github.com/gofiber/schema` from v1.9.3 to v1.10.0.
- Update `github.com/gofiber/utils/v2` from v2.6.4 to v2.6.5.
- Update `github.com/klauspost/compress` from v1.20.3 to v1.20.5.
- Update `github.com/valyala/fasthttp` from v1.78.0 to v1.80.0.
- Update `golang.org/x/crypto` from v0.58.0 to v0.59.0.
- Update `golang.org/x/net` from v0.62.0 to v0.63.0.
- Update `golang.org/x/sys` from v0.49.0 to v0.50.0.
- Update `golang.org/x/text` from v0.43.0 to v0.44.0.

The module path and the public API are unchanged. The minimum Go version remains 1.27.4.

## [4.0.6] - 2026-11-13

### Changed

- Update `github.com/gofiber/fiber/v3` from v3.6.1 to v3.7.0. An application that selects this release is moved to Fiber v3.7.0 or later.

The module path and the public API are unchanged.

## [4.0.5] - 2026-11-13

### Changed

- Require Go 1.27.4, which carries security fixes to cgo, the go command, the compiler, the linker and `net/http`.

The module path and the public API are unchanged.

## [4.0.4] - 2026-11-01

### Changed

- Update the optional Fiber adapter dependency from Fiber v3.5.0 to v3.6.1 and refresh compatible transitive dependencies.
- Update the optional YAML loader dependency to go.yaml.in/yaml/v3 v3.0.6.
- Pin Go Test Report v1.1.3 and Go Report Card v1.1.4; retain Go 1.27.3, the public API and all translations.

