# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

## [0.1.3] - 2026-02-27

### Changed
- Automate release workflow via GitHub Actions: tags and GitHub Releases are now created automatically when a `release/*` branch PR is merged into `main`

## [0.1.2] - 2026-02-11

### Changed
- E2B backend now supports token acquisition via GET /sandboxes/{id}, removing the limitation that tokens were only available at instance creation time
- Unified token recovery logic for both Cloud and E2B backends when token cache is missing

## [0.1.1] - 2026-01-20

### Changed
- Separate control plane and data plane with token caching

## [0.1.0] - 2026-01-16

### Added
- Initial release
- Update module path to github.com/TencentCloudAgentRuntime/ags-cli
- Replace all git.woa.com references with github.com/TencentCloudAgentRuntime/ags-go-sdk v0.0.10
