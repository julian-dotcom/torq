import styles from "./NodeSettings.module.scss";
import Select from "components/forms/select/Select";
import { SelectOption } from "features/forms/Select";
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
import File from "components/forms/file/File";
import Input from "components/forms/input/Input";
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
import Button, { buttonColor, buttonPosition } from "components/buttons/Button";
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
    skip: !nodeId || nodeId == 0,
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
    setNodeConfigurationState({
      grpcAddress: "",
      nodeId: 0,
      status: 0,
      implementation: 0,
      name: "",
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
    form.append("implementation", "" + nodeConfigurationState.implementation);
    form.append("name", nodeConfigurationState.name ?? "");
    form.append("nodeId", "" + nodeConfigurationState.nodeId);
    form.append("status", "" + nodeConfigurationState.status);
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
    } else {
      updateNodeConfiguration(form)
        .unwrap()
        .then((_) => {
          setSaveEnabledState(true);
          toastRef?.current?.addToast("Local node info saved", toastCategory.success);
        })
        .catch((error) => {
          setSaveEnabledState(true);
          toastRef?.current?.addToast(error.data["errors"]["server"][0].split(":")[0], toastCategory.error);
        });
    }
  };

  React.useEffect(() => {
    if (nodeConfigurationData) {
      setNodeConfigurationState(nodeConfigurationData);
    } else {
      setNodeConfigurationState({ implementation: 0, nodeId: 0, status: 0 } as nodeConfiguration);
    }
  }, [nodeConfigurationData]);

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

  const handleModalDeleteClick = () => {
    setShowModalState(false);
    setDeleteConfirmationTextInputState("");
    setDeleteEnabled(false);
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

  const implementationOptions = [{ value: "0", label: "LND" } as SelectOption];

  const menuButton = <MoreIcon className={styles.moreIcon} />;
  return (
    <>
      {!addMode && (
        <div className={styles.header} onClick={handleCollapseClick}>
          <div
            className={classNames(styles.connectionIcon, {
              [styles.connected]: true,
              [styles.disabled]: nodeConfigurationState.status == 1,
            })}
          >
            {nodeConfigurationState.status == 0 && <ConnectedIcon />}
            {nodeConfigurationState.status == 1 && <DisconnectedIcon />}
          </div>
          <div className={styles.title}>{nodeConfigurationState?.name}</div>
          <div className={classNames(styles.collapseIcon, { [styles.collapsed]: collapsedState })}>
            {collapsedState ? <CollapsedIcon /> : <ExpandedIcon />}
          </div>
        </div>
      )}
      <Collapse collapsed={collapsedState} animate={!addMode}>
        <div className={classNames(styles.nodeDetailsCollapseContainer, { [styles.addMode]: addMode })}>
          {!addMode && (
            <>
              <div className={styles.borderSection}>
                <div className={styles.detailHeader}>
                  <h4 className={styles.detailsTitle}>Node Details</h4>
                  <Popover button={menuButton} className={classNames("right", styles.moreButton)} ref={popoverRef}>
                    <div className={styles.nodeMenu}>
                      <Button
                        buttonColor={buttonColor.secondary}
                        text={nodeConfigurationState.status == 1 ? "Enable node" : "Disable node"}
                        icon={nodeConfigurationState.status == 1 ? <PlayIcon /> : <PauseIcon />}
                        onClick={handleStatusClick}
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
          <div>
            <form className={styles.nodeForm} onSubmit={handleSubmit}>
              <Select
                label="Implementation"
                onChange={() => {
                  return;
                }}
                options={implementationOptions}
                value={implementationOptions.find((io) => io.value == "" + nodeConfigurationState.implementation)}
              />
              <Input
                label="Node Name"
                value={nodeConfigurationState.name}
                type={"text"}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleNodeNameChange(e.target.value)}
                placeholder="Node 1"
              />

              <Input
                label="GRPC Address (IP or Tor)"
                type={"text"}
                value={nodeConfigurationState.grpcAddress}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleAddressChange(e.target.value)}
                placeholder="100.100.100.100:10009"
              />

              <File
                label="TLS Certificate"
                onFileChange={handleTLSFileChange}
                fileName={nodeConfigurationState?.tlsFileName}
              />

              <File
                label="Macaroon"
                onFileChange={handleMacaroonFileChange}
                fileName={nodeConfigurationState?.macaroonFileName}
              />

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
            value={deleteConfirmationTextInputState}
            type={"text"}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
              handleDeleteConfirmationTextInputChange(e.target.value)
            }
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
  );
});
export default NodeSettings;
