import { Save20Regular as SaveIcon } from "@fluentui/react-icons";
import Page from "layout/Page";
import Box from "./Box";
import style from "./settings.module.css";
import Select, { SelectOption } from "../forms/Select";
import SubmitButton from "../forms/SubmitButton";
import React, { useEffect } from "react";
import { defaultStaticRangesFn } from "../timeIntervalSelect/customRanges";
import { useGetSettingsQuery, useUpdateSettingsMutation } from "apiSlice";
import { settings } from "apiTypes";
import classNames from "classnames";

function Settings() {
  const { data: settingsData } = useGetSettingsQuery();
  const [updateSettings] = useUpdateSettingsMutation();

  const [savedMessage, setSavedMessage] = React.useState("");
  const [settingsState, setSettingsState] = React.useState({
    preferredTimezone: 0,
  } as settings);

  React.useEffect(() => {
    // do some checking here to ensure data exist
    if (settingsData) {
      // mutate data if you need to
      setSettingsState(settingsData);
    }
  }, [settingsData]);

  const defaultDateRangeLabels: {
    label: string;
    code: string;
  }[] = defaultStaticRangesFn(0);

  const defaultDateRangeOptions: SelectOption[] = defaultDateRangeLabels.map(
    (dsr) => ({ value: dsr.code, label: dsr.label })
  );

  let preferredTimezoneOptions: SelectOption[] = [];
  for (let i = -11; i <= 12; i++) {
    let label = "UTC";
    if (i < 0) {
      label += " " + i;
    }
    if (i > 0) {
      label += " +" + i;
    }
    preferredTimezoneOptions.push({ label: label, value: i.toString() });
  }

  const weekStartsOnOptions: SelectOption[] = [
    { label: "Saturday", value: "saturday" },
    { label: "Sunday", value: "sunday" },
    { label: "Monday", value: "monday" },
  ];

  const handleDefaultDateRangeChange = (combiner: any) => {
    setSettingsState({ ...settingsState, defaultDateRange: combiner.value });
  };

  const handlePreferredTimezoneChange = (combiner: any) => {
    setSettingsState({
      ...settingsState,
      preferredTimezone: parseInt(combiner.value),
    });
  };

  const handleWeekStartsOnChange = (combiner: any) => {
    setSettingsState({ ...settingsState, weekStartsOn: combiner.value });
  };

  const submitPreferences = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    updateSettings(settingsState);
    setSavedMessage("Saved!");
  };

  useEffect(() => {
    const timer = setTimeout(() => {
      setSavedMessage("");
    }, 4000);
    return () => clearTimeout(timer);
  }, [savedMessage]);

  return (
    <Page>
      <React.Fragment>
        <div>
          <div className={style.center}>
            <Box minWidth={440} title="Date & time settings">
              <form onSubmit={submitPreferences}>
                <Select
                  label="Default date range"
                  onChange={handleDefaultDateRangeChange}
                  options={defaultDateRangeOptions}
                  value={defaultDateRangeOptions.find(
                    (dd) => dd.value === settingsState?.defaultDateRange
                  )}
                />
                <Select
                  label="Preferred timezone"
                  onChange={handlePreferredTimezoneChange}
                  options={preferredTimezoneOptions}
                  value={preferredTimezoneOptions.find(
                    (tz) =>
                      tz.value === settingsState?.preferredTimezone.toString()
                  )}
                />
                <Select
                  label="Week starts on"
                  onChange={handleWeekStartsOnChange}
                  options={weekStartsOnOptions}
                  value={weekStartsOnOptions.find(
                    (dd) => dd.value === settingsState?.weekStartsOn
                  )}
                />
                <SubmitButton>
                  <React.Fragment>
                    <SaveIcon />
                    Save
                  </React.Fragment>
                </SubmitButton>
                {savedMessage && (
                  <div className={classNames(style.center, style.savedMessage)}>
                    <span>{savedMessage}</span>
                  </div>
                )}
              </form>
            </Box>
          </div>
        </div>
      </React.Fragment>
    </Page>
  );
}

export default Settings;
