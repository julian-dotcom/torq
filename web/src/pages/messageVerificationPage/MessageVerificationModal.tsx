import {
  CheckmarkStarburst16Regular as VerifyButtonIcon,
  Signature16Regular as SignatureButtonIcon,
  Signature24Regular as SignatureIcon,
} from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import PopoutPageTemplate from "features/templates/popoutPageTemplate/PopoutPageTemplate";
import { Form, RadioChips, Select, TextArea } from "components/forms/forms";
import styles from "./message_verification.module.scss";
import Button, { ButtonPosition, ColorVariant } from "components/buttons/Button";
import { useNavigate } from "react-router";
import { useEffect, useRef, useState } from "react";
import classNames from "classnames";
import { useSignMessageMutation, useVerifyMessageMutation } from "./messageVerificationApi";
import { ActionMeta } from "react-select";
import { useGetNodeConfigurationsQuery } from "apiSlice";
import { IsNumericOption } from "utils/typeChecking";
import { userEvents } from "utils/userEvents";

export default function MessageVerificationModal() {
  const { t } = useTranslations();
  const { track } = userEvents();
  const navigate = useNavigate();

  const { data: nodeConfigurations } = useGetNodeConfigurationsQuery();

  let nodeConfigurationOptions: Array<{ value: number; label: string | undefined }> = [{ value: 0, label: undefined }];
  if (nodeConfigurations) {
    nodeConfigurationOptions = nodeConfigurations.map((nodeConfiguration) => {
      return { value: nodeConfiguration.nodeId, label: nodeConfiguration.name };
    });
  }
  const [selectedNodeId, setSelectedNodeId] = useState<number>(0);
  const [currentAction, setCurrentAction] = useState<"sign" | "verify">("sign");
  const [verifyMessage, verifyMessageResponse] = useVerifyMessageMutation();
  const [signMessage, signMessageResponse] = useSignMessageMutation();
  const [emptySignatureField, setEmptySignatureField] = useState(false);
  const [emptyMessageField, setEmptyMessageField] = useState(false);
  const [emptyMessageSignField, setEmptyMessageSignField] = useState(false);

  const formRef = useRef<HTMLFormElement>(null);
  const buttonRef = useRef<HTMLButtonElement>(null);

  useEffect(() => {
    if (nodeConfigurationOptions !== undefined) {
      setSelectedNodeId(nodeConfigurationOptions[0].value);
    }
  }, [nodeConfigurationOptions]);

  const closeAndReset = () => {
    navigate(-1);
  };

  function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (currentAction === "sign") {
      setEmptyMessageSignField(false);
      signMessage({ nodeId: selectedNodeId, message: formRef.current?.signMessage?.value });
      track("Sign Message", { nodeId: selectedNodeId });
    } else {
      setEmptySignatureField(false);
      setEmptyMessageField(false);
      verifyMessage({
        nodeId: selectedNodeId,
        message: formRef.current?.message?.value,
        signature: formRef.current?.signature?.value && formRef.current.signature.value.trim(),
      });
      track("Verify Message", { nodeId: selectedNodeId });
    }
  }

  function handleRadioChange(event: React.ChangeEvent<HTMLInputElement>) {
    if (event.target.id === "sign" || event.target.id === "verify") {
      setCurrentAction(event.target.id);
      track(event.target.id === "sign" ? "Select Sign Message" : "Select Verify Message");
    }
  }

  const textAreaKeyboardSubmit = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    // TODO: This logic should be replaced with normal html5 form validation
    if (e.currentTarget.id === "message" && e.currentTarget.value !== "") {
      setEmptyMessageField(false);
    }
    if (e.currentTarget.id === "signature" && e.currentTarget.value !== "") {
      setEmptySignatureField(false);
    }
    if (e.currentTarget.id === "signMessage" && e.currentTarget.value !== "") {
      setEmptyMessageSignField(false);
    }

    // Press cmd+enter or ctrl+enter to submit the form. This is a workaround for the fact that
    // the form is not submitted when pressing cmd+enter in a textarea.
    if (e.metaKey && e.key === "Enter") {
      buttonRef?.current && formRef?.current && formRef.current.requestSubmit(buttonRef.current);
    }
  };

  return (
    <PopoutPageTemplate title={t.MessageVerification} show={true} icon={<SignatureIcon />} onClose={closeAndReset}>
      <div className={styles.activeColumns}>
        <Form onSubmit={handleSubmit} name={"messageVerificationForm"} ref={formRef}>
          <RadioChips
            groupName={"action"}
            options={[
              { label: "Sign Message", checked: currentAction === "sign", onChange: handleRadioChange, id: "sign" },
              {
                label: "Verify Signature",
                checked: currentAction === "verify",
                onChange: handleRadioChange,
                id: "verify",
              },
            ]}
          />

          <Select
            label={t.node}
            autoFocus={true}
            defaultValue={nodeConfigurationOptions[0]}
            onChange={(newValue: unknown, _: ActionMeta<unknown>) => {
              if (IsNumericOption(newValue)) {
                setSelectedNodeId(newValue.value);
              }
            }}
            options={nodeConfigurationOptions}
            value={nodeConfigurationOptions.find((option) => option.value === selectedNodeId)}
          />
          <div className={classNames(styles.signMessageWrapper, { [styles.hidden]: currentAction !== "sign" })}>
            <TextArea
              rows={6}
              label={t.message}
              onKeyDown={textAreaKeyboardSubmit}
              name={"signMessage"}
              required={currentAction === "sign"}
              onInvalid={() => setEmptyMessageSignField(true)}
              errorText={emptyMessageSignField ? t.missingSignMessage : undefined}
            />
            <TextArea
              rows={3}
              label={t.signature}
              disabled={signMessageResponse?.data?.signature === undefined}
              placeholder={"Unsigned"}
              value={signMessageResponse?.data?.signature}
            />
          </div>

          <div className={classNames(styles.verifyMessageWrapper, { [styles.hidden]: currentAction !== "verify" })}>
            <TextArea
              rows={6}
              id={"message"}
              label={t.message}
              autoFocus={true}
              errorText={emptyMessageField ? t.missingMessage : undefined}
              required={currentAction === "verify"}
              onInvalid={() => setEmptyMessageField(true)}
              onKeyDown={textAreaKeyboardSubmit}
            />
            <TextArea
              rows={3}
              id={"signature"}
              label={t.signature}
              errorText={emptySignatureField ? t.missingSignature : undefined}
              required={currentAction === "verify"}
              onInvalid={() => setEmptySignatureField(true)}
              onKeyDown={textAreaKeyboardSubmit}
            />
            {verifyMessageResponse?.data && (
              <div
                className={classNames(styles.resultWrapper, {
                  [styles.verified]: verifyMessageResponse?.data?.valid,
                  [styles.invalid]: !verifyMessageResponse?.data?.valid,
                })}
              >
                <div>{verifyMessageResponse?.data?.valid ? t.validSignature : t.invalidSignature}</div>
                {/*<div>{verifyMessageResponse?.data?.valid ? <CheckIcon /> : <CloseIcon />}</div>*/}
                <div>{verifyMessageResponse?.data?.pubKey}</div>
              </div>
            )}
          </div>

          <Button
            type={"submit"}
            buttonColor={ColorVariant.primary}
            buttonPosition={ButtonPosition.fullWidth}
            icon={currentAction === "sign" ? <SignatureButtonIcon /> : <VerifyButtonIcon />}
            ref={buttonRef}
          >
            {currentAction === "sign" ? t.sign : t.verify}
          </Button>
        </Form>
      </div>
    </PopoutPageTemplate>
  );
}
