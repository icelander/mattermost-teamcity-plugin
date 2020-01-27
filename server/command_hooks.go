package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"

	"github.com/icelander/teamcity-sdk-go/teamcity"
	"github.com/icelander/teamcity-sdk-go/types"

	"github.com/olekukonko/tablewriter"
)

const (
	configTeamCityVersion = "2018.1"
	// Magic Date: Mon Jan 2 15:04:05 MST 2006
	fmtDateTime = "Jan 2, 2006 3:04 PM MST"

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

	errorNotInstalled   = "To use the TeamCity Plugin first install it with `/teamcity install <teamcity url> <token>`"
	errorDisabled       = "TeamCity Plugin disabled. First enable it with `/teamcity enable`"
	errorWhatList       = "Try `/teamcity list builds` or `/teamcity list projects`"
	errorNoBuildID      = "Please provide a build ID, `/teamcity build start <build_id>`"
	errorNoBuildCommand = "Please provide a build command, e.g. `/teamcity build start <build_id>`"

	msgInstalled = "TeamCity Plugin Installed!"
	msgEnabled   = "TeamCity Plugin Enabled"
	msgDisabled  = "TeamCity Plugin Disabled"

	iconGood = ":white_check_mark:"
	iconBad  = ":x:"

	commandDialogHelp = "Use one of the following slash commands to interact with TeamCity from within Mattermost\n" +
		"- `/teamcity install <teamcity url> <token>` - Set up the TeamCity plugin\n" +
		"- `/teamcity list projects` - List projects with description and project id\n" +
		"- `/teamcity list builds` - List builds with description, project, and build id\n" +
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

// ExecuteCommand executes the slash commands
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
			Text:         commandDialogHelp,
		}
	}

	if !strings.HasPrefix(args.Command, "/"+commandTriggerHooks) {
		return &model.CommandResponse{
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			Text:         commandDialogHelp,
		}
	}

	if !configuration.Installed() && cArgs[1] != commandTriggerInstall {
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
		if len(cArgs) == 2 {
			return p.postEphemeral(errorWhatList)
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
		if len(cArgs) == 3 {
			return p.postEphemeral(errorNoBuildCommand)
		}
		switch cArgs[2] {
		case commandTriggerBuildStart:
			return p.executeCommandTriggerBuildStart(cArgs[3])
		case commandTriggerBuildCancel:
			return p.executeCommandTriggerBuildCancel(args)
		default:
			return p.invalidCommand(args)
		}

	case commandTriggerStats:
		if configuration.disabled {
			return p.postEphemeral(errorDisabled)
		}
		return p.executeCommandTriggerStats(args)

	default:
		return &model.CommandResponse{
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			Text:         commandDialogHelp,
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
	//  - [3] : token

	// TODO: Write test for not enough arguments
	if len(cArgs) != 4 {
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
	// TODO: Write "could not connect" test
	client := teamcity.New(u.String(), cArgs[3])
	server, err := client.Server()

	if err != nil {
		return p.postEphemeral("Could not connect to server.\nError: `" + err.Error() + "`")
	}

	configuration.TeamCityURL = u.String()
	configuration.TeamCityToken = cArgs[3]

	p.setConfiguration(configuration)

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
	configuration := p.getConfiguration()
	client := teamcity.New(configuration.TeamCityURL,
		configuration.TeamCityToken)

	projects, err := client.GetShortProjects()

	if err != nil {
		return p.postEphemeral(fmt.Sprintf("Error listing projects: %s", err.Error()))
	}

	if len(projects) == 0 {
		return p.postEphemeral("No projects found")
	}

	message := "**TeamCity Projects:**\n\n"

	for _, project := range projects {
		if project.ID == "_Root" {
			continue
		}
		message += fmt.Sprintf(" - [%s (ID: %s)](%s)", project.Name, project.WebURL, project.ID)
	}

	return &model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_IN_CHANNEL,
		Text:         message,
	}
}

func (p *Plugin) executeCommandTriggerListBuilds(args *model.CommandArgs) *model.CommandResponse {
	configuration := p.getConfiguration()
	client := teamcity.New(configuration.TeamCityURL,
		configuration.TeamCityToken)

	builds, err := client.GetBuilds()

	if err != nil {
		return p.postEphemeral("Error listing builds: `" + err.Error() + "`")
	}

	if len(builds) == 0 {
		return p.postEphemeral("No builds found")
	}

	message := "**TeamCity Builds:**\n\n"

	buf := new(bytes.Buffer)
	buildTable := tablewriter.NewWriter(buf)
	buildTable.SetAutoWrapText(false)
	buildTable.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	buildTable.SetCenterSeparator("|")
	buildTable.SetHeader([]string{"Build", "Project", "Build Start", "Build Finish"})
	buildTable.SetAutoFormatHeaders(false)
	buildTable.SetHeaderAlignment(tablewriter.ALIGN_CENTER)

	maxBuilds := configuration.GetMaxBuilds()

	if len(builds)-1 < maxBuilds {
		maxBuilds = len(builds) - 1
	}

	for i := 0; i < maxBuilds; i++ {
		build := builds[i]

		nameLink := fmt.Sprintf("[%s #%s](%s)", build.BuildTypeID, build.Number, build.WebURL)
		buildStartDate := build.StartDate.Time().Format(fmtDateTime)
		buildFinishDate := build.FinishDate.Time().Format(fmtDateTime)

		buildTable.Append([]string{
			nameLink,
			build.BuildType.ProjectName,
			buildStartDate,
			buildFinishDate,
		})
	}

	buildTable.Render()
	message += buf.String()

	return &model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_IN_CHANNEL,
		Text:         message,
	}
}

func (p *Plugin) executeCommandTriggerBuildStart(buildTypeID string) *model.CommandResponse {
	configuration := p.getConfiguration()
	client := teamcity.New(configuration.TeamCityURL,
		configuration.TeamCityToken)

	var emptyMap = make(map[string]string)

	// Check BuildTypeID is correct
	var emptyBuildType *types.BuildType
	buildType, err := client.GetBuildType(buildTypeID)

	if emptyBuildType == buildType {
		return p.postEphemeral("Invalid Build ID: `" + buildTypeID + "`")
	}

	build, err := client.QueueBuild(buildTypeID, "", emptyMap)

	if err != nil {
		return p.postEphemeral("Error starting build: `" + err.Error() + "`")
	}

	message := "**TEAMCITY BUILD STARTED**\n\n" +
		" - Build Type: [%s](%s)\n" +
		" - [Build ID: %d](%s)\n" +
		" - State: %s\n\n" +
		" Stop this build with this slash command: `/teamcity build cancel %d`"

	respText := fmt.Sprintf(message, build.BuildType.Name, build.BuildType.WebURL, build.ID, build.WebURL, build.State, build.ID)

	return &model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_IN_CHANNEL,
		Text:         respText,
	}
}

