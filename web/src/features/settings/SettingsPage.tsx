import { AddSquare20Regular as AddIcon, Save20Regular as SaveIcon } from "@fluentui/react-icons";
import Page from "layout/Page";
import Box from "features/settings/Box";
import Button, { buttonColor, buttonPosition } from "components/buttons/Button";
import style from "features/settings/settings.module.css";
import Select, { SelectOption } from "features/forms/Select";
import React from "react";
import { defaultStaticRangesFn } from "features/timeIntervalSelect/customRanges";
import {
  useGetNodeConfigurationsQuery,
  useGetSettingsQuery,
  useGetTimeZonesQuery,
  useUpdateSettingsMutation,
} from "apiSlice";
import { nodeConfiguration, settings } from "apiTypes";
import { toastCategory } from "features/toast/Toasts";
import ToastContext from "features/toast/context";
import NodeSettings from "features/settings/NodeSettings";
import Modal from "features/modal/Modal";
import useTranslations from "services/i18n/useTranslations";
import { supportedLangs } from "config/i18nConfig";

function Settings() {
  const { t, setLang } = useTranslations();
  const { data: settingsData } = useGetSettingsQuery();
  const { data: nodeConfigurations } = useGetNodeConfigurationsQuery();
  const { data: timeZones = [] } = useGetTimeZonesQuery();
  const [updateSettings] = useUpdateSettingsMutation();
  const toastRef = React.useContext(ToastContext);
  const addNodeRef = React.useRef(null);

  const [showAddNodeState, setShowAddNodeState] = React.useState(false);
  const [settingsState, setSettingsState] = React.useState({} as settings);
  const [nodeConfigurationsState, setNodeConfigurationsState] = React.useState([] as nodeConfiguration[]);

  React.useEffect(() => {
    if (settingsData) {
      setSettingsState(settingsData);
    }
  }, [settingsData]);

  React.useEffect(() => {
    if (nodeConfigurations) {
      setNodeConfigurationsState(nodeConfigurations);
    }
  }, [nodeConfigurations]);

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
    { label: t.saturday, value: "saturday" },
    { label: t.sunday, value: "sunday" },
    { label: t.monday, value: "monday" },
  ];

  // When adding a language also add it to web/src/config/i18nConfig.js
  const languageOptions: SelectOption[] = [
    { label: supportedLangs.en, value: "en" },
    { label: supportedLangs.nl, value: "nl" },
  ];

  const handleDefaultDateRangeChange = (combiner: any) => {
    setSettingsState({ ...settingsState, defaultDateRange: combiner.value });
  };

  const handleDefaultLanguageRangeChange = (combiner: any) => {
    setSettingsState({ ...settingsState, defaultLanguage: combiner.value });
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
    setLang(settingsState?.defaultLanguage);
    toastRef?.current?.addToast(t.toast.settingsSaved, toastCategory.success);
  };

  const addNodeConfiguration = () => {
    setShowAddNodeState(true);
  };

  const handleNewNodeModalOnClose = () => {
    if (addNodeRef.current) {
      (addNodeRef.current as { clear: () => void }).clear();
    }
    setShowAddNodeState(false);
  };

  const handleOnAddSuccess = () => {
    setShowAddNodeState(false);
  };

  return (
    <Page>
      <React.Fragment>
        <div>
          <div className={style.center}>
            <div>
              <h3>Date & time settings</h3>
              <Box>
                <form onSubmit={submitPreferences}>
                  <Select
                    label={t.defaultDateRange}
                    onChange={handleDefaultDateRangeChange}
                    options={defaultDateRangeOptions}
                    value={defaultDateRangeOptions.find((dd) => dd.value === settingsState?.defaultDateRange)}
                  />
                  <Select
                    label={t.language}
                    onChange={handleDefaultLanguageRangeChange}
                    options={languageOptions}
                    value={languageOptions.find((lo) => lo.value === settingsState?.defaultLanguage)}
                  />
                  <div>
                    <Select
                      label={t.preferredTimezone}
                      onChange={handlePreferredTimezoneChange}
                      options={preferredTimezoneOptions}
                      value={preferredTimezoneOptions.find((tz) => tz.value === settingsState?.preferredTimezone)}
                    />
                  </div>
                  <Select
                    label={t.weekStartsOn}
                    onChange={handleWeekStartsOnChange}
                    options={weekStartsOnOptions}
                    value={weekStartsOnOptions.find((dd) => dd.value === settingsState?.weekStartsOn)}
                  />
                  <Button
                    type={"submit"}
                    submit={true}
                    text={t.save}
                    icon={<SaveIcon />}
                    buttonColor={buttonColor.green}
                    buttonPosition={buttonPosition.fullWidth}
                  />
                </form>
              </Box>
            </div>
            <div>
              <h3>{t.header.nodes}</h3>
              <h4>{t.header.pingSystem}</h4>
              <h5>{t.header.ambossPingSystem}</h5>
              <h5>{t.header.vectorPingSystem}</h5>
              {nodeConfigurationsState &&
                nodeConfigurationsState?.map((nodeConfiguration) => (
                  <NodeSettings
                    nodeId={nodeConfiguration.nodeId}
                    key={nodeConfiguration.nodeId ?? 0}
                    collapsed={true}
                  />
                ))}
            </div>
            <Button
              buttonColor={buttonColor.primary}
              onClick={addNodeConfiguration}
              icon={<AddIcon />}
              text={t.addNode}
            />
            <Modal title={t.addNode} show={showAddNodeState} onClose={handleNewNodeModalOnClose}>
              <NodeSettings
                ref={addNodeRef}
                addMode={true}
                nodeId={0}
                collapsed={false}
                onAddSuccess={handleOnAddSuccess}
              />
            </Modal>
          </div>
        </div>
      </React.Fragment>
    </Page>
  );
}

export default Settings;
