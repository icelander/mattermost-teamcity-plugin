package main

import (
	"encoding/csv"
	"fmt"
	"net/url"
	"strings"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"

	"github.com/Cardfree/teamcity-sdk-go/teamcity"
)

const (
	commandTriggerHooks        = "teamcity"
	commandTriggerEnable       = "enable"
	commandTriggerDisable      = "disable"
	commandTriggerInstall      = "install"
	commandTriggerList         = "list"
	commandTriggerListProjects = "projects"
	commandTriggerListBuilds   = "builds"
	commandTriggerBuild        = "build"
	commandTriggerBuildStart   = "start"
	commandTriggerBuildCancel  = "cancel"
	commandTriggerStats        = "stats"

	errorNotInstalled = "To use the TeamCity Plugin first install it with `/teamcity install <teamcity url> <username> <password>`"
	errorDisabled     = "TeamCity Plugin disabled. First enable it with `/teamcity enable`"

	msgInstalled = "TeamCity Plugin Installed!"
	msgEnabled   = "TeamCity Plugin Enabled"
	msgDisabled  = "TeamCity Plugin Disabled"

	commandDialogHelp = "Use one of the following slash commands to interact with TeamCity from within Mattermost\n" +
		"- `/teamcity install <teamcity url> <username> <password>` - Set up the TeamCity plugin\n" +
		"- `/teamcity project list` - List projects with description and project id\n" +
		"- `/teamcity build list` - List builds with description, project, and build id\n" +
		"- `/teamcity build status <build_id>` - Get the status of a specific build\n" +
		"- `/teamcity build start <project>` - Trigger a build on a specific project\n" +
		"- `/teamcity build cancel <build_id>` - Cancel a build\n" +
		"- `/teamcity stats` - Basic build statistics (Project Level and Build Configuration level)"
)

func (p *Plugin) registerCommands() error {
	if err := p.API.RegisterCommand(&model.Command{
		Trigger:          commandTriggerHooks,
		DisplayName:      "TeamCity",
		Description:      "Integration with JetBeans TeamCity",
		AutoComplete:     true,
		AutoCompleteHint: "[command]",
		AutoCompleteDesc: "Available commands: install, list, build, stats",
	}); err != nil {
		return errors.Wrapf(err, "failed to register %s command", commandTriggerHooks)
	}

	if err := p.API.RegisterCommand(&model.Command{
		Trigger:          "/" + commandTriggerHooks,
		DisplayName:      "TeamCity",
		Description:      "Integration with JetBeans TeamCity",
		AutoComplete:     true,
		AutoCompleteHint: "[command]",
		AutoCompleteDesc: "Available commands: install, list, build, stats",
	}); err != nil {
		return errors.Wrapf(err, "failed to register %s command", commandTriggerHooks)
	}

	return nil
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {

	cArgs, err := p.extractCommandArgs(args.Command)

	if err != nil {
		return p.postEphemeral("Could not extract command arguments"), nil
	}

	if cArgs[0] == "/"+commandTriggerHooks {
		return p.executeCommandHooks(args), nil
	}

	return &model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
		Text:         fmt.Sprintf("Unknown command: " + args.Command),
	}, nil
}

func (p *Plugin) extractCommandArgs(cArgs string) ([]string, error) {
	r := csv.NewReader(strings.NewReader(cArgs))
	r.Comma = ' ' // space
	fields, err := r.Read()

	return fields, err
}

func (p *Plugin) invalidCommand(args *model.CommandArgs) *model.CommandResponse {
	return &model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
		Text:         commandDialogHelp,
	}
}

func (p *Plugin) executeCommandHooks(args *model.CommandArgs) *model.CommandResponse {
	configuration := p.getConfiguration()

	cArgs, err := p.extractCommandArgs(args.Command)

	if err != nil {
		return &model.CommandResponse{
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			Text:         "There was an error processing that command = `" + args.ToJson() + "`",
		}
	}

	if !strings.HasPrefix(args.Command, "/"+commandTriggerHooks) {
		return &model.CommandResponse{
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			Text:         "Invalid Command: " + cArgs[0],
		}
	}

	if !configuration.installed && cArgs[1] != commandTriggerInstall {
		return p.postEphemeral(errorNotInstalled)
	}

	switch cArgs[1] {
	case commandTriggerDisable:
		return p.executeCommandTriggerDisable(args)
	case commandTriggerEnable:
		return p.executeCommandTriggerEnable(args)
	case commandTriggerInstall:
		return p.executeCommandTriggerInstall(args)
	case commandTriggerList:
		if configuration.disabled {
			return p.postEphemeral(errorDisabled)
		}
		switch cArgs[2] {
		case commandTriggerListBuilds:
			return p.executeCommandTriggerListBuilds(args)
		case commandTriggerListProjects:
			return p.executeCommandTriggerListProjects(args)
		default:
			return p.invalidCommand(args)
		}

	case commandTriggerBuild:
		if configuration.disabled {
			return p.postEphemeral(errorDisabled)
		}
		switch cArgs[2] {
		case commandTriggerBuildStart:
			return p.executeCommandTriggerBuildStart(args)
		case commandTriggerBuildCancel:
			return p.executeCommandTriggerBuildCancel(args)
		default:
			return p.invalidCommand(args)
		}

	default:
		return &model.CommandResponse{
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			Text: "**cArgs[1]: ** `" + cArgs[1] + "`\n" +
				"**cArgs: ** `" + strings.Join(cArgs, " - ") + "`\n" +
				"**args:** `" + args.ToJson() + "`\n",
		}
	}
}

