import Box from "./Box";
import styles from "./NodeSettings.module.scss";
import SubmitButton from "../forms/SubmitButton";
import Select, { SelectOption } from "../forms/Select";
import React, { useState } from "react";
import {
  Save20Regular as SaveIcon,
  PlugConnected20Regular as ConnectedIcon,
  PlugDisconnected20Regular as DisconnectedIcon,
  ChevronDown20Regular as ChevronIcon,
} from "@fluentui/react-icons";
import { toastCategory } from "../toast/Toasts";
import ToastContext from "../toast/context";
import File from "../forms/File";
import TextInput from "../forms/TextInput";
import { useGetLocalNodeQuery, useUpdateLocalNodeMutation } from "apiSlice";
import { localNode } from "apiTypes";
import classNames from "classnames";
import Collapse from "features/collapse/Collapse";

interface nodeProps {
  localNodeId: number;
}
function NodeSettings({ localNodeId }: nodeProps) {
  const toastRef = React.useContext(ToastContext);

  const { data: localNodeData } = useGetLocalNodeQuery(localNodeId);
  const [updateLocalNode] = useUpdateLocalNodeMutation();

  const [localState, setLocalState] = useState({} as localNode);
  const [collapsedState, setCollapsedState] = useState(false);

  const submitNodeSettings = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = new FormData();
    form.append("implementation", "LND");
    form.append("grpcAddress", localState.grpcAddress ?? "");
    if (localState.tlsFile) {
      form.append("tlsFile", localState.tlsFile, localState.tlsFileName);
    }
    if (localState.macaroonFile) {
      form.append("macaroonFile", localState.macaroonFile, localState.macaroonFileName);
    }
    updateLocalNode(form);
    toastRef?.current?.addToast("Local node info saved", toastCategory.success);
  };

  React.useEffect(() => {
    if (localNodeData) {
      setLocalState(localNodeData);
    }
  }, [localNodeData]);

  const handleTLSFileChange = (file: File) => {
    setLocalState({ ...localState, tlsFile: file, tlsFileName: file ? file.name : undefined });
  };

  const handleMacaroonFileChange = (file: File) => {
    setLocalState({ ...localState, macaroonFile: file, macaroonFileName: file ? file.name : undefined });
  };

  const handleAddressChange = (value: string) => {
    setLocalState({ ...localState, grpcAddress: value });
  };

  const handleCollapseClick = () => {
    setCollapsedState(!collapsedState);
  };

  const implementationOptions = [{ value: "LND", label: "LND" } as SelectOption];
  return (
    <Box title="Node Settings">
      <div className={styles.container}>
        <div className={styles.header}>
          <div className={styles.connectionIcon}>
            <ConnectedIcon />
          </div>
          <div>LN.Capital [1] LN Capital</div>
          <div className={classNames(styles.collapseIcon, { [styles.collapsed]: collapsedState })}>
            <ChevronIcon onClick={handleCollapseClick} />
          </div>
        </div>
        <Collapse collapsed={collapsedState}>
          <div className={styles.body}>
            <form onSubmit={submitNodeSettings}>
              <Select
                label="Implementation"
                onChange={() => {}}
                options={implementationOptions}
                value={implementationOptions.find((io) => io.value === localState?.implementation)}
              />
              <span id="address">
                <TextInput
                  label="GRPC Address (IP or Tor)"
                  value={localState?.grpcAddress}
                  onChange={handleAddressChange}
                  placeholder="100.100.100.100:10009"
                />
              </span>
              <span id="tls">
                <File label="TLS Certificate" onFileChange={handleTLSFileChange} fileName={localState?.tlsFileName} />
              </span>
              <span id="macaroon">
                <File
                  label="Macaroon"
                  onFileChange={handleMacaroonFileChange}
                  fileName={localState?.macaroonFileName}
                />
              </span>
              <SubmitButton>
                <React.Fragment>
                  <SaveIcon />
                  Save node details
                </React.Fragment>
              </SubmitButton>
            </form>
          </div>
        </Collapse>
      </div>
    </Box>
  );
}
export default NodeSettings;
