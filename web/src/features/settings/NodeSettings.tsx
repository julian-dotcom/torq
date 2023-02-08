import React, { useState } from "react";
import {
  ChevronDown20Regular as CollapsedIcon,
  Delete20Regular as DeleteIcon,
  Delete24Regular as DeleteIconHeader,
  LineHorizontal120Regular as ExpandedIcon,
  MoreCircle20Regular as MoreIcon,
  Pause16Regular as DisconnectedIcon,
  Pause20Regular as PauseIcon,
  Play16Regular as ConnectedIcon,
  Play20Regular as PlayIcon,
  Save20Regular as SaveIcon,
} from "@fluentui/react-icons";
import mixpanel from "mixpanel-browser";
import styles from "./NodeSettings.module.scss";
import Select, { SelectOption } from "features/forms/Select";
import Spinny from "features/spinny/Spinny";
import { toastCategory } from "features/toast/Toasts";
import ToastContext from "features/toast/context";
import File from "components/forms/file/File";
import Input from "components/formsWithValidation/input/InputWithValidation";
import {
  useAddNodeConfigurationMutation,
  useGetNodeConfigurationQuery,
  useUpdateNodeConfigurationMutation,
  useUpdateNodeConfigurationStatusMutation,
  useUpdateNodePingSystemStatusMutation,
  useGetLndServicesQuery,
} from "apiSlice";
import { nodeConfiguration } from "apiTypes";
import classNames from "classnames";
import Collapse from "features/collapse/Collapse";
import Popover from "features/popover/Popover";
import Button, { ButtonPosition, ColorVariant } from "components/buttons/Button";
import Modal from "features/modal/Modal";
import Switch from "components/forms/switch/Switch";
import useTranslations from "services/i18n/useTranslations";
import Form from "components/forms/form/Form";
import Note, { NoteType } from "features/note/Note";
import ErrorSummary from "components/errors/ErrorSummary";
import { FormErrors, mergeServerError } from "components/errors/errors";

interface nodeProps {
  nodeId: number;
  collapsed?: boolean;
  addMode?: boolean;
  onAddSuccess?: () => void;
  onAddFailure?: () => void;
}

const nodeConfigurationTemplate = {
  createdOn: undefined,
  grpcAddress: "",
  macaroonFileName: "",
  name: "",
  tlsFileName: "",
  updatedOn: undefined,
  implementation: 0,
  nodeId: 0,
  status: 0,
  pingSystem: 0,
  customSettings: 0,
};

const importFailedPayments = "importFailedPayments";
const importFailedPaymentsValue = 1;
const importHtlcEvents = "importHtlcEvents";
const importHtlcEventsValue = 2;
const importPeerEvents = "importPeerEvents";
const importPeerEventsValue = 4;
const importTransactions = "importTransactions";
const importTransactionsValue = 8;
const importPayments = "importPayments";
const importPaymentsValue = 16;
const importInvoices = "importInvoices";
const importInvoicesValue = 32;
const importForwards = "importForwards";
const importForwardsValue = 64;
const importForwardsHistory = "importForwardsHistory";
const importForwardsHistoryValue = 128;

const customSettingsDefault = {
  importFailedPayments: true,
  importHtlcEvents: true,
  importPeerEvents: true,
  importTransactions: true,
  importPayments: true,
  importInvoices: true,
  importForwards: true,
  importForwardsHistory: true,
};

interface importProps {
  value: number;
  label?: string;
}

const customSettingsSidebarData = new Map<string, importProps>([
  [importFailedPayments, { value: importFailedPaymentsValue, label: undefined }],
  [importHtlcEvents, { value: importHtlcEventsValue, label: "Htlc events" }],
  [importPeerEvents, { value: importPeerEventsValue, label: "Peer events" }],
  [importTransactions, { value: importTransactionsValue, label: "Transactions" }],
  [importPayments, { value: importPaymentsValue, label: "Payments" }],
  [importInvoices, { value: importInvoicesValue, label: "Invoices" }],
  [importForwards, { value: importForwardsValue, label: "Forwards" }],
  [importForwardsHistory, { value: importForwardsHistoryValue, label: "Forwards History" }],
]);

