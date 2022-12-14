import { Tag24Regular as TagHeaderIcon } from "@fluentui/react-icons";
// import { useGetNodeConfigurationsQuery, WS_URL } from "apiSlice";
// import Button, { buttonColor, ButtonWrapper } from "components/buttons/Button";
// import ProgressHeader, { ProgressStepState, Step } from "features/progressTabs/ProgressHeader";
// import ProgressTabs, { ProgressTabContainer } from "features/progressTabs/ProgressTab";
import PopoutPageTemplate from "features/templates/popoutPageTemplate/PopoutPageTemplate";
// import { useState } from "react";
import { useNavigate } from "react-router";
// import useWebSocket from "react-use-websocket";
// import styles from "features/transact/newAddress/newAddress.module.scss";
import useTranslations from "services/i18n/useTranslations";
// import { nodeConfiguration } from "apiTypes";
// import Select from "features/forms/Select";
// import Note, { NoteType } from "features/note/Note";
// import {
//   DetailsContainer,
//   DetailsRowLinkAndCopy,
// } from "features/templates/popoutPageTemplate/popoutDetails/PopoutDetails";
// import { StatusIcon } from "features/templates/popoutPageTemplate/popoutDetails/StatusIcon";

function ModifyTagModal() {
  const { t } = useTranslations();

  // const [Create, setAddressTypeState] = useState(ProgressStepState.active);
  // const [doneState, setDoneState] = useState(ProgressStepState.disabled);
  // const [stepIndex, setStepIndex] = useState(0);
  //
  // function onNewAddressMessage(event: MessageEvent<string>) {
  //   const response = JSON.parse(event.data);
  //   setDoneState(ProgressStepState.completed);
  //   if (response?.type == "Error") {
  //     setNewAddressError(response.error);
  //     onNewAddressError(response as NewAddressError);
  //     return;
  //   }
  //   onNewAddressResponse(response as NewAddressResponse);
  // }
  //
  // function onNewAddressResponse(resp: NewAddressResponse) {
  //   console.log(resp.address.length);
  //   setResponse(resp);
  //   if (resp.address.length) {
  //     setDoneState(ProgressStepState.completed);
  //   } else {
  //     setNewAddressError(resp.failureReason);
  //     setDoneState(ProgressStepState.error);
  //   }
  // }
  //
  // function onNewAddressError(message: NewAddressError) {
  //   setDoneState(ProgressStepState.error);
  //   console.log("error", message);
  // }
  //
  // // This can also be an async getter function. See notes below on Async Urls.
  // const { sendJsonMessage } = useWebSocket(WS_URL, {
  //   //Will attempt to reconnect on all close events, such as server shutting down
  //   shouldReconnect: () => true,
  //   share: true,
  //   onMessage: onNewAddressMessage,
  // });

  const closeAndReset = () => {
    // setStepIndex(0);
    // setAddressTypeState(ProgressStepState.active);
    // setDoneState(ProgressStepState.disabled);
  };

  // const handleClickNext = (addType: AddressType) => {
  //   setStepIndex(1);
  //   setAddressTypeState(ProgressStepState.completed);
  //   setDoneState(ProgressStepState.processing);
  //   sendJsonMessage({
  //     reqId: "randId",
  //     type: "newAddress",
  //     newAddressRequest: {
  //       nodeId: selectedNodeId,
  //       type: addType,
  //       // TODO: account empty so the default wallet account is used
  //       // account: {account},
  //     },
  //   });
  // };

  const navigate = useNavigate();

  return (
    <PopoutPageTemplate title={t.createTag} show={true} onClose={() => navigate(-1)} icon={<TagHeaderIcon />}>
      {/*<ProgressHeader modalCloseHandler={closeAndReset}>*/}
      {/*  <Step label={t.addressType} state={addressTypeState} last={false} />*/}
      {/*  <Step label={"Done"} state={doneState} last={true} />*/}
      {/*</ProgressHeader>*/}

      {/*<ProgressTabs showTabIndex={stepIndex}>*/}
      {/*  <ProgressTabContainer>*/}
      {/*    <Select*/}
      {/*      label={t.yourNode}*/}
      {/*      onChange={(newValue: any) => {*/}
      {/*        setSelectedNodeId(newValue?.value || 0);*/}
      {/*      }}*/}
      {/*      options={nodeConfigurationOptions}*/}
      {/*      value={nodeConfigurationOptions.find((option) => option.value === selectedNodeId)}*/}
      {/*    />*/}
      {/*    <div className={styles.addressTypeWrapper}>*/}
      {/*      <div className={styles.addressTypes}>*/}
      {/*        {addressTypeOptions.map((addType, index) => {*/}
      {/*          return (*/}
      {/*            <Button*/}
      {/*              text={addType.label}*/}
      {/*              disabled={!selectedNodeId}*/}
      {/*              buttonColor={buttonColor.subtle}*/}
      {/*              key={index + addType.label}*/}
      {/*              onClick={() => {*/}
      {/*                if (selectedNodeId) {*/}
      {/*                  handleClickNext(addType.value);*/}
      {/*                }*/}
      {/*              }}*/}
      {/*            />*/}
      {/*          );*/}
      {/*        })}*/}
      {/*      </div>*/}
      {/*    </div>*/}
      {/*  </ProgressTabContainer>*/}

      {/*  <ProgressTabContainer>*/}
      {/*    {newAddressError && (*/}
      {/*      <Note title={"Error"} noteType={NoteType.error}>*/}
      {/*        {newAddressError}*/}
      {/*      </Note>*/}
      {/*    )}*/}
      {/*    {!newAddressError && <StatusIcon state={response?.address ? "success" : "processing"} />}*/}
      {/*    <DetailsContainer>*/}
      {/*      <DetailsRowLinkAndCopy label={"Address:"} copy={response?.address}>*/}
      {/*        {response?.address}*/}
      {/*      </DetailsRowLinkAndCopy>*/}
      {/*    </DetailsContainer>*/}
      {/*    <ButtonWrapper*/}
      {/*      className={styles.customButtonWrapperStyles}*/}
      {/*      rightChildren={*/}
      {/*        <Button*/}
      {/*          text={t.newAddress}*/}
      {/*          onClick={() => {*/}
      {/*            setAddressTypeState(ProgressStepState.active);*/}
      {/*            setDoneState(ProgressStepState.disabled);*/}
      {/*            setStepIndex(0);*/}
      {/*            setResponse(undefined);*/}
      {/*          }}*/}
      {/*          buttonColor={buttonColor.subtle}*/}
      {/*        />*/}
      {/*      }*/}
      {/*    />*/}
      {/*  </ProgressTabContainer>*/}
      {/*</ProgressTabs>*/}
    </PopoutPageTemplate>
  );
}

export default ModifyTagModal;
