# Mattermost TeamCity Plugin

## About

This plugin integrates Mattermost with JetBrains' TeamCity CI/CD appliance. 

## Installation & Use

1. Download the latest release from the releases page
2. Install it in Mattermost by [following these instructions](https://docs.mattermost.com/administration/plugins.html#custom-plugins)
3. Run the slash command `/teamcity install <teamcity server> <teamcity username> <teamcity password>` to connect to the TeamCity server 
4. Use one of the following slash commands to interact with TeamCity from within Mattermost
 	- `/teamcity project list` - List projects with description and project id
 	- `/teamcity build list` - List builds with description, project, and build id
	- `/teamcity build start <project>` - Trigger a build on a specific project
	- `/teamcity build cancel <build_id>` - Cancel a build
	- `/teamcity stats` - Shows agents and the current build queue

## Configure TeamCity to report build events *Coming in a future release*

1. Enter the plugin webhook URL in TeamCity
2. It starts sending build statuses

## To Do

 - [x] Implement TeamCity API
 - [x] Read TeamCity Config from config file
 - [x] Implement slash commands
 	- [x] `/teamcity project list`
 	- [x] `/teamcity build list`
 	- [x] `/teamcity build start`
 	- [x] `/teamcity build cancel`
 	- [x] `/teamcity stats`
 		- [x] Agent Stats
 		- [x] Build Queue
- [ ] Implement Webhook
- [ ] Add Screenshots to Readme