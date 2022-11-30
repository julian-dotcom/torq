package commons

import (
	"context"
)

var ManagedSettingsChannel = make(chan ManagedSettings) //nolint:gochecknoglobals

type ManagedSettingsCacheOperationType uint

const (
	// READ_SETTINGS please provide Out
	READ_SETTINGS ManagedSettingsCacheOperationType = iota
	// WRITE_SETTINGS please provide defaultLanguage, preferredTimeZone, defaultDateRange and weekStartsOn
	WRITE_SETTINGS
)

type ManagedSettings struct {
	Type              ManagedSettingsCacheOperationType
	DefaultLanguage   string
	PreferredTimeZone string
	DefaultDateRange  string
	WeekStartsOn      string
	Out               chan ManagedSettings
}

func ManagedSettingsCache(ch chan ManagedSettings, ctx context.Context) {
	var defaultLanguage string
	var preferredTimeZone string
	var defaultDateRange string
	var weekStartsOn string
	for {
		select {
		case <-ctx.Done():
			return
		case managedSettings := <-ch:
			defaultLanguage, preferredTimeZone, defaultDateRange, weekStartsOn =
				processManagedSettings(managedSettings, defaultLanguage, preferredTimeZone, defaultDateRange, weekStartsOn)
		}
	}
}

func processManagedSettings(managedSettings ManagedSettings,
	defaultLanguage string, preferredTimeZone string, defaultDateRange string,
	weekStartsOn string) (string, string, string, string) {

	switch managedSettings.Type {
	case READ_SETTINGS:
		managedSettings.DefaultLanguage = defaultLanguage
		managedSettings.PreferredTimeZone = preferredTimeZone
		managedSettings.DefaultDateRange = defaultDateRange
		managedSettings.WeekStartsOn = weekStartsOn
		go SendToManagedSettingsChannel(managedSettings.Out, managedSettings)
	case WRITE_SETTINGS:
		defaultLanguage = managedSettings.DefaultLanguage
		preferredTimeZone = managedSettings.PreferredTimeZone
		defaultDateRange = managedSettings.DefaultDateRange
		weekStartsOn = managedSettings.WeekStartsOn
	}
	return defaultLanguage, preferredTimeZone, defaultDateRange, weekStartsOn
}

func SendToManagedSettingsChannel(ch chan ManagedSettings, managedSettings ManagedSettings) {
	ch <- managedSettings
}

func GetSettings() ManagedSettings {
	settingsResponseChannel := make(chan ManagedSettings)
	managedSettings := ManagedSettings{
		Type: READ_SETTINGS,
		Out:  settingsResponseChannel,
	}
	ManagedSettingsChannel <- managedSettings
	return <-settingsResponseChannel
}

func SetSettings(defaultDateRange, defaultLanguage, weekStartsOn, preferredTimeZone string) {
	managedSettings := ManagedSettings{
		DefaultDateRange:  defaultDateRange,
		DefaultLanguage:   defaultLanguage,
		WeekStartsOn:      weekStartsOn,
		PreferredTimeZone: preferredTimeZone,
		Type:              WRITE_SETTINGS,
	}
	ManagedSettingsChannel <- managedSettings
}
