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

    console.log("trying to submit");
    console.log(localState);
    const form = new FormData();
    form.append("implementation", "LND");
    form.append("grpcAddress", localState.grpcAddress ?? "");
    console.log(localState.tlsFile);
    form.append("tlsData", localState.tlsFile, localState.tlsFileName);
    form.append("macaroonData", localState.macaroonFile, localState.macaroonFileName);
    updateLocalNode(form);
    toastRef?.current?.addToast("Local node info saved", toastCategory.success);
  };

  React.useEffect(() => {
    console.log("I ram");
    console.log(localNodeData);
    console.log(localState);
    if (localNodeData) {
      console.log("going to set the state");
      setLocalState(localNodeData);
    }
    console.log(localNodeData);
    console.log(localState);
  }, [localNodeData]);

  const handleTLSFileChange = (file: File) => {
    console.log("change");
    console.log(localState);
    setLocalState({ ...localState, tlsFile: file, tlsFileName: file ? file.name : undefined });
  };

  const handleMacaroonFileChange = (file: File) => {
    console.log("change macaroon");
    console.log(localState);
    setLocalState({ ...localState, macaroonFile: file, macaroonFileName: file ? file.name : undefined });
  };

  const handleAddressChange = (value: string) => {
    console.log(" add change");
    console.log(localState);
    setLocalState({ ...localState, grpcAddress: value });
  };

  const implementationOptions = [{ value: "LND", label: "LND" } as SelectOption];
  return (
    <Box minWidth={440} title="Node Settings">
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
