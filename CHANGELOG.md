# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## 1.0.0 - 2020-01-27
### Added
 - System Console settings page for plugin

### Changed
 - Changed format of build list to tables
 - Configurable limit to number of builds returned, default 5
 - Returns help text on error or invalid command

### Fixed
 - `/build cancel` now works with no arguments
 - `/build start` returns arguments in correct order
 - `/build cancel` has formatted dates


## 0.3.0 - 2020-01-24
### Added
- `/teamcity stats` - Now includes information about current build queue (if present)

## 0.2.0 - 2020-01-23
### Added
- `/teamcity stats` - Displays build agent status

## 0.1.0 - 2020-01-19
### Added
- `/teamcity enable`
- `/teamcity disable`
- `/teamcity install`
- `/teamcity list projects`
- `/teamcity list builds`
- `/teamcity build start`
- `/teamcity build cancel`
