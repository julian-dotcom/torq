import Box from "./Box";
import SubmitButton from "../forms/SubmitButton";
import Select, { SelectOption } from "../forms/Select";
import React from "react";
import { Save20Regular as SaveIcon } from "@fluentui/react-icons";
import { toastCategory, addToastHandle } from "../toast/Toasts";
import ToastContext from "../toast/context";
import File from "../forms/File";
import TextInput from "../forms/TextInput";
import { useGetLocalNodeQuery, useUpdateLocalNodeMutation } from "apiSlice";

function NodeSettings() {
  const toastRef = React.useContext(ToastContext);

  const { data: localNodeData } = useGetLocalNodeQuery();
  const [updateLocalNode] = useUpdateLocalNodeMutation();

  const submitNodeSettings = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    const form = new FormData();
    form.append("implementation", "LND");
    form.append("grpcAddress", "addr");
    form.append("tlsFileName", "txs file name");
    form.append("macaroonFileName", "mac file name");
    form.append("macaroonData", new Blob([new Uint8Array([2]).buffer]), "macaroonData");
    form.append("tlsData", new Blob([new Uint8Array([3]).buffer]), "somethingElse");
    updateLocalNode(form);
    toastRef?.current?.addToast("Local node info saved", toastCategory.success);
  };

  React.useEffect(() => {
    console.log(localNodeData);
  });

  const handleTLSFileChange = (file: File) => {
    console.log(file);
  };
  const handleMacaroonFileChange = (file: File) => {
    console.log(file);
  };

  const implementationOptions = [{ value: "LND", label: "LND" }];
  return (
    <Box minWidth={440} title="Node Settings">
      <form onSubmit={submitNodeSettings}>
        <Select
          label="Implementation"
          onChange={() => {}}
          options={implementationOptions}
          value={implementationOptions[0]}
        />
        <TextInput label="GRPC Address (IP or Tor)" />
        <File label="TLS Certificate" onFileChange={handleTLSFileChange} />
        <File label="Macaroon" onFileChange={handleMacaroonFileChange} />
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
