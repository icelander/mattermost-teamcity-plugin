# Mattermost TeamCity Plugin

## About

This plugin integrates Mattermost with JetBrains' TeamCity CI/CD appliance. 

## Installation

1. Download the latest release from the releases page
2. Install it in Mattermost by [following these instructions](https://docs.mattermost.com/administration/plugins.html#custom-plugins)
3. Enable & Configure your TeamCity settings in `System Console` > `Plugins` > `TeamCity`
4. Use one of the following slash commands to interact with TeamCity from within Mattermost
 	- `/teamcity project list` - List projects with description and project id
 	- `/teamcity build list` - List builds with description, project, and build id
	- `/teamcity build status <build_id>` - Get the status of a specific build
	- `/teamcity build start <project>` - Trigger a build on a specific project
	- `/teamcity build cancel <build_id>` - Cancel a build
	- `/teamcity stats` - Basic build statistics (Project Level and Build Configuration level)

## Configure TeamCity to report build status

1. Enter the plugin webhook URL in TeamCity
2. It starts sending build statuses

## TODO

 - [ ] Implement TeamCity API
 - [ ] Read TeamCity Config from config file
 - [ ] Implement slash commands
 	- [ ] `/teamcity project list`
 	- [ ] `/teamcity build list`
 	- [ ] `/teamcity build status`
 	- [ ] `/teamcity build start`
 	- [ ] `/teamcity build start`
 	- [ ] `/teamcity build cancel`
 	- [ ] `/teamcity stats`