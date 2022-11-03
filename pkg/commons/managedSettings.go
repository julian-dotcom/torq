package commons

import (
	"context"
)

var ManagedSettingsChannel = make(chan ManagedSettings) //nolint:gochecknoglobals

type ManagedSettingsCacheOperationType uint

const (
	// READ please provide Out
	READ ManagedSettingsCacheOperationType = iota
	// WRITE please provide defaultLanguage, preferredTimeZone, defaultDateRange and weekStartsOn
	WRITE
)

type ManagedSettings struct {
	Type              ManagedSettingsCacheOperationType
	DefaultLanguage   string
	PreferredTimeZone string
	DefaultDateRange  string
	WeekStartsOn      string
	Out               chan ManagedSettings
}

// ManagedSettingsCache parameter Context is for test cases...
func ManagedSettingsCache(ch chan ManagedSettings, ctx context.Context) {
	var defaultLanguage string
	var preferredTimeZone string
	var defaultDateRange string
	var weekStartsOn string
	for {
		if ctx == nil {
			managedSettings := <-ch
			defaultLanguage, preferredTimeZone, defaultDateRange, weekStartsOn =
				processManagedSettings(managedSettings, defaultLanguage, preferredTimeZone, defaultDateRange, weekStartsOn)
		} else {
			// TODO: The code itself is fine here but special case only for test cases?
			// Running Torq we don't have nor need to be able to cancel but we do for test cases because global var is shared
			select {
			case <-ctx.Done():
				return
			case managedSettings := <-ch:
				defaultLanguage, preferredTimeZone, defaultDateRange, weekStartsOn =
					processManagedSettings(managedSettings, defaultLanguage, preferredTimeZone, defaultDateRange, weekStartsOn)
			}
		}
	}
}

func processManagedSettings(managedSettings ManagedSettings, defaultLanguage string, preferredTimeZone string, defaultDateRange string, weekStartsOn string) (string, string, string, string) {
	switch managedSettings.Type {
	case READ:
		managedSettings.DefaultLanguage = defaultLanguage
		managedSettings.PreferredTimeZone = preferredTimeZone
		managedSettings.DefaultDateRange = defaultDateRange
		managedSettings.WeekStartsOn = weekStartsOn
		go SendToManagedSettingsChannel(managedSettings.Out, managedSettings)
	case WRITE:
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
		Type: READ,
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
		Type:              WRITE,
	}
	ManagedSettingsChannel <- managedSettings
}
