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
	TorqUuid          string
	MixpanelOptOut    bool
	Out               chan ManagedSettings
}

func ManagedSettingsCache(ch chan ManagedSettings, ctx context.Context) {
	var defaultLanguage string
	var preferredTimeZone string
	var defaultDateRange string
	var weekStartsOn string
	var torqUuid string
	var mixpanelOptOut bool
	for {
		select {
		case <-ctx.Done():
			return
		case managedSettings := <-ch:
			defaultLanguage, preferredTimeZone, defaultDateRange, weekStartsOn, torqUuid, mixpanelOptOut =
				processManagedSettings(managedSettings, defaultLanguage, preferredTimeZone, defaultDateRange,
					weekStartsOn, torqUuid, mixpanelOptOut)
		}
	}
}

func processManagedSettings(managedSettings ManagedSettings,
	defaultLanguage string, preferredTimeZone string, defaultDateRange string,
	weekStartsOn string, torqUuid string, mixpanelOptOut bool) (string, string, string, string, string, bool) {

	switch managedSettings.Type {
	case READ_SETTINGS:
		managedSettings.DefaultLanguage = defaultLanguage
		if defaultLanguage == "" {
			managedSettings.DefaultLanguage = "en"
		}
		managedSettings.PreferredTimeZone = preferredTimeZone
		if preferredTimeZone == "" {
			managedSettings.PreferredTimeZone = "UTC"
		}
		managedSettings.DefaultDateRange = defaultDateRange
		if defaultDateRange == "" {
			managedSettings.DefaultDateRange = "last7days"
		}
		managedSettings.WeekStartsOn = weekStartsOn
		if weekStartsOn == "" {
			managedSettings.WeekStartsOn = "monday"
		}
		managedSettings.TorqUuid = torqUuid
		managedSettings.MixpanelOptOut = mixpanelOptOut
		SendToManagedSettingsChannel(managedSettings.Out, managedSettings)
	case WRITE_SETTINGS:
		defaultLanguage = managedSettings.DefaultLanguage
		preferredTimeZone = managedSettings.PreferredTimeZone
		defaultDateRange = managedSettings.DefaultDateRange
		weekStartsOn = managedSettings.WeekStartsOn
		torqUuid = managedSettings.TorqUuid
		mixpanelOptOut = managedSettings.MixpanelOptOut
	}
	return defaultLanguage, preferredTimeZone, defaultDateRange, weekStartsOn, torqUuid, mixpanelOptOut
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

func SetSettings(defaultDateRange string, defaultLanguage string, weekStartsOn string, preferredTimeZone string,
	torqUuid string, mixpanelOptOut bool) {
	managedSettings := ManagedSettings{
		DefaultDateRange:  defaultDateRange,
		DefaultLanguage:   defaultLanguage,
		WeekStartsOn:      weekStartsOn,
		PreferredTimeZone: preferredTimeZone,
		TorqUuid:          torqUuid,
		MixpanelOptOut:    mixpanelOptOut,
		Type:              WRITE_SETTINGS,
	}
	ManagedSettingsChannel <- managedSettings
}
