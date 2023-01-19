import {
  CheckmarkStarburst16Regular as VerifyButtonIcon,
  Signature16Regular as SignatureButtonIcon,
  Signature24Regular as SignatureIcon,
} from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import PopoutPageTemplate from "features/templates/popoutPageTemplate/PopoutPageTemplate";
import { Form, RadioChips, TextArea } from "components/forms/forms";
import styles from "./message_verification.module.scss";
import Button, { ButtonPosition, ColorVariant } from "components/buttons/Button";
import { useNavigate } from "react-router";
import { useRef, useState } from "react";
import classNames from "classnames";
import { useSignMessageMutation, useVerifyMessageMutation } from "./messageVerificationApi";

export default function MessageVerificationModal() {
  const { t } = useTranslations();
  const navigate = useNavigate();
  const [currentAction, setCurrentAction] = useState<"sign" | "verify">("sign");
  const [verifyMessage] = useVerifyMessageMutation();
  const [signMessage] = useSignMessageMutation();
  const [emptySignatureField, setEmptySignatureField] = useState(false);
  const [emptyMessageField, setEmptyMessageField] = useState(false);
  const formRef = useRef<HTMLFormElement>(null);
  const buttonRef = useRef<HTMLButtonElement>(null);

  const closeAndReset = () => {
    navigate(-1);
  };

  function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (currentAction === "sign") {
      signMessage({ nodeId: 1, message: "test" });
    } else {
      verifyMessage({
        nodeId: 1,
        message: "test",
        signature: "test",
      });
    }
  }

  function handleRadioChange(event: React.ChangeEvent<HTMLInputElement>) {
    if (event.target.id === "sign" || event.target.id === "verify") {
      setCurrentAction(event.target.id);
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

          <div className={classNames(styles.signMessageWrapper, { [styles.hidden]: currentAction !== "sign" })}>
            <TextArea rows={6} label={t.message} onKeyDown={textAreaKeyboardSubmit} autoFocus={true} />
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
            <div className={styles.resultWrapper}>{t.validSignature}</div>
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