func (p *Plugin) executeCommandTriggerBuildCancel(args *model.CommandArgs) *model.CommandResponse {
	configuration := p.getConfiguration()
	client := teamcity.New(configuration.TeamCityURL,
		configuration.TeamCityToken)

	cArgs, err := p.extractCommandArgs(args.Command)

	if err != nil {
		return p.postEphemeral(fmt.Sprintf("Error parsing arguments: `%s`", err.Error()))
	}

	// Cancel command is like this:
	//  - [0] : /teamcity
	//  - [1] : build
	//  - [2] : cancel
	//  - [3] : buildID
	//  - [4] : Comments

	// Verify the buildID
	buildID, err := strconv.ParseInt(cArgs[3], 10, 64)

	if buildID == 0 {
		return p.postEphemeral(fmt.Sprintf("Invalid Build ID: %s", cArgs[3]))
	}

	if err != nil {
		return p.postEphemeral(fmt.Sprintf("Error Cancelling Build: %s", err.Error()))
	}

	buildComment := ""

	if len(cArgs) >= 5 {
		var commentSlice []string = cArgs[4:]
		buildComment = strings.Join(commentSlice, " ")
	}

	build, err := client.CancelBuild(buildID, buildComment)

	if err != nil {
		return p.postEphemeral(fmt.Sprintf("Error Cancelling Build: %s", err.Error()))
	}

	message := "**TEAMCITY BUILD CANCELLED**\n"

	message += "----\n"
	message += " - Build : [" + build.BuildTypeID + " #" + build.Number + "](" + build.WebURL + "))\n" +
		"\t - Project: " + build.BuildType.ProjectName + "\n"

	if build.Status == "SUCCESS" {
		message += "\t - Status: " + build.StatusText + "\n"
	} else {
		message += "\t - Status: **" + build.StatusText + "**\n"
	}

	buildStartDate := build.StartDate.Time().Format(fmtDateTime)
	buildFinishDate := build.FinishDate.Time().Format(fmtDateTime)

	message += "\t - Build Start: " + buildStartDate + "\n" +
		"\t - Build Finish: " + buildFinishDate + "\n"

	return &model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
		Text:         message,
	}
}

