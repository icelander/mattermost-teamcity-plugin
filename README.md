# Mattermost TeamCity Plugin

## About

This plugin integrates Mattermost with JetBrains' TeamCity CI/CD software. 

## Installation & Use

1. Download the latest release from the releases page
2. Install it in Mattermost by [following these instructions](https://docs.mattermost.com/administration/plugins.html#custom-plugins)
3. [Create a TeamCity authentication token](https://www.jetbrains.com/help/teamcity/managing-your-user-account.html#ManagingyourUserAccount-ManagingAccessTokens)
4. Run the slash command `/teamcity install <teamcity server> <auth token>` to connect to the TeamCity server 
5. Use one of the following slash commands to interact with TeamCity from within Mattermost:
 	- `/teamcity project list` - List projects with description and project id
 	- `/teamcity build list` - List builds with description, project, and build id
	- `/teamcity build start <project>` - Trigger a build on a specific project
	- `/teamcity build cancel <build_id>` - Cancel a build
	- `/teamcity stats` - Shows agents and the current build queue (if any)

## Configure TeamCity to report build events via webhook

1. [Create an incoming webhook in Mattermost](https://docs.mattermost.com/developer/webhooks-incoming.html), e.g. `https://mattermost.example.com/hooks/kp1zt6uxk3dzxffnzkrjff3e3o`
2. In TeamCity, install the [Web Hooks (tcWebHooks)](https://plugins.jetbrains.com/plugin/8948-web-hooks-tcwebhooks-/) plugin
3. In your build Settings, click `WebHooks`: ![Webhook Link](https://i.imgur.com/9BdzzmG.png)
4. Add a webhook for every build, a specific project, or a specific build ![Add webhook to site, project, or build](https://i.imgur.com/04dlOuc.png)
5. Click "Click to create new WebHook" ![Click to create new webhook](https://i.imgur.com/nDEGmDx.png)
6. Enter the Mattermost webhook URL in the `URL` field and for `Paylod Format` select either `Slack.com JSON templates (JSON)` or `Slack.com Compact Notification (JSON)`
7. Select the build events to post to this webhook ![Webhook config screen](https://i.imgur.com/W9yaOm6.png)
8. Click `Save Web Hook`


