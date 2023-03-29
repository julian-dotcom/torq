package commons

import (
	"context"
)

var ManagedSettingsChannel = make(chan ManagedSettings) //nolint:gochecknoglobals

type ManagedSettingsCacheOperationType uint

const (
	// readSettings please provide Out
	readSettings ManagedSettingsCacheOperationType = iota
	// writeSettings please provide defaultLanguage, preferredTimeZone, defaultDateRange and weekStartsOn
	writeSettings
	writeBlockHeight
	writeVectorUrl
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
	VectorUrl         string
	Out               chan<- ManagedSettings
}

type ManagedSettingsData struct {
	DefaultLanguage   string
	PreferredTimeZone string
	DefaultDateRange  string
	WeekStartsOn      string
	TorqUuid          string
	MixpanelOptOut    bool
	BlockHeight       uint32
	VectorUrl         string
}

func ManagedSettingsCache(ch <-chan ManagedSettings, ctx context.Context) {
	var data ManagedSettingsData
	for {
		select {
		case <-ctx.Done():
			return
		case managedSettings := <-ch:
			data = processManagedSettings(managedSettings, data)
		}
	}
}

func processManagedSettings(managedSettings ManagedSettings, data ManagedSettingsData) ManagedSettingsData {
	switch managedSettings.Type {
	case readSettings:
		managedSettings.DefaultLanguage = data.DefaultLanguage
		if data.DefaultLanguage == "" {
			managedSettings.DefaultLanguage = "en"
		}
		managedSettings.PreferredTimeZone = data.PreferredTimeZone
		if data.PreferredTimeZone == "" {
			managedSettings.PreferredTimeZone = "UTC"
		}
		managedSettings.DefaultDateRange = data.DefaultDateRange
		if data.DefaultDateRange == "" {
			managedSettings.DefaultDateRange = "last7days"
		}
		managedSettings.WeekStartsOn = data.WeekStartsOn
		if data.WeekStartsOn == "" {
			managedSettings.WeekStartsOn = "monday"
		}
		managedSettings.TorqUuid = data.TorqUuid
		managedSettings.MixpanelOptOut = data.MixpanelOptOut
		managedSettings.BlockHeight = data.BlockHeight
		managedSettings.VectorUrl = data.VectorUrl
		SendToManagedSettingsChannel(managedSettings.Out, managedSettings)
	case writeSettings:
		data.DefaultLanguage = managedSettings.DefaultLanguage
		data.PreferredTimeZone = managedSettings.PreferredTimeZone
		data.DefaultDateRange = managedSettings.DefaultDateRange
		data.WeekStartsOn = managedSettings.WeekStartsOn
		data.TorqUuid = managedSettings.TorqUuid
		data.MixpanelOptOut = managedSettings.MixpanelOptOut
	case writeVectorUrl:
		data.VectorUrl = managedSettings.VectorUrl
	case writeBlockHeight:
		data.BlockHeight = managedSettings.BlockHeight
	}
	return data
}

func SendToManagedSettingsChannel(ch chan<- ManagedSettings, managedSettings ManagedSettings) {
	ch <- managedSettings
}

func GetSettings() ManagedSettings {
	settingsResponseChannel := make(chan ManagedSettings)
	managedSettings := ManagedSettings{
		Type: readSettings,
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
		Type:              writeSettings,
	}
	ManagedSettingsChannel <- managedSettings
}

func GetBlockHeight() uint32 {
	settingsResponseChannel := make(chan ManagedSettings)
	managedSettings := ManagedSettings{
		Type: readSettings,
		Out:  settingsResponseChannel,
	}
	ManagedSettingsChannel <- managedSettings
	settings := <-settingsResponseChannel
	return settings.BlockHeight
}

func SetBlockHeight(blockHeight uint32) {
	managedSettings := ManagedSettings{
		BlockHeight: blockHeight,
		Type:        writeBlockHeight,
	}
	ManagedSettingsChannel <- managedSettings
}

func SetVectorUrlBase(vectorUrlBase string) {
	managedSettings := ManagedSettings{
		VectorUrl: vectorUrlBase,
		Type:      writeVectorUrl,
	}
	ManagedSettingsChannel <- managedSettings
}

func GetVectorUrlBase() string {
	settingsResponseChannel := make(chan ManagedSettings)
	managedSettings := ManagedSettings{
		Type: readSettings,
		Out:  settingsResponseChannel,
	}
	ManagedSettingsChannel <- managedSettings
	settings := <-settingsResponseChannel
	return settings.VectorUrl
}
