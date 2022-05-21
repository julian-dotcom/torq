import Box from "./Box";
import SubmitButton from "../forms/SubmitButton";
import Select, { SelectOption } from "../forms/Select";
import React from "react";
import { Save20Regular as SaveIcon } from "@fluentui/react-icons";
import { toastCategory, addToastHandle } from "../toast/Toasts";
import ToastContext from "../toast/context";
import File from "../forms/File";
import TextInput from "../forms/TextInput";

function NodeSettings() {
  const toastRef = React.useContext(ToastContext);
  const submitNodeSettings = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    toastRef?.current?.addToast("Node settings saved", toastCategory.success);
  };

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
