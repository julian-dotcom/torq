import Box from "./Box";
import styles from "./NodeSettings.module.scss";
import Select, { SelectOption } from "features/forms/Select";
import React, { useState } from "react";
import {
  Save20Regular as SaveIcon,
  Play16Regular as ConnectedIcon,
  Pause16Regular as DisconnectedIcon,
  ChevronDown20Regular as CollapsedIcon,
  LineHorizontal120Regular as ExpandedIcon,
  MoreCircle20Regular as MoreIcon,
  Delete24Regular as DeleteIconHeader,
  Delete20Regular as DeleteIcon,
  Pause20Regular as PauseIcon,
  Play20Regular as PlayIcon,
} from "@fluentui/react-icons";
import Spinny from "features/spinny/Spinny";
import { toastCategory } from "features/toast/Toasts";
import ToastContext from "features/toast/context";
import File from "features/forms/File";
import TextInput from "features/forms/TextInput";
import {
  useGetNodeConfigurationQuery,
  useUpdateNodeConfigurationMutation,
  useAddNodeConfigurationMutation,
  useUpdateNodeConfigurationStatusMutation,
} from "apiSlice";
import { nodeConfiguration } from "apiTypes";
import classNames from "classnames";
import Collapse from "features/collapse/Collapse";
import Popover from "features/popover/Popover";
import Button, { buttonColor, buttonPosition } from "features/buttons/Button";
import Modal from "features/modal/Modal";

interface nodeProps {
  nodeId: number;
  collapsed?: boolean;
  addMode?: boolean;
  onAddSuccess?: () => void;
  onAddFailure?: () => void;
}

