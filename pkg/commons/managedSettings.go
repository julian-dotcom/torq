package commons

var ManagedSettingsChannel = make(chan ManagedSettings)

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

func ManagedSettingsCache(ch chan ManagedSettings) {
	var defaultLanguage string
	var preferredTimeZone string
	var defaultDateRange string
	var weekStartsOn string
	for {
		managedSettings := <-ch
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
	}
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
