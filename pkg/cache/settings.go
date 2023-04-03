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

<<<<<<< serviceRefactoring:pkg/cache/settings.go
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
=======
type ManagedSettings struct {
	Type                            ManagedSettingsCacheOperationType
	DefaultLanguage                 string
	PreferredTimeZone               string
	DefaultDateRange                string
	WeekStartsOn                    string
	TorqUuid                        string
	MixpanelOptOut                  bool
	SlackOAuthToken                 *string
	SlackBotAppToken                *string
	TelegramHighPriorityCredentials *string
	TelegramLowPriorityCredentials  *string
	BlockHeight                     uint32
	VectorUrl                       string
	Out                             chan<- ManagedSettings
}

func (s ManagedSettings) GetTelegramCredential(highPriority bool) string {
	if highPriority {
		if s.TelegramHighPriorityCredentials == nil {
			return ""
		}
		return *s.TelegramHighPriorityCredentials
	}
	if s.TelegramLowPriorityCredentials == nil {
		return ""
	}
	return *s.TelegramLowPriorityCredentials
}

func (s ManagedSettings) GetSlackCredential() (string, string) {
	oauth := ""
	if s.SlackOAuthToken != nil {
		oauth = *s.SlackOAuthToken
	}
	appToken := ""
	if s.SlackBotAppToken != nil {
		oauth = *s.SlackBotAppToken
	}
	return oauth, appToken
}

type ManagedSettingsData struct {
	DefaultLanguage                 string
	PreferredTimeZone               string
	DefaultDateRange                string
	WeekStartsOn                    string
	TorqUuid                        string
	MixpanelOptOut                  bool
	SlackOAuthToken                 *string
	SlackBotAppToken                *string
	TelegramHighPriorityCredentials *string
	TelegramLowPriorityCredentials  *string
	BlockHeight                     uint32
	VectorUrl                       string
>>>>>>> Add Slack and Telegram support.:pkg/commons/managedSettings.go
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
<<<<<<< serviceRefactoring:pkg/cache/settings.go
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
=======
		managedSettings.TorqUuid = data.TorqUuid
		managedSettings.MixpanelOptOut = data.MixpanelOptOut
		managedSettings.SlackOAuthToken = data.SlackOAuthToken
		managedSettings.SlackBotAppToken = data.SlackBotAppToken
		managedSettings.TelegramHighPriorityCredentials = data.TelegramHighPriorityCredentials
		managedSettings.TelegramLowPriorityCredentials = data.TelegramLowPriorityCredentials
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
		data.SlackOAuthToken = managedSettings.SlackOAuthToken
		data.SlackBotAppToken = managedSettings.SlackBotAppToken
		data.TelegramHighPriorityCredentials = managedSettings.TelegramHighPriorityCredentials
		data.TelegramLowPriorityCredentials = managedSettings.TelegramLowPriorityCredentials
>>>>>>> Add Slack and Telegram support.:pkg/commons/managedSettings.go
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
	torqUuid string, mixpanelOptOut bool,
	slackOAuthToken *string, slackBotAppToken *string,
	telegramHighPriorityCredentials *string, telegramLowPriorityCredentials *string) {

<<<<<<< serviceRefactoring:pkg/cache/settings.go
	settingsCache := SettingsCache{
		DefaultDateRange:  defaultDateRange,
		DefaultLanguage:   defaultLanguage,
		WeekStartsOn:      weekStartsOn,
		PreferredTimeZone: preferredTimeZone,
		TorqUuid:          torqUuid,
		MixpanelOptOut:    mixpanelOptOut,
		Type:              writeSettings,
=======
	managedSettings := ManagedSettings{
		DefaultDateRange:                defaultDateRange,
		DefaultLanguage:                 defaultLanguage,
		WeekStartsOn:                    weekStartsOn,
		PreferredTimeZone:               preferredTimeZone,
		TorqUuid:                        torqUuid,
		MixpanelOptOut:                  mixpanelOptOut,
		SlackOAuthToken:                 slackOAuthToken,
		SlackBotAppToken:                slackBotAppToken,
		TelegramHighPriorityCredentials: telegramHighPriorityCredentials,
		TelegramLowPriorityCredentials:  telegramLowPriorityCredentials,
		Type:                            writeSettings,
>>>>>>> Add Slack and Telegram support.:pkg/commons/managedSettings.go
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