const NodeSettings = React.forwardRef(function NodeSettings(
  { nodeId, collapsed, addMode, onAddSuccess }: nodeProps,
  ref
) {
  const toastRef = React.useContext(ToastContext);
  const popoverRef = React.useRef();

  const { data: nodeConfigurationData } = useGetNodeConfigurationQuery(nodeId, {
    skip: !nodeId,
  });
  const [updateNodeConfiguration] = useUpdateNodeConfigurationMutation();
  const [addNodeConfiguration] = useAddNodeConfigurationMutation();
  const [setNodeConfigurationStatus] = useUpdateNodeConfigurationStatusMutation();

  const [nodeConfigurationState, setNodeConfigurationState] = useState({} as nodeConfiguration);
  const [collapsedState, setCollapsedState] = useState(collapsed ?? false);
  const [showModalState, setShowModalState] = useState(false);
  const [deleteConfirmationTextInputState, setDeleteConfirmationTextInputState] = useState("");
  const [deleteEnabled, setDeleteEnabled] = useState(false);
  const [saveEnabledState, setSaveEnabledState] = useState(true);
  const [enableEnableButtonState, setEnableEnableButtonState] = useState(true);

  React.useImperativeHandle(ref, () => ({
    clear() {
      clear();
    },
  }));

  const clear = () => {
    setNodeConfigurationState({ grpcAddress: "", implementation: "", name: "" } as nodeConfiguration);
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
    setNodeConfigurationState({} as nodeConfiguration);
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
    form.append("implementation", "0");
    form.append("name", nodeConfigurationState.name ?? "");
    form.append("grpcAddress", nodeConfigurationState.grpcAddress ?? "");
    if (nodeConfigurationState.tlsFile) {
      form.append("tlsFile", nodeConfigurationState.tlsFile, nodeConfigurationState.tlsFileName);
    }
    if (nodeConfigurationState.macaroonFile) {
      form.append("macaroonFile", nodeConfigurationState.macaroonFile, nodeConfigurationState.macaroonFileName);
    }
    // we are adding new node
    if (!nodeConfigurationState.nodeId) {
      addNodeConfiguration(form)
        .unwrap()
        .then((_) => {
          setSaveEnabledState(true);
          toastRef?.current?.addToast("Local node added", toastCategory.success);
          if (onAddSuccess) {
            onAddSuccess();
          }
        })
        .catch((error) => {
          setSaveEnabledState(true);
          toastRef?.current?.addToast(error.data["errors"]["server"][0].split(":")[0], toastCategory.error);
        });

      return;
    }
    updateNodeConfiguration({ form, nodeId: nodeConfigurationState.nodeId })
      .unwrap()
      .then((_) => {
        setSaveEnabledState(true);
        toastRef?.current?.addToast("Local node info saved", toastCategory.success);
      })
      .catch((error) => {
        setSaveEnabledState(true);
        toastRef?.current?.addToast(error.data["errors"]["server"][0].split(":")[0], toastCategory.error);
      });
  };

  React.useEffect(() => {
    if (nodeConfigurationData) {
      setNodeConfigurationState(nodeConfigurationData);
    }
  }, [nodeConfigurationData]);

  const handleTLSFileChange = (file: File | null) => {
    setNodeConfigurationState({ ...nodeConfigurationState, tlsFile: file, tlsFileName: file ? file.name : undefined });
  };

  const handleMacaroonFileChange = (file: File | null) => {
    setNodeConfigurationState({ ...nodeConfigurationState, macaroonFile: file, macaroonFileName: file ? file.name : undefined });
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

  const handleModalDeleteClick = () => {
    setShowModalState(false);
    setDeleteConfirmationTextInputState("");
    setDeleteEnabled(false);
    setNodeConfigurationStatus({ nodeId: nodeConfigurationState.nodeId, status: 3});
  };

  const handleDeleteConfirmationTextInputChange = (value: string) => {
    setDeleteConfirmationTextInputState(value as string);
    setDeleteEnabled(value.toLowerCase() === "delete");
  };

  const handleDisableClick = () => {
    setEnableEnableButtonState(false);
    if (nodeConfigurationState.status == 0) {
      nodeConfigurationState.status = 1
    } else {
      nodeConfigurationState.status = 0
    }
    setNodeConfigurationStatus({ nodeId: nodeConfigurationState.nodeId, status: nodeConfigurationState.status})
      .unwrap()
      .finally(() => {
        setEnableEnableButtonState(true);
      });
    if (popoverRef.current) {
      (popoverRef.current as { close: () => void }).close();
    }
  };

  const implementationOptions = [{ value: "0", label: "LND" } as SelectOption];

  const menuButton = <MoreIcon className={styles.moreIcon} />;
  return (
    <Box>
      <>
        {!addMode && (
          <div className={styles.header} onClick={handleCollapseClick}>
            <div
              className={classNames(styles.connectionIcon, {
                [styles.connected]: true,
                [styles.disabled]: nodeConfigurationState.status==1,
              })}
            >
              {nodeConfigurationState.status==0 && <ConnectedIcon />}
              {nodeConfigurationState.status==1 && <DisconnectedIcon />}
            </div>
            <div className={styles.title}>{nodeConfigurationState?.name}</div>
            <div className={classNames(styles.collapseIcon, { [styles.collapsed]: collapsedState })}>
              {collapsedState ? <CollapsedIcon /> : <ExpandedIcon />}
            </div>
          </div>
        )}
        <Collapse collapsed={collapsedState} animate={!addMode}>
          <>
            {!addMode && (
              <>
                <div className={styles.borderSection}>
                  <div className={styles.detailHeader}>
                    <h4 className={styles.detailsTitle}>Node Details</h4>
                    <Popover button={menuButton} className={classNames("right", styles.moreButton)} ref={popoverRef}>
                      <div className={styles.nodeMenu}>
                        <Button
                          buttonColor={buttonColor.secondary}
                          text={nodeConfigurationState.status==0 ? "Enable node" : "Disable node"}
                          icon={nodeConfigurationState.status==0 ? <PlayIcon /> : <PauseIcon />}
                          onClick={handleDisableClick}
                          disabled={!enableEnableButtonState}
                        />
                        <Button
                          buttonColor={buttonColor.warning}
                          text={"Delete node"}
                          icon={<DeleteIcon />}
                          onClick={handleDeleteClick}
                        />
                      </div>
                    </Popover>
                  </div>
                </div>
              </>
            )}
            <div className={""}>
              <form onSubmit={handleSubmit}>
                <Select
                  label="Implementation"
                  onChange={() => {
                    return;
                  }}
                  options={implementationOptions}
                  value={implementationOptions.find((io) => io.value === nodeConfigurationState.implementation)}
                />
                <span id="name">
                  <TextInput
                    label="Node Name"
                    value={nodeConfigurationState.name}
                    inputType="text"
                    onChange={(e) => handleNodeNameChange(e as string)}
                    placeholder="Node 1"
                  />
                </span>
                <span id="address">
                  <TextInput
                    label="GRPC Address (IP or Tor)"
                    value={nodeConfigurationState.grpcAddress}
                    onChange={(e) => handleAddressChange(e as string)}
                    placeholder="100.100.100.100:10009"
                  />
                </span>
                <span id="tls">
                  <File label="TLS Certificate" onFileChange={handleTLSFileChange} fileName={nodeConfigurationState?.tlsFileName} />
                </span>
                <span id="macaroon">
                  <File
                    label="Macaroon"
                    onFileChange={handleMacaroonFileChange}
                    fileName={nodeConfigurationState?.macaroonFileName}
                  />
                </span>
                <Button
                  id={"save-node"}
                  buttonColor={buttonColor.green}
                  text={addMode ? "Add Node" : saveEnabledState ? "Save node details" : "Saving..."}
                  icon={saveEnabledState ? <SaveIcon /> : <Spinny />}
                  onClick={submitNodeSettings}
                  buttonPosition={buttonPosition.fullWidth}
                  disabled={!saveEnabledState}
                />
              </form>
            </div>
          </>
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

            <TextInput
              value={deleteConfirmationTextInputState}
              onChange={(e) => handleDeleteConfirmationTextInputChange(e as string)}
            />
            <div className={styles.deleteConfirmButtons}>
              <Button
                buttonColor={buttonColor.warning}
                buttonPosition={buttonPosition.fullWidth}
                text={"Delete node"}
                icon={<DeleteIcon />}
                onClick={handleModalDeleteClick}
                disabled={!deleteEnabled}
              />
            </div>
          </div>
        </Modal>
      </>
    </Box>
  );
});
export default NodeSettings;
