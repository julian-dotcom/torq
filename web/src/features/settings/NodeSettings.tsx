import Box from "./Box";
import SubmitButton from "../forms/SubmitButton";
import Select, { SelectOption } from "../forms/Select";
import React, { useState } from "react";
import { Save20Regular as SaveIcon } from "@fluentui/react-icons";
import { toastCategory } from "../toast/Toasts";
import ToastContext from "../toast/context";
import File from "../forms/File";
import TextInput from "../forms/TextInput";
import { useGetLocalNodeQuery, useUpdateLocalNodeMutation } from "apiSlice";
import { localNode } from "apiTypes";

function NodeSettings() {
  const toastRef = React.useContext(ToastContext);

  const { data: localNodeData } = useGetLocalNodeQuery();
  const [updateLocalNode] = useUpdateLocalNodeMutation();

  const [localState, setLocalState] = useState({} as localNode);

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

  const implementationOptions = [{ value: "LND", label: "LND" } as SelectOption];
  return (
    <Box title="Node Settings">
      <form onSubmit={submitNodeSettings}>
        <Select
          label="Implementation"
          onChange={() => {}}
          options={implementationOptions}
          value={implementationOptions.find((io) => io.value === localState?.implementation)}
        />
        <TextInput
          label="GRPC Address (IP or Tor)"
          value={localState?.grpcAddress}
          onChange={handleAddressChange}
          placeholder="100.100.100.100:10009"
        />
        <File label="TLS Certificate" onFileChange={handleTLSFileChange} fileName={localState?.tlsFileName} />
        <File label="Macaroon" onFileChange={handleMacaroonFileChange} fileName={localState?.macaroonFileName} />
        <SubmitButton>
          <React.Fragment>
            <SaveIcon />
            Save
          </React.Fragment>
        </SubmitButton>
      </form>
    </Box>
  );
}
export default NodeSettings;
