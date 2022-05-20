import Box from "./Box";
import SubmitButton from "../forms/SubmitButton";
import Select, { SelectOption } from "../forms/Select";
import React from "react";
import { Save20Regular as SaveIcon } from "@fluentui/react-icons";
import { toastCategory, addToastHandle } from "../toast/Toasts";
import ToastContext from "../toast/context";

function NodeSettings() {
  const toastRef = React.useContext(ToastContext);
  const submitNodeSettings = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    toastRef?.current?.addToast("Node settings saved", toastCategory.success);
  };

  const clientOptions = [{ value: "LND", label: "LND" }];
  return (
    <Box minWidth={440} title="Node Settings">
      <form onSubmit={submitNodeSettings}>
        <Select label="Client" onChange={() => {}} options={clientOptions} value={clientOptions[0]} />
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
