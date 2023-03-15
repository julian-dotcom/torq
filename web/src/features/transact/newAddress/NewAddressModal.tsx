import { Copy20Regular as CopyIcon, Link20Regular as TransactionIconModal } from "@fluentui/react-icons";
import { useGetNodeConfigurationsQuery } from "apiSlice";
import Button, { ButtonPosition, ColorVariant, SizeVariant } from "components/buttons/Button";
import PopoutPageTemplate from "features/templates/popoutPageTemplate/PopoutPageTemplate";
import { useContext, useState } from "react";
import { useNavigate } from "react-router";
import styles from "features/transact/newAddress/newAddress.module.scss";
import useTranslations from "services/i18n/useTranslations";
import { nodeConfiguration } from "apiTypes";
import Select from "features/forms/Select";
import mixpanel from "mixpanel-browser";
import { useNewAddressMutation } from "./newAddressApi";
import { AddressType } from "./newAddressTypes";
import Note, { NoteType } from "features/note/Note";
import ToastContext from "features/toast/context";
import { toastCategory } from "features/toast/Toasts";
import Spinny from "features/spinny/Spinny";
import { ServerErrorType } from "components/errors/errors";
import ErrorSummary from "components/errors/ErrorSummary";

function NewAddressModal() {
  const { t } = useTranslations();
  const toastRef = useContext(ToastContext);

  const { data: nodeConfigurations } = useGetNodeConfigurationsQuery();

  interface Option {
    label: string;
    value: number;
  }

  let nodeConfigurationOptions: Array<Option> = [{ value: 0, label: "Select a local node" }];
  if (nodeConfigurations) {
    nodeConfigurationOptions = nodeConfigurations.map((nodeConfiguration: nodeConfiguration) => {
      return { value: nodeConfiguration.nodeId, label: nodeConfiguration.name ?? "" };
    });
  }

  const addressTypeOptions = [
    { label: t.p2wpkh, value: AddressType.P2WPKH }, // Wrapped Segwit
    { label: t.p2wkh, value: AddressType.P2WKH }, // Segwit
    { label: t.p2tr, value: AddressType.P2TR }, // Taproot
  ];

  const [selectedNodeId, setSelectedNodeId] = useState<number>(nodeConfigurationOptions[0].value);

  const [newAddress, { error, data, isLoading, isSuccess, isError, isUninitialized }] = useNewAddressMutation();

  const handleClickNext = (addType: AddressType) => {
    newAddress({
      nodeId: selectedNodeId,
      type: addType,
      // account: {account},
    });
  };

  const navigate = useNavigate();

  return (
    <PopoutPageTemplate
      title={t.header.newAddress}
      show={true}
      onClose={() => navigate(-1)}
      icon={<TransactionIconModal />}
    >
      <div className={styles.nodeSelectionWrapper}>
        <div className={styles.nodeSelection}>
          <Select
            label={t.yourNode}
            onChange={(newValue: unknown) => {
              const value = newValue as Option;
              if (value && value.value > 0) setSelectedNodeId(value.value);
            }}
            options={nodeConfigurationOptions}
            value={nodeConfigurationOptions.find((option) => option.value === selectedNodeId)}
          />
        </div>
      </div>
      <div className={styles.addressTypeWrapper}>
        <div className={styles.addressTypes}>
          {addressTypeOptions.map((addType, index) => {
            return (
              <Button
                disabled={isLoading}
                buttonColor={ColorVariant.primary}
                key={index + addType.label}
                icon={isLoading && <Spinny />}
                onClick={() => {
                  if (selectedNodeId) {
                    handleClickNext(addType.value);
                    mixpanel.track("Select Address Type", { addressType: addType.label });
                  }
                }}
              >
                {addType.label}
              </Button>
            );
          })}
        </div>
      </div>
      <div className={styles.addressResultWrapper}>
        {isUninitialized && (
          <Note
            title={t.newAddress}
            noteType={isLoading ? NoteType.warning : NoteType.info}
            icon={<TransactionIconModal />}
          >
            {isLoading ? t.loading : t.selectAddressType}
          </Note>
        )}
        {data?.address && (
          <Note title={t.newAddress} noteType={NoteType.success} icon={<TransactionIconModal />}>
            {data?.address || t.selectAddressType}
          </Note>
        )}
        {data?.address && isSuccess && (
          <Button
            buttonColor={ColorVariant.success}
            buttonSize={SizeVariant.normal}
            buttonPosition={ButtonPosition.fullWidth}
            icon={<CopyIcon />}
            onClick={() => {
              if (data?.address) {
                toastRef?.current?.addToast("Copied to clipboard", toastCategory.success);
                navigator.clipboard.writeText(data?.address);
              }
            }}
          >
            {t.Copy}
          </Button>
        )}
        {isError && <ErrorSummary title={t.error} errors={(error as { data: ServerErrorType })?.data?.errors} />}
      </div>
    </PopoutPageTemplate>
  );
}

export default NewAddressModal;
