package main

import (
	"reflect"

	"github.com/pkg/errors"
	"net/url"

	"github.com/icelander/teamcity-sdk-go/teamcity"
)

const defaultListBuildsMax = 5

type configuration struct {
	disabled          bool
	TeamCityURL       string
	TeamCityToken     string
	TeamCityMaxBuilds int
}

func (c *configuration) GetMaxBuilds() int {
	if c.TeamCityMaxBuilds == 0 {
		c.TeamCityMaxBuilds = defaultListBuildsMax
	}

	return c.TeamCityMaxBuilds
}

// Installed returns true if the plugin is configured and can connect to the server
func (c *configuration) Installed() bool {
	if c.TeamCityToken == "" || c.TeamCityURL == "" {
		return false
	}

	_, err := url.ParseRequestURI(c.TeamCityURL)

	if err != nil {
		return false
	}

	client := teamcity.New(c.TeamCityURL, c.TeamCityToken)
	_, err = client.Server()

	if err != nil {
		return false
	}

	return true
}

// Clone shallow copies the configuration. Your implementation may require a deep copy if
// your configuration has reference types.
func (c *configuration) Clone() *configuration {
	var clone = *c
	return &clone
}

// getConfiguration retrieves the active configuration under lock, making it safe to use
// concurrently. The active configuration may change underneath the client of this method, but
// the struct returned by this API call is considered immutable.
func (p *Plugin) getConfiguration() *configuration {
	p.configurationLock.RLock()
	defer p.configurationLock.RUnlock()

	if p.configuration == nil {
		return &configuration{}
	}

	return p.configuration
}

// setConfiguration replaces the active configuration under lock.
//
// Do not call setConfiguration while holding the configurationLock, as sync.Mutex is not
// reentrant. In particular, avoid using the plugin API entirely, as this may in turn trigger a
// hook back into the plugin. If that hook attempts to acquire this lock, a deadlock may occur.
//
// This method panics if setConfiguration is called with the existing configuration. This almost
// certainly means that the configuration was modified without being cloned and may result in
// an unsafe access.
func (p *Plugin) setConfiguration(configuration *configuration) {
	p.configurationLock.Lock()
	defer p.configurationLock.Unlock()

	if configuration != nil && p.configuration == configuration {
		// Ignore assignment if the configuration struct is empty. Go will optimize the
		// allocation for same to point at the same memory address, breaking the check
		// above.
		if reflect.ValueOf(*configuration).NumField() == 0 {
			return
		}
	}

	p.configuration = configuration
}

// OnConfigurationChange is invoked when configuration changes may have been made.
func (p *Plugin) OnConfigurationChange() error {
	var configuration = new(configuration)

	// Load the public configuration fields from the Mattermost server configuration.
	if err := p.API.LoadPluginConfiguration(configuration); err != nil {
		return errors.Wrap(err, "failed to load plugin configuration")
	}

	p.setConfiguration(configuration)

	return nil
}
