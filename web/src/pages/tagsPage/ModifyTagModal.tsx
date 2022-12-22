import { Tag24Regular as TagHeaderIcon } from "@fluentui/react-icons";
import { WS_URL } from "apiSlice";
import Button, { buttonColor, ButtonWrapper } from "components/buttons/Button";
import ProgressHeader, { ProgressStepState, Step } from "features/progressTabs/ProgressHeader";
import ProgressTabs, { ProgressTabContainer } from "features/progressTabs/ProgressTab";
import PopoutPageTemplate from "features/templates/popoutPageTemplate/PopoutPageTemplate";
import { ChangeEvent, useState } from "react";
import { useNavigate } from "react-router";
import useWebSocket from "react-use-websocket";
import styles from "features/transact/newAddress/newAddress.module.scss";
import useTranslations from "services/i18n/useTranslations";
import Select, { SelectOptions } from "features/forms/Select";
import { ActionMeta } from "react-select";
// import { nodeConfiguration } from "apiTypes";
import Note, { NoteType } from "features/note/Note";
import {
  DetailsContainer,
  DetailsRowLinkAndCopy,
} from "features/templates/popoutPageTemplate/popoutDetails/PopoutDetails";
import { StatusIcon } from "features/templates/popoutPageTemplate/popoutDetails/StatusIcon";
import { NewAddressResponse, NewAddressError, AddressType } from "features/transact/newAddress/NewAddressModal"
import FormRow from "features/forms/FormWrappers";
import Input from "components/forms/input/Input";
import { useGetCategoriesQuery } from "pages/categoriesPage/categoriesApi";
import { Category } from "pages/categoriesPage/categoriesTypes";
import { useGetNodesChannelsQuery } from "./tagsApi";
import { ChannelNode, ChannelForTag, NodeForTag } from "./tagsTypes"

