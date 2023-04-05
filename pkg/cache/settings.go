package cache

import (
	"context"
)

var SettingsCacheChannel = make(chan SettingsCache) //nolint:gochecknoglobals

type SettingsCacheOperationType uint

const (
	// readSettings please provide Out
	readSettings SettingsCacheOperationType = iota
	// writeSettings please provide defaultLanguage, preferredTimeZone, defaultDateRange and weekStartsOn
	writeSettings
	writeBlockHeight
	writeVectorUrl
)

type SettingsCache struct {
	Type              SettingsCacheOperationType
	DefaultLanguage   string
	PreferredTimeZone string
	DefaultDateRange  string
	WeekStartsOn      string
	TorqUuid          string
	MixpanelOptOut    bool
	BlockHeight       uint32
	VectorUrl         string
	Out               chan<- SettingsCache
}

type SettingsDataCache struct {
	DefaultLanguage   string
	PreferredTimeZone string
	DefaultDateRange  string
	WeekStartsOn      string
	TorqUuid          string
	MixpanelOptOut    bool
	BlockHeight       uint32
	VectorUrl         string
}

func SettingsCacheHandle(ch <-chan SettingsCache, ctx context.Context) {
	var data SettingsDataCache
	for {
		select {
		case <-ctx.Done():
			return
		case settingsCache := <-ch:
			data = handleSettingsOperation(settingsCache, data)
		}
	}
}

func handleSettingsOperation(settingsCache SettingsCache, data SettingsDataCache) SettingsDataCache {
	switch settingsCache.Type {
	case readSettings:
		settingsCache.DefaultLanguage = data.DefaultLanguage
		if data.DefaultLanguage == "" {
			settingsCache.DefaultLanguage = "en"
		}
		settingsCache.PreferredTimeZone = data.PreferredTimeZone
		if data.PreferredTimeZone == "" {
			settingsCache.PreferredTimeZone = "UTC"
		}
		settingsCache.DefaultDateRange = data.DefaultDateRange
		if data.DefaultDateRange == "" {
			settingsCache.DefaultDateRange = "last7days"
		}
		settingsCache.WeekStartsOn = data.WeekStartsOn
		if data.WeekStartsOn == "" {
			settingsCache.WeekStartsOn = "monday"
		}
		settingsCache.TorqUuid = data.TorqUuid
		settingsCache.MixpanelOptOut = data.MixpanelOptOut
		settingsCache.BlockHeight = data.BlockHeight
		settingsCache.VectorUrl = data.VectorUrl
		settingsCache.Out <- settingsCache
	case writeSettings:
		data.DefaultLanguage = settingsCache.DefaultLanguage
		data.PreferredTimeZone = settingsCache.PreferredTimeZone
		data.DefaultDateRange = settingsCache.DefaultDateRange
		data.WeekStartsOn = settingsCache.WeekStartsOn
		data.TorqUuid = settingsCache.TorqUuid
		data.MixpanelOptOut = settingsCache.MixpanelOptOut
	case writeVectorUrl:
		data.VectorUrl = settingsCache.VectorUrl
	case writeBlockHeight:
		data.BlockHeight = settingsCache.BlockHeight
	}
	return data
}

func GetSettings() SettingsCache {
	settingsResponseChannel := make(chan SettingsCache)
	settingsCache := SettingsCache{
		Type: readSettings,
		Out:  settingsResponseChannel,
	}
	SettingsCacheChannel <- settingsCache
	return <-settingsResponseChannel
}

func SetSettings(defaultDateRange string, defaultLanguage string, weekStartsOn string, preferredTimeZone string,
	torqUuid string, mixpanelOptOut bool) {

	settingsCache := SettingsCache{
		DefaultDateRange:  defaultDateRange,
		DefaultLanguage:   defaultLanguage,
		WeekStartsOn:      weekStartsOn,
		PreferredTimeZone: preferredTimeZone,
		TorqUuid:          torqUuid,
		MixpanelOptOut:    mixpanelOptOut,
		Type:              writeSettings,
	}
	SettingsCacheChannel <- settingsCache
}

func GetBlockHeight() uint32 {
	settingsResponseChannel := make(chan SettingsCache)
	settingsCache := SettingsCache{
		Type: readSettings,
		Out:  settingsResponseChannel,
	}
	SettingsCacheChannel <- settingsCache
	settings := <-settingsResponseChannel
	return settings.BlockHeight
}

func SetBlockHeight(blockHeight uint32) {
	settingsCache := SettingsCache{
		BlockHeight: blockHeight,
		Type:        writeBlockHeight,
	}
	SettingsCacheChannel <- settingsCache
}

func SetVectorUrlBase(vectorUrlBase string) {
	settingsCache := SettingsCache{
		VectorUrl: vectorUrlBase,
		Type:      writeVectorUrl,
	}
	SettingsCacheChannel <- settingsCache
}

func GetVectorUrlBase() string {
	settingsResponseChannel := make(chan SettingsCache)
	settingsCache := SettingsCache{
		Type: readSettings,
		Out:  settingsResponseChannel,
	}
	SettingsCacheChannel <- settingsCache
	settings := <-settingsResponseChannel
	return settings.VectorUrl
}