func (p *Plugin) postEphemeral(text string) *model.CommandResponse {
	return &model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
		Text:         text,
	}
}

func (p *Plugin) executeCommandTriggerInstall(args *model.CommandArgs) *model.CommandResponse {

	configuration := p.getConfiguration()

	cArgs, err := p.extractCommandArgs(args.Command)

	if err != nil {
		p.API.LogError(
			"Could not extract arguments",
			"error", err.Error(),
		)
		return p.postEphemeral("Could not read arguments")
	}

	// Install command is like this:
	//  - [0] : /teamcity
	//  - [1] : install
	//  - [2] : url
	//  - [3] : username
	//  - [4] : password
	// TODO: Write test for not enough arguments
	if len(cArgs) < 5 {
		return p.postEphemeral(errorNotInstalled)
	}

	// Validate URL
	// TODO: Write invalid URL test
	u, err := url.ParseRequestURI(cArgs[2])
	if err != nil {
		// p.API.LogError("Invalid TeamCity URL: " + u.String())
		return p.postEphemeral("Invalid URL: `" + err.Error() + "`\n")
	}

	// Attempt a test command
	// Fourth argument is API version, setting is based on recommendation from here:
	// https://www.jetbrains.com/help/teamcity/rest-api.html#RESTAPI-RESTAPIVersions
	// TODO: Write "could not connect" test
	client := teamcity.New(u.String(), cArgs[3], cArgs[4], "2018.1")
	server, err := client.Server()

	if err != nil {
		return p.postEphemeral("Could not connect to server.\nError: `" + err.Error() + "`")
	}

	configuration.teamCityURL = u.String()
	configuration.teamCityUsername = cArgs[3]
	configuration.teamCityPassword = cArgs[4]
	configuration.installed = true
	p.setConfiguration(configuration)

	// Return an error if it fails
	return p.postEphemeral("TeamCity Installed! Here are the server details:\n" +
		"**Server:** " + u.String() + "\n" +
		"**Server Version:** " + server.Version + "\n" +
		"**Build Number:** " + server.BuildNumber)
}

func (p *Plugin) executeCommandTriggerEnable(args *model.CommandArgs) *model.CommandResponse {
	configuration := p.getConfiguration()
	
	if configuration.disabled {
		configuration.disabled = false
		p.setConfiguration(configuration)	
	}

	return p.postEphemeral(msgEnabled)
}

func (p *Plugin) executeCommandTriggerDisable(args *model.CommandArgs) *model.CommandResponse {
	configuration := p.getConfiguration()
	if !configuration.disabled {
		configuration.disabled = true
		p.setConfiguration(configuration)	
	}

	return p.postEphemeral(msgDisabled)
}

func (p *Plugin) executeCommandTriggerListProjects(args *model.CommandArgs) *model.CommandResponse {
	return &model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
		Text:         "Running List Projects Command",
	}
}

func (p *Plugin) executeCommandTriggerListBuilds(args *model.CommandArgs) *model.CommandResponse {
	return &model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
		Text:         "Running List Builds Command",
	}
}

func (p *Plugin) executeCommandTriggerBuildStart(args *model.CommandArgs) *model.CommandResponse {
	return &model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
		Text:         "Running Build Start Command",
	}
}

func (p *Plugin) executeCommandTriggerBuildCancel(args *model.CommandArgs) *model.CommandResponse {
	return &model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
		Text:         "Running Build Cancel Command",
	}
}

func (p *Plugin) executeCommandTriggerStats(args *model.CommandArgs) *model.CommandResponse {
	return &model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
		Text:         "Running Stats Command",
	}
}