function ModifyTagModal() {
  const { t } = useTranslations();

  const [createTageState, setCreateTagState] = useState(ProgressStepState.active);
  const [updateTagState, setUpdateTagState] = useState(ProgressStepState.active);
  const [doneState, setDoneState] = useState(ProgressStepState.disabled);
  const [stepIndex, setStepIndex] = useState(0);
  const [response, setResponse] = useState<NewAddressResponse>();
  const [newAddressError, setNewAddressError] = useState("");
  const [tagName, setTagName] = useState<string>("");
  const tagColorOptions: SelectOptions[] = [
    { value: 0, label: "Select your color" },
    { value: 1, label: "Yellow" },
    { value: 2, label: "Red" },
    { value: 3, label: "Orange" },
  ];

  const [selectedTagColor, setTagColor] = useState<number>(tagColorOptions[0].value as number);

  const { data: categoriesResponse }  = useGetCategoriesQuery<{
    data: Array<Category>;
    isLoading: boolean;
    isFetching: boolean;
    isUninitialized: boolean;
    isSuccess: boolean;
  }>();
  let tagCategorieOptions: SelectOptions[] = [{ value: 0, label: "Select your Category" }];
  if (categoriesResponse !== undefined) {
    tagCategorieOptions = categoriesResponse.map((tagCategorieOptions: Category) => {
      return { value: tagCategorieOptions.categoryId, label: tagCategorieOptions.name };
    });
  }
  const [selectedTagCategory, setTagCategory] = useState<number>(tagCategorieOptions[0].value as number);

  const { data: channelsNodesResponse } = useGetNodesChannelsQuery<{
    data: ChannelNode;
    isLoading: boolean;
    isFetching: boolean;
    isUninitialized: boolean;
    isSuccess: boolean;
  }>();
  console.log('channelsNodesResponse', channelsNodesResponse)
  let channelsNodesOptions: SelectOptions[] = [{ value: 0, label: "Channel or Node" }];
  let channelsOptions: SelectOptions[] = [];
  let nodesOptions: SelectOptions[] = [];
  if (channelsNodesResponse !== undefined) {
    channelsOptions = channelsNodesResponse.nodes.map((nodesOptions: NodeForTag) => {
      return { value: nodesOptions.nodeId, label: nodesOptions.alias };
    });
    nodesOptions = channelsNodesResponse.channels.map((channelsOptions: ChannelForTag) => {
      return { value: channelsOptions.shortChannelId, label: channelsOptions.shortChannelId };
    });
  }
  channelsNodesOptions = nodesOptions.concat(channelsOptions)
  console.log('channelsNodesOptions', channelsNodesOptions)
  const [selectedTarget, setTarget] = useState<number>(channelsNodesOptions[0].value as number);


  function onNewAddressMessage(event: MessageEvent<string>) {
    const response = JSON.parse(event.data);
    setDoneState(ProgressStepState.completed);
    if (response?.type == "Error") {
      setNewAddressError(response.error);
      onNewAddressError(response as NewAddressError);
      return;
    }
    onNewAddressResponse(response as NewAddressResponse);
  }

  function onNewAddressResponse(resp: NewAddressResponse) {
    console.log(resp.address.length);
    setResponse(resp);
    if (resp.address.length) {
      setDoneState(ProgressStepState.completed);
    } else {
      setNewAddressError(resp.failureReason);
      setDoneState(ProgressStepState.error);
    }
  }

  function onNewAddressError(message: NewAddressError) {
    setDoneState(ProgressStepState.error);
    console.log("error", message);
  }

  // This can also be an async getter function. See notes below on Async Urls.
  const { sendJsonMessage } = useWebSocket(WS_URL, {
    //Will attempt to reconnect on all close events, such as server shutting down
    shouldReconnect: () => true,
    share: true,
    onMessage: onNewAddressMessage,
  });

  const closeAndReset = () => {
    setStepIndex(0);
    setCreateTagState(ProgressStepState.active);
    setDoneState(ProgressStepState.disabled);
  };

  const handleClickNext = (addType: AddressType) => {
    setStepIndex(1);
    setCreateTagState(ProgressStepState.completed);
    setDoneState(ProgressStepState.processing);
    sendJsonMessage({
      reqId: "randId",
      type: "newAddress",
      newAddressRequest: {
        // nodeId: selectedNodeId,
        type: addType,
        // TODO: account empty so the default wallet account is used
        // account: {account},
      },
    });
  };

  const navigate = useNavigate();

  return (
    <PopoutPageTemplate title={t.createTag} show={true} onClose={() => navigate(-1)} icon={<TagHeaderIcon />}>
      <ProgressHeader modalCloseHandler={closeAndReset}>
        <Step label={t.customTag} state={createTageState} last={false} />
        <Step label={"Done"} state={doneState} last={true} />
      </ProgressHeader>

      <ProgressTabs showTabIndex={stepIndex}>
        <ProgressTabContainer>
          <div className={styles.activeColumns}>
            <FormRow>
              <div className={styles.updateChannelTableDouble}>
                <div className={styles.input}>
                  <Input
                    label={"Name"}
                    formatted={true}
                    className={styles.simple}
                    value={tagName}
                    onChange={(e: ChangeEvent<HTMLInputElement>) => {
                      setTagName(e.target.value);
                    }}
                  />
                </div>
              </div>
            </FormRow>
            <FormRow>
              <div className={styles.updateChannelTableDouble}>
                <Select
                  label={t.tagColor}
                  onChange={(newValue: unknown, _: ActionMeta<unknown>) => {
                    const selectOptions = newValue as SelectOptions;
                    setTagColor(selectOptions?.value as number);
                  }}
                  options={tagColorOptions}
                  value={tagColorOptions.find((option) => option.value === selectedTagColor)}
                />
              </div>
              <div className={styles.updateChannelTableDouble}>
                <Select
                  label={t.category}
                  onChange={(newValue: unknown, _: ActionMeta<unknown>) => {
                    const selectOptions = newValue as SelectOptions;
                    setTagCategory(selectOptions?.value as number);
                  }}
                  options={tagCategorieOptions}
                  value={tagCategorieOptions.find((option) => option.value === selectedTagCategory)}
                />
              </div>
            </FormRow>
            <FormRow>
            <div className={styles.openChannelTableRow}>
                <Select
                  label={t.tagColor}
                  onChange={(newValue: unknown, _: ActionMeta<unknown>) => {
                    const selectOptions = newValue as SelectOptions;
                    setTarget(selectOptions?.value as number);
                  }}
                  options={channelsNodesOptions}
                  value={channelsNodesOptions.find((option) => option.value === selectedTarget)}
                />
              </div>
            </FormRow>

          </div>
       </ProgressTabContainer>

       <ProgressTabContainer>
         {newAddressError && (
           <Note title={"Error"} noteType={NoteType.error}>
             {newAddressError}
           </Note>
         )}
         {!newAddressError && <StatusIcon state={response?.address ? "success" : "processing"} />}
         <DetailsContainer>
           <DetailsRowLinkAndCopy label={"Address:"} copy={response?.address}>
             {response?.address}
           </DetailsRowLinkAndCopy>
         </DetailsContainer>
         <ButtonWrapper
           className={styles.customButtonWrapperStyles}
           rightChildren={
             <Button
               text={t.newAddress}
               onClick={() => {
                setCreateTagState(ProgressStepState.active);
                 setDoneState(ProgressStepState.disabled);
                 setStepIndex(0);
                 setResponse(undefined);
               }}
               buttonColor={buttonColor.subtle}
             />
           }
         />
       </ProgressTabContainer>
      </ProgressTabs>
    </PopoutPageTemplate>
  );
}

export default ModifyTagModal;
