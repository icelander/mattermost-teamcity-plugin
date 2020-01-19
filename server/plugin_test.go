package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost-server/v5/model"
)

type commandArgs struct {
	UserID string `json:"user_id"`
	ChannelID string `json:"channel_id"`
	TeamID string `json:"team_id"`
	RootID string `json:"root_id"`
	ParentID string `json:"parent_id"`
	TriggerID string `json:"trigger_id"`
	Command string `json:"command"`
}

func generateArgs(cmd string) *model.CommandArgs {
	var cArgs = &model.CommandArgs{
		UserId: "31rs9bjkm38rxq6666x1tgm9to",
		ChannelId: "8orrfwp793yypgucsbysuusu4c",
		TeamId: "qoocd8165ibjdynn96qo1p6d8w",
		RootId: "",
		ParentId: "",
		TriggerId: "OW1oY3V0d2Rwam42ZnAxbWV1ZWc0M25lYnI6MzFyczliamttMzhyeHE2NjY2eDF0Z205dG86MTU3OTQwNjE5NDg1NjpNRVFDSUNEaEZ2RGcxNFUwdFRiZDFwY1lmZkJFNW9QOXA4c3UrblB0Y2E2d3BYRUlBaUJmRmo5Q1hTcEZwZTVpOVlOcjF0WmhwQjRUK3R5bm1KWXh3bkZzU0dpZ2d3PT0=",
		Command: "/teamcity " + cmd,
	}

	return cArgs
}

func TestPluginNotInstalled(t *testing.T) {
	assert := assert.New(t)
	plugin := Plugin{}

	
	cArgs := generateArgs("list projects")
	response := plugin.executeCommandHooks(cArgs)

	assert.Equal(errorNotInstalled, response.Text)
}

func TestInstallPlugin(t *testing.T) {
	assert := assert.New(t)
	plugin := Plugin{}

	
	cArgs := generateArgs("install http://127.0.0.1:8111/ paul mac4life")
	response := plugin.executeCommandHooks(cArgs)
	
	configuration := plugin.getConfiguration()

	assert.Equal("http://127.0.0.1:8111/", configuration.teamCityURL)
	assert.Equal("paul", configuration.teamCityUsername)
	assert.Equal("mac4life", configuration.teamCityPassword)
	assert.True(configuration.installed)

	assert.Contains(response.Text, "TeamCity Installed!")
}

func TestEnablePlugin(t *testing.T) {
	assert := assert.New(t)
	plugin := Plugin{}

	// Install it first
	plugin.executeCommandHooks(generateArgs("install http://127.0.0.1:8111/ paul mac4life"))

	
	cArgs := generateArgs("enable")
	response := plugin.executeCommandHooks(cArgs)
	
	configuration := plugin.getConfiguration()

	assert.False(configuration.disabled, "configuration.disabled should be false after enabling it")
	assert.Equal(msgEnabled, response.Text)
}

func TestDisablePlugin(t *testing.T) {
	assert := assert.New(t)
	plugin := Plugin{}

	// Install it first
	plugin.executeCommandHooks(generateArgs("install http://127.0.0.1:8111/ paul mac4life"))
	
	cArgs := generateArgs("disable")
	response := plugin.executeCommandHooks(cArgs)
	
	configuration := plugin.getConfiguration()

	assert.True(configuration.disabled, "configuration.disabled should be true after disabling it")
	assert.Equal(msgDisabled, response.Text)
}

func TestPluginDisabled(t *testing.T) {
	assert := assert.New(t)
	plugin := Plugin{}
	
	// Install it first
	plugin.executeCommandHooks(generateArgs("install http://127.0.0.1:8111/ paul mac4life"))


	// Make sure the plugin is disabled
	plugin.executeCommandHooks(generateArgs("disable"))
	configuration := plugin.getConfiguration()

	assert.True(configuration.disabled, "configuration.disabled should be true")
	assert.True(configuration.installed, "configuration.installed should be true")	
	
	cArgs := generateArgs("list projects")
	response := plugin.executeCommandHooks(cArgs)

	assert.Equal(errorDisabled, response.Text)
}

func TestListProjects(t *testing.T) {
	assert := assert.New(t)
	plugin := Plugin{}
	
	// Install it first
	plugin.executeCommandHooks(generateArgs("install http://127.0.0.1:8111/ paul mac4life"))

	cArgs := generateArgs("list projects")
	response := plugin.executeCommandHooks(cArgs)

	assert.Contains(response.Text, "TeamCity Projects")
}

func TestListBuilds(t *testing.T) {
	assert := assert.New(t)
	plugin := Plugin{}
	
	// Install it first
	plugin.executeCommandHooks(generateArgs("install http://127.0.0.1:8111/ paul mac4life"))

	cArgs := generateArgs("list builds")
	response := plugin.executeCommandHooks(cArgs)

	assert.Contains(response.Text, "TeamCity Builds")
}