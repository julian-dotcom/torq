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
	WRITE_BLOCKHEIGHT
)

type ManagedSettings struct {
	Type              ManagedSettingsCacheOperationType
	DefaultLanguage   string
	PreferredTimeZone string
	DefaultDateRange  string
	WeekStartsOn      string
	TorqUuid          string
	MixpanelOptOut    bool
	BlockHeight       uint32
	Out               chan<- ManagedSettings
}

func ManagedSettingsCache(ch <-chan ManagedSettings, ctx context.Context) {
	var defaultLanguage string
	var preferredTimeZone string
	var defaultDateRange string
	var weekStartsOn string
	var torqUuid string
	var mixpanelOptOut bool
	var blockHeight uint32
	for {
		select {
		case <-ctx.Done():
			return
		case managedSettings := <-ch:
			defaultLanguage, preferredTimeZone, defaultDateRange, weekStartsOn, torqUuid, mixpanelOptOut, blockHeight =
				processManagedSettings(managedSettings, defaultLanguage, preferredTimeZone, defaultDateRange,
					weekStartsOn, torqUuid, mixpanelOptOut, blockHeight)
		}
	}
}

func processManagedSettings(managedSettings ManagedSettings,
	defaultLanguage string, preferredTimeZone string, defaultDateRange string,
	weekStartsOn string, torqUuid string, mixpanelOptOut bool,
	blockHeight uint32) (string, string, string, string, string, bool, uint32) {

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
		managedSettings.BlockHeight = blockHeight
		SendToManagedSettingsChannel(managedSettings.Out, managedSettings)
	case WRITE_SETTINGS:
		defaultLanguage = managedSettings.DefaultLanguage
		preferredTimeZone = managedSettings.PreferredTimeZone
		defaultDateRange = managedSettings.DefaultDateRange
		weekStartsOn = managedSettings.WeekStartsOn
		torqUuid = managedSettings.TorqUuid
		mixpanelOptOut = managedSettings.MixpanelOptOut
	case WRITE_BLOCKHEIGHT:
		blockHeight = managedSettings.BlockHeight
	}
	return defaultLanguage, preferredTimeZone, defaultDateRange, weekStartsOn, torqUuid, mixpanelOptOut, blockHeight
}

func SendToManagedSettingsChannel(ch chan<- ManagedSettings, managedSettings ManagedSettings) {
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

func GetBlockHeight() uint32 {
	settingsResponseChannel := make(chan ManagedSettings)
	managedSettings := ManagedSettings{
		Type: READ_SETTINGS,
		Out:  settingsResponseChannel,
	}
	ManagedSettingsChannel <- managedSettings
	settings := <-settingsResponseChannel
	return settings.BlockHeight
}

func SetBlockHeight(blockHeight uint32) {
	managedSettings := ManagedSettings{
		BlockHeight: blockHeight,
		Type:        WRITE_BLOCKHEIGHT,
	}
	ManagedSettingsChannel <- managedSettings
}
