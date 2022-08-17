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
import { localNode } from "apiTypes";

function Settings() {
  const { data: settingsData } = useGetSettingsQuery();
  const { data: localNodes } = useGetLocalNodesQuery();
  const { data: timeZones = [] } = useGetTimeZonesQuery();
  const [updateSettings] = useUpdateSettingsMutation();
  const toastRef = React.useContext(ToastContext);

  const [settingsState, setSettingsState] = React.useState({} as settings);
  const [localNodesState, setLocalNodesState] = React.useState([] as localNode[]);

  React.useEffect(() => {
    if (settingsData) {
      setSettingsState(settingsData);
    }
  }, [settingsData]);

  React.useEffect(() => {
    if (localNodes) {
      setLocalNodesState(localNodes);
    }
  }, [localNodes]);

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
    if (!localNodesState.some((item) => item.localNodeId === undefined)) {
      setLocalNodesState([...localNodesState, {} as localNode]);
    }
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
            <span>Node Settings</span>
            {localNodesState &&
              localNodesState?.map((localNode) => (
                <NodeSettings
                  localNodeId={localNode.localNodeId}
                  key={localNode.localNodeId ?? 0}
                  collapsed={!!localNode.localNodeId}
                />
              ))}
            <Button variant={buttonVariants.primary} onClick={addLocalNode} icon={<AddIcon />} text="Add Node" />
          </div>
        </div>
      </React.Fragment>
    </Page>
  );
}

export default Settings;
