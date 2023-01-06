import { AddSquare20Regular as AddIcon, Save20Regular as SaveIcon } from "@fluentui/react-icons";
import Page from "layout/Page";
import Button, { ColorVariant, buttonPosition } from "components/buttons/Button";
import styles from "features/settings/settings.module.css";
import { SelectOption } from "features/forms/Select";
import Select from "components/forms/select/Select";
import React from "react";
import { defaultStaticRangesFn } from "features/timeIntervalSelect/customRanges";
import {
  useGetNodeConfigurationsQuery,
  useGetSettingsQuery,
  useGetTimeZonesQuery,
  useUpdateSettingsMutation,
} from "apiSlice";
import { settings } from "apiTypes";
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
    { label: t.saturday, value: "saturday" },
    { label: t.sunday, value: "sunday" },
    { label: t.monday, value: "monday" },
  ];

  // When adding a language also add it to web/src/config/i18nConfig.js
  const languageOptions: SelectOption[] = [
    { label: supportedLangs.en, value: "en" },
    { label: supportedLangs.nl, value: "nl" },
  ];

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const handleDefaultDateRangeChange = (combiner: any) => {
    setSettingsState({ ...settingsState, defaultDateRange: combiner.value });
  };

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const handleDefaultLanguageRangeChange = (combiner: any) => {
    setSettingsState({ ...settingsState, defaultLanguage: combiner.value });
  };

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const handlePreferredTimezoneChange = (combiner: any) => {
    setSettingsState({
      ...settingsState,
      preferredTimezone: combiner.value,
    });
  };

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
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
          <div className={styles.center}>
            <div>
              <h3>Date & time settings</h3>

              <form onSubmit={submitPreferences} className={styles.settingsForm}>
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
                  icon={<SaveIcon />}
                  buttonColor={ColorVariant.success}
                  buttonPosition={buttonPosition.fullWidth}
                >
                  {t.save}
                </Button>
              </form>
            </div>
            <div>
              <h3>{t.header.nodes}</h3>
              {nodeConfigurations &&
                nodeConfigurations?.map((nodeConfiguration) => (
                  <NodeSettings
                    nodeId={nodeConfiguration.nodeId}
                    key={nodeConfiguration.nodeId ?? 0}
                    collapsed={true}
                  />
                ))}
            </div>
            <Button buttonColor={ColorVariant.success} onClick={addNodeConfiguration} icon={<AddIcon />}>
              {t.addNode}
            </Button>
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