func (p *Plugin) executeCommandTriggerStats(args *model.CommandArgs) *model.CommandResponse {
	configuration := p.getConfiguration()
	client := teamcity.New(configuration.TeamCityURL,
		configuration.TeamCityToken)

	agents, err := client.GetAgentStats()

	if err != nil {
		return p.postEphemeral(fmt.Sprintf("Error getting agent stats: %s", err.Error()))
	}

	builds, err := client.GetBuildQueue()

	if err != nil {
		return p.postEphemeral(fmt.Sprintf("Error getting build queue: %s", err.Error()))
	}

	message := "**Agent Stats**\n\n"

	buf := new(bytes.Buffer)
	agentTable := tablewriter.NewWriter(buf)
	agentTable.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	agentTable.SetCenterSeparator("|")
	agentTable.SetHeader([]string{"Name", "Enabled", "Authorized", "Up to Date", "Connected", "Working"})
	agentTable.SetAutoFormatHeaders(false)
	agentTable.SetAutoWrapText(false)
	agentTable.SetColumnAlignment([]int{
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_CENTER,
		tablewriter.ALIGN_CENTER,
		tablewriter.ALIGN_CENTER,
		tablewriter.ALIGN_CENTER,
		tablewriter.ALIGN_CENTER,
	})

	for _, agent := range agents {
		working := (agent.ActiveBuild.BuildTypeID == "")
		nameLink := fmt.Sprintf("[%s](%s)", agent.Name, agent.WebURL)

		agentTable.Append([]string{nameLink,
			p.redOrGreen(agent.Enabled),
			p.redOrGreen(agent.Authorized),
			p.redOrGreen(agent.UpToDate),
			p.redOrGreen(agent.Connected),
			p.redOrGreen(working),
		})
	}

	agentTable.Render()
	message += buf.String()

	if len(builds) != 0 {
		message += fmt.Sprintf("\n---\n**Build Queue** - Total Builds: %d\n\n", len(builds))

		buf := new(bytes.Buffer)
		buildTable := tablewriter.NewWriter(buf)
		buildTable.SetAutoWrapText(false)
		buildTable.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
		buildTable.SetCenterSeparator("|")
		buildTable.SetHeader([]string{"Project", "Build Name", "Date Queued", "Queue Position"})
		buildTable.SetAutoFormatHeaders(false)
		buildTable.SetHeaderAlignment(tablewriter.ALIGN_CENTER)

		for _, build := range builds {
			nameLink := fmt.Sprintf("[%s](%s)", build.BuildType.Name, build.WebURL)
			queuedTime := build.QueuedDate.Time().Format(fmtDateTime)

			buildTable.Append([]string{
				build.BuildType.ProjectName,
				nameLink,
				queuedTime,
				fmt.Sprintf("%d", build.QueuePosition),
			})
		}

		buildTable.Render()
		message += buf.String()
	}

	return &model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_IN_CHANNEL,
		Text:         message,
	}
}

func (p *Plugin) redOrGreen(t bool) string {
	if t {
		return iconGood
	}

	return iconBad
}
