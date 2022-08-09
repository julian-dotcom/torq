import { Save20Regular as SaveIcon, AddSquare20Regular as AddIcon } from "@fluentui/react-icons";
import Page from "layout/Page";
import Box from "./Box";
import Button, { buttonVariants } from "features/buttons/Button";
import style from "./settings.module.css";
import Select, { SelectOption } from "../forms/Select";
import SubmitButton from "../forms/SubmitButton";
import React from "react";
import { defaultStaticRangesFn } from "../timeIntervalSelect/customRanges";
import { useGetSettingsQuery, useUpdateSettingsMutation, useGetTimeZonesQuery, useGetLocalNodesQuery } from "apiSlice";
import { settings } from "apiTypes";
import { toastCategory } from "../toast/Toasts";
import ToastContext from "../toast/context";
import NodeSettings from "./NodeSettings";

function Settings() {
  const { data: settingsData } = useGetSettingsQuery();
  const { data: localNodes } = useGetLocalNodesQuery();
  const { data: timeZones = [] } = useGetTimeZonesQuery();
  const [updateSettings] = useUpdateSettingsMutation();
  const toastRef = React.useContext(ToastContext);

  const [settingsState, setSettingsState] = React.useState({} as settings);

  React.useEffect(() => {
    if (settingsData) {
      setSettingsState(settingsData);
    }
  }, [settingsData]);

  const defaultDateRangeLabels: {
    label: string;
    code: string;
  }[] = defaultStaticRangesFn(0);

  const defaultDateRangeOptions: SelectOption[] = defaultDateRangeLabels.map((dsr) => ({
    value: dsr.code,
    label: dsr.label,
  }));

  const preferredTimezoneOptions: SelectOption[] = timeZones.map((tz) => ({
    value: tz.name,
    label: tz.name,
  }));

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
      preferredTimezone: combiner.value,
    });
  };

  const handleWeekStartsOnChange = (combiner: any) => {
    setSettingsState({ ...settingsState, weekStartsOn: combiner.value });
  };

  const submitPreferences = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    updateSettings(settingsState);
    toastRef?.current?.addToast("Settings saved", toastCategory.success);
  };

  const addLocalNode = () => {
    console.log("Adding additional node");
  };

  return (
    <Page>
      <React.Fragment>
        <div>
          <div className={style.center}>
            <Box title="Date & time settings">
              <form onSubmit={submitPreferences}>
                <Select
                  label="Default date range"
                  onChange={handleDefaultDateRangeChange}
                  options={defaultDateRangeOptions}
                  value={defaultDateRangeOptions.find((dd) => dd.value === settingsState?.defaultDateRange)}
                />
                <div>
                  <Select
                    label="Preferred timezone"
                    onChange={handlePreferredTimezoneChange}
                    options={preferredTimezoneOptions}
                    value={preferredTimezoneOptions.find((tz) => tz.value === settingsState?.preferredTimezone)}
                  />
                </div>
                <Select
                  label="Week starts on"
                  onChange={handleWeekStartsOnChange}
                  options={weekStartsOnOptions}
                  value={weekStartsOnOptions.find((dd) => dd.value === settingsState?.weekStartsOn)}
                />
                <SubmitButton>
                  <React.Fragment>
                    <SaveIcon />
                    Save
                  </React.Fragment>
                </SubmitButton>
              </form>
            </Box>
            {localNodes?.map((localNode) => (
              <NodeSettings localNodeId={localNode.localNodeId} />
            ))}
            <NodeSettings localNodeId={1} />
            <Button variant={buttonVariants.primary} onClick={addLocalNode} icon={<AddIcon />} text="Add Node" />
          </div>
        </div>
      </React.Fragment>
    </Page>
  );
}

export default Settings;