const NodeSettings = React.forwardRef(function NodeSettings(
  { nodeId, collapsed, addMode, onAddSuccess }: nodeProps,
  ref
) {
  const { t } = useTranslations();
  const toastRef = React.useContext(ToastContext);
  const popoverRef = React.useRef();

  const { data: nodeConfigurationData } = useGetNodeConfigurationQuery(nodeId, {
    skip: !nodeId || nodeId == 0,
  });
  const { data: lndServicesData } = useGetLndServicesQuery(nodeId, {
    skip: !nodeId || nodeId == 0,
  });
  const [updateNodeConfiguration] = useUpdateNodeConfigurationMutation();
  const [addNodeConfiguration] = useAddNodeConfigurationMutation();
  const [setNodeConfigurationStatus] = useUpdateNodeConfigurationStatusMutation();
  const [setNodePingSystemStatus] = useUpdateNodePingSystemStatusMutation();

  const [nodeConfigurationState, setNodeConfigurationState] = useState<nodeConfiguration>(nodeConfigurationTemplate);
  const [collapsedState, setCollapsedState] = useState(collapsed ?? false);
  const [customSettingsCollapsedState, setCustomSettingsCollapsedState] = useState(true);
  const [showModalState, setShowModalState] = useState(false);
  const [deleteConfirmationTextInputState, setDeleteConfirmationTextInputState] = useState("");
  const [deleteEnabled, setDeleteEnabled] = useState(false);
  const [saveEnabledState, setSaveEnabledState] = useState(true);
  const [saveBootstrappingState, setSaveBootstrappingState] = useState(false);
  const [enableEnableButtonState, setEnableEnableButtonState] = useState(true);
  const [customSettingsState, setCustomSettingsState] = React.useState(customSettingsDefault);
  const [formErrorState, setFormErrorState] = React.useState({} as FormErrors);

  const { data: bootingCheck } = useGetLndServicesQuery(nodeId, {
    skip: !saveBootstrappingState,
    pollingInterval: 5 * 1000, // get status every 5 seconds
  });

  React.useImperativeHandle(ref, () => ({
    clear() {
      clear();
    },
  }));

  const clear = () => {
    setCustomSettingsState(customSettingsDefault);
    setNodeConfigurationState({
      grpcAddress: "",
      nodeId: 0,
      status: 0,
      implementation: 0,
      pingSystem: 0,
      name: "",
      customSettings: 0,
    } as nodeConfiguration);
  };

  React.useEffect(() => {
    if (collapsed != undefined) {
      setCollapsedState(collapsed);
    }
  }, [collapsed]);

  const handleConfirmationModalClose = () => {
    setShowModalState(false);
    setDeleteConfirmationTextInputState("");
    setDeleteEnabled(false);
    setNodeConfigurationState({
      implementation: 0,
      nodeId: 0,
      status: 0,
      pingSystem: 0,
      customSettings: 0,
    });
  };

  const handleDeleteClick = () => {
    if (popoverRef.current) {
      (popoverRef.current as { close: () => void }).close();
    }
    setShowModalState(true);
  };

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    submitNodeSettings();
  };

  const submitNodeSettings = async () => {
    setSaveEnabledState(false);
    const form = new FormData();
    form.append("implementation", "" + nodeConfigurationState.implementation);
    form.append("name", nodeConfigurationState.name ?? "");
    form.append("nodeId", "" + nodeConfigurationState.nodeId);
    form.append("status", "" + nodeConfigurationState.status);
    form.append("pingSystem", "" + nodeConfigurationState.pingSystem);
    form.append("customSettings", "" + nodeConfigurationState.customSettings);
    form.append("grpcAddress", nodeConfigurationState.grpcAddress ?? "");
    if (nodeConfigurationState.tlsFile) {
      form.append("tlsFile", nodeConfigurationState.tlsFile, nodeConfigurationState.tlsFileName);
    }
    if (nodeConfigurationState.macaroonFile) {
      form.append("macaroonFile", nodeConfigurationState.macaroonFile, nodeConfigurationState.macaroonFileName);
    }
    // we are adding new node
    if (!nodeConfigurationState.nodeId || nodeConfigurationState.nodeId == 0) {
      addNodeConfiguration(form)
        .unwrap()
        .then((_) => {
          setSaveEnabledState(false);
          setEnableEnableButtonState(true);
          toastRef?.current?.addToast("Local node added", toastCategory.success);
          if (onAddSuccess) {
            onAddSuccess();
          }
        })
        .catch((error) => {
          setSaveEnabledState(true);
          /* toastRef?.current?.addToast(error.data["errors"]["server"][0].split(":")[0], toastCategory.error); */
          const mergedErrors = mergeServerError(error.data, formErrorState);
          setFormErrorState(mergedErrors);
        });
      mixpanel.track("Add Local Node");
      return;
    } else {
      updateNodeConfiguration(form)
        .unwrap()
        .then((_) => {
          setSaveEnabledState(false);
          setSaveBootstrappingState(true);
          toastRef?.current?.addToast("Local node info saved", toastCategory.success);
        })
        .catch((error) => {
          setSaveEnabledState(true);
          /* toastRef?.current?.addToast(error.data["errors"]["server"][0].split(":")[0], toastCategory.error); */
          const mergedErrors = mergeServerError(error.data, formErrorState);
          setFormErrorState(mergedErrors);
        });
      mixpanel.track("Update Local Node", { nodeId: nodeConfigurationState.nodeId });
    }
  };

  React.useEffect(() => {
    setNodeConfigurationState(nodeConfigurationData || nodeConfigurationTemplate);
    if (nodeConfigurationData == undefined) {
      setNodeConfigurationState({ ...nodeConfigurationState, customSettings: 254 });
    } else {
      setCustomSettingsState({
        importFailedPayments:
          nodeConfigurationData.customSettings % (importFailedPaymentsValue * 2) >= importFailedPaymentsValue,
        importHtlcEvents: nodeConfigurationData.customSettings % (importHtlcEventsValue * 2) >= importHtlcEventsValue,
        importPeerEvents: nodeConfigurationData.customSettings % (importPeerEventsValue * 2) >= importPeerEventsValue,
        importTransactions:
          nodeConfigurationData.customSettings % (importTransactionsValue * 2) >= importTransactionsValue,
        importPayments: nodeConfigurationData.customSettings % (importPaymentsValue * 2) >= importPaymentsValue,
        importInvoices: nodeConfigurationData.customSettings % (importInvoicesValue * 2) >= importInvoicesValue,
        importForwards: nodeConfigurationData.customSettings % (importForwardsValue * 2) >= importForwardsValue,
        importForwardsHistory:
          nodeConfigurationData.customSettings % (importForwardsHistoryValue * 2) >= importForwardsHistoryValue,
      });
    }
    if (lndServicesData !== undefined && lndServicesData.status !== 1) {
      setSaveBootstrappingState(true);
      setSaveEnabledState(false);
    }
  }, [nodeConfigurationData, lndServicesData]);

  React.useEffect(() => {
    if (bootingCheck !== undefined && bootingCheck.status === 1) {
      setSaveBootstrappingState(false);
      setSaveEnabledState(true);
    }
  }, [bootingCheck]);

  const getCustomSettingsState = (key: string) => {
    const data = customSettingsSidebarData.get(key);
    if (data !== undefined && nodeConfigurationState != undefined) {
      return nodeConfigurationState.customSettings % (data.value * 2) >= data.value;
    }
    return true;
  };

  const toggleCustomSettingsState = (key: string) => {
    const data = customSettingsSidebarData.get(key);
    if (data !== undefined) {
      if (getCustomSettingsState(key)) {
        setNodeConfigurationState({
          ...nodeConfigurationState,
          customSettings: nodeConfigurationState.customSettings - data.value,
        });
      } else {
        setNodeConfigurationState({
          ...nodeConfigurationState,
          customSettings: nodeConfigurationState.customSettings + data.value,
        });
      }
    }
  };

  const handleTLSFileChange = (file: File | null) => {
    setNodeConfigurationState({ ...nodeConfigurationState, tlsFile: file, tlsFileName: file ? file.name : undefined });
  };

  const handleMacaroonFileChange = (file: File | null) => {
    setNodeConfigurationState({
      ...nodeConfigurationState,
      macaroonFile: file,
      macaroonFileName: file ? file.name : undefined,
    });
  };

  const handleAddressChange = (value: string) => {
    setNodeConfigurationState({ ...nodeConfigurationState, grpcAddress: value });
  };

  const handleNodeNameChange = (value: string) => {
    setNodeConfigurationState({ ...nodeConfigurationState, name: value });
  };

  const handleCollapseClick = () => {
    setCollapsedState(!collapsedState);
  };

  const handleCustomSettingsCollapseClick = () => {
    setCustomSettingsCollapsedState(!customSettingsCollapsedState);
  };

  const handleModalDeleteClick = () => {
    setShowModalState(false);
    setDeleteConfirmationTextInputState("");
    setDeleteEnabled(false);
    mixpanel.track("Delete Local Node", { nodeId: nodeConfigurationState.nodeId });
    setNodeConfigurationStatus({ nodeId: nodeConfigurationState.nodeId, status: 3 });
  };

  const handleDeleteConfirmationTextInputChange = (value: string) => {
    setDeleteConfirmationTextInputState(value as string);
    setDeleteEnabled(value.toLowerCase() === "delete");
  };

  const handleStatusClick = () => {
    setEnableEnableButtonState(false);
    let statusId = 0;
    if (nodeConfigurationState.status == 0) {
      statusId = 1;
      mixpanel.track("Enable Local Node", { nodeId: nodeConfigurationState.nodeId });
    } else {
      mixpanel.track("Disable Local Node", { nodeId: nodeConfigurationState.nodeId });
    }
    setNodeConfigurationStatus({ nodeId: nodeConfigurationState.nodeId, status: statusId })
      .unwrap()
      .finally(() => {
        setEnableEnableButtonState(true);
      });
    if (popoverRef.current) {
      (popoverRef.current as { close: () => void }).close();
    }
  };

  const handleAmbossPingClick = () => {
    const ambossActive = nodeConfigurationState.pingSystem % 2 >= 1;
    setNodePingSystemStatus({ nodeId: nodeConfigurationState.nodeId, pingSystem: 1, statusId: ambossActive ? 0 : 1 })
      .unwrap()
      .then((_) => {
        if (ambossActive) {
          setNodeConfigurationState({ ...nodeConfigurationState, pingSystem: nodeConfigurationState.pingSystem - 1 });
        } else {
          setNodeConfigurationState({ ...nodeConfigurationState, pingSystem: nodeConfigurationState.pingSystem + 1 });
        }
      })
      .catch((error) => {
        toastRef?.current?.addToast(error.data["errors"]["server"][0].split(":")[0], toastCategory.error);
      });
    if (popoverRef.current) {
      (popoverRef.current as { close: () => void }).close();
    }
  };

  const handleVectorPingClick = () => {
    const vectorActive = nodeConfigurationState.pingSystem % 4 >= 2;
    setNodePingSystemStatus({ nodeId: nodeConfigurationState.nodeId, pingSystem: 2, statusId: vectorActive ? 0 : 1 })
      .unwrap()
      .then((_) => {
        if (vectorActive) {
          setNodeConfigurationState({ ...nodeConfigurationState, pingSystem: nodeConfigurationState.pingSystem - 2 });
        } else {
          setNodeConfigurationState({ ...nodeConfigurationState, pingSystem: nodeConfigurationState.pingSystem + 2 });
        }
      })
      .catch((error) => {
        toastRef?.current?.addToast(error.data["errors"]["server"][0].split(":")[0], toastCategory.error);
      });
    if (popoverRef.current) {
      (popoverRef.current as { close: () => void }).close();
    }
  };

  const implementationOptions: Array<SelectOption> = [{ value: "0", label: "LND" }];

  const menuButton = <MoreIcon className={styles.moreIcon} />;
  return (
    <>
      {!addMode && (
        <div
          className={classNames(styles.header, { [styles.expanded]: !collapsedState })}
          onClick={handleCollapseClick}
        >
          <div
            className={classNames(styles.connectionIcon, {
              [styles.connected]: true,
              [styles.disabled]: nodeConfigurationState.status == 0,
            })}
          >
            {nodeConfigurationState.status == 0 && <DisconnectedIcon />}
            {nodeConfigurationState.status == 1 && <ConnectedIcon />}
          </div>
          <div className={styles.title}>{nodeConfigurationState?.name}</div>
          <div className={classNames(styles.collapseIcon, { [styles.collapsed]: collapsedState })}>
            {collapsedState ? <CollapsedIcon /> : <ExpandedIcon />}
          </div>
        </div>
      )}
      <Collapse collapsed={collapsedState} animate={!addMode}>
        <div className={classNames(styles.collapseContentWrappper, { [styles.addMode]: addMode })}>
          {!addMode && (
            <>
              <div className={styles.borderSection}>
                <div className={styles.detailHeader}>
                  <h4 className={styles.detailsTitle}>Node Details</h4>
                  <Popover button={menuButton} className={classNames("right", styles.moreButton)} ref={popoverRef}>
                    <div className={styles.nodeMenu}>
                      <Button
                        buttonColor={ColorVariant.warning}
                        icon={nodeConfigurationState.status == 0 ? <PlayIcon /> : <PauseIcon />}
                        onClick={handleStatusClick}
                        disabled={!enableEnableButtonState}
                      >
                        {nodeConfigurationState.status == 0 ? "Enable node" : "Disable node"}
                      </Button>
                      <Button
                        buttonColor={ColorVariant.error}
                        icon={<DeleteIcon />}
                        onClick={handleDeleteClick}
                        buttonPosition={ButtonPosition.fullWidth}
                      >
                        {"Delete node"}
                      </Button>
                    </div>
                  </Popover>
                </div>
              </div>
            </>
          )}
          <div className={""}>
            <Form onSubmit={handleSubmit}>
              <Select
                label={t.implementation}
                onChange={() => {
                  return;
                }}
                options={implementationOptions}
                value={implementationOptions.find((io) => io.value == "" + nodeConfigurationState.implementation)}
              />
              <span id="name">
                <Input
                  label={t.nodeName}
                  value={nodeConfigurationState.name}
                  type={"text"}
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleNodeNameChange(e.target.value)}
                  placeholder="Node 1"
                />
              </span>
              <span id="address">
                <Input
                  label={t.grpcAddress}
                  type={"text"}
                  value={nodeConfigurationState.grpcAddress}
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleAddressChange(e.target.value)}
                  placeholder="100.100.100.100:10009"
                  name="grpcAddress"
                  errors={formErrorState}
                />
              </span>
              <span id="tls">
                <File
                  label={t.tlsCertificate}
                  onFileChange={handleTLSFileChange}
                  fileName={nodeConfigurationState?.tlsFileName}
                />
              </span>
              <span id="macaroon">
                <File
                  label={t.macaroon}
                  onFileChange={handleMacaroonFileChange}
                  fileName={nodeConfigurationState?.macaroonFileName}
                />
              </span>
              <div className={styles.customImportSettings}>
                <div
                  className={classNames(styles.header, { [styles.expanded]: !customSettingsCollapsedState })}
                  onClick={handleCustomSettingsCollapseClick}
                >
                  <div className={styles.title}>{t.advancedSettings}</div>
                  <div
                    className={classNames(styles.collapseIcon, { [styles.collapsed]: customSettingsCollapsedState })}
                  >
                    {customSettingsCollapsedState ? <CollapsedIcon /> : <ExpandedIcon />}
                  </div>
                </div>
                <Collapse collapsed={customSettingsCollapsedState} animate={true}>
                  <div className={styles.customImportSettingsBody}>
                    {addMode &&
                      Object.keys(customSettingsState).map((key) => {
                        const k = key as keyof typeof customSettingsState;
                        const data = customSettingsSidebarData.get(key);
                        if (data !== undefined && data.label !== undefined) {
                          return (
                            <div className={styles.import} key={key}>
                              <Switch
                                label={data.label}
                                checked={nodeConfigurationState.customSettings % (data.value * 2) >= data.value}
                                onChange={() => {
                                  toggleCustomSettingsState(k);
                                }}
                              />
                            </div>
                          );
                        }
                      })}
                    {!addMode &&
                      Object.keys(customSettingsState).map((key) => {
                        const k = key as keyof typeof customSettingsState;
                        const data = customSettingsSidebarData.get(key);
                        if (data !== undefined && data.label !== undefined) {
                          return (
                            <div className={styles.import} key={key}>
                              <Switch
                                label={data.label || ""}
                                checked={customSettingsState[k]}
                                onChange={() => {
                                  setCustomSettingsState({
                                    ...customSettingsState,
                                    [key]: !customSettingsState[k],
                                  });
                                  toggleCustomSettingsState(key);
                                }}
                              />
                            </div>
                          );
                        }
                      })}
                    <Switch
                      label={t.ImportFailedPayments}
                      checked={
                        nodeConfigurationState.customSettings % (importFailedPaymentsValue * 2) >=
                        importFailedPaymentsValue
                      }
                      onChange={() => {
                        toggleCustomSettingsState(importFailedPayments);
                      }}
                    />
                    <div className={styles.importFailedPayments}>
                      <Note title={"Failed Payments"} noteType={NoteType.warning}>
                        {t.importFailedPayments}
                      </Note>
                    </div>
                  </div>
                </Collapse>
              </div>
              <ErrorSummary errors={formErrorState} />
              <Button
                id={"save-node"}
                buttonColor={ColorVariant.success}
                icon={saveEnabledState ? <SaveIcon /> : <Spinny />}
                onClick={submitNodeSettings}
                buttonPosition={ButtonPosition.fullWidth}
                disabled={!saveEnabledState}
              >
                {addMode
                  ? "Add Node"
                  : saveBootstrappingState
                  ? "Bootstrapping..."
                  : saveEnabledState
                  ? "Save node details"
                  : "Saving..."}
              </Button>
              {!addMode && (
                <div className={styles.pingSystems}>
                  <div className={styles.vectorPingSystem}>
                    <Switch
                      label="Vector Ping"
                      checked={nodeConfigurationState.pingSystem % 4 >= 2}
                      onChange={handleVectorPingClick}
                    />
                  </div>
                  <div className={styles.ambossPingSystem}>
                    <Switch
                      label="Amboss Ping"
                      checked={nodeConfigurationState.pingSystem % 2 >= 1}
                      onChange={handleAmbossPingClick}
                    />
                  </div>
                  <Note title={t.header.pingNoteHeader} noteType={NoteType.info}>
                    <p>{t.pingNote}</p>
                    <p>{t.header.pingSystem}</p>
                    <p>{t.header.vectorPingSystem}</p>
                    <p>{t.header.ambossPingSystem}</p>
                  </Note>
                </div>
              )}
            </Form>
          </div>
        </div>
      </Collapse>
      <Modal
        title={"Are you sure?"}
        icon={<DeleteIconHeader />}
        onClose={handleConfirmationModalClose}
        show={showModalState}
      >
        <div className={styles.deleteConfirm}>
          <p>
            Deleting the node will prevent you from viewing it&apos;s data in Torq. Alternatively set node to disabled
            to simply stop the data subscription but keep data collected so far.
          </p>
          <p>
            This operation cannot be undone, type &quot;<span className={styles.red}>delete</span>&quot; to confirm.
          </p>
          <Input
            placeholder={t.header.typeDeleteHere}
            value={deleteConfirmationTextInputState}
            type={"text"}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
              handleDeleteConfirmationTextInputChange(e.target.value)
            }
          />
          <div className={styles.deleteConfirmButtons}>
            <Button
              buttonColor={ColorVariant.error}
              buttonPosition={ButtonPosition.fullWidth}
              icon={<DeleteIcon />}
              onClick={handleModalDeleteClick}
              disabled={!deleteEnabled}
            >
              {t.delete}
            </Button>
          </div>
        </div>
      </Modal>
    </>
  );
});
export default NodeSettings;
