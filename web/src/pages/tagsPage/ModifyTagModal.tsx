import {
  ArrowSyncFilled as ProcessingIcon,
  CheckmarkRegular as SuccessIcon,
  DismissRegular as FailedIcon,
  Tag24Regular as TagHeaderIcon,
  Note20Regular as NoteIcon,
} from "@fluentui/react-icons";
import { WS_URL } from "apiSlice";
import Button, { buttonColor, ButtonWrapper } from "components/buttons/Button";
import ProgressHeader, { ProgressStepState, Step } from "features/progressTabs/ProgressHeader";
import ProgressTabs, { ProgressTabContainer } from "features/progressTabs/ProgressTab";
import PopoutPageTemplate from "features/templates/popoutPageTemplate/PopoutPageTemplate";
import { ChangeEvent, useState, useEffect } from "react";
import { useNavigate } from "react-router";
import useWebSocket from "react-use-websocket";
import styles from "./modifyTagModal.module.scss";
import useTranslations from "services/i18n/useTranslations";
import Select, { SelectOptions } from "features/forms/Select";
import { ActionMeta } from "react-select";
import clone from "clone";
// import { nodeConfiguration } from "apiTypes";
// import Note, { NoteType } from "features/note/Note";
// import {
//   DetailsContainer,
//   DetailsRowLinkAndCopy,
// } from "features/templates/popoutPageTemplate/popoutDetails/PopoutDetails";
import classNames from "classnames";
import { NewAddressResponse, NewAddressError, AddressType } from "features/transact/newAddress/NewAddressModal"
import FormRow from "features/forms/FormWrappers";
import Input from "components/forms/input/Input";
import { useGetCategoriesQuery } from "pages/categoriesPage/categoriesApi";
import { Category } from "pages/categoriesPage/categoriesTypes";
import { useGetNodesChannelsQuery, useAddTagMutation, useAddChannelsGroupsMutation } from "./tagsApi";
import { ChannelNode, ChannelForTag, NodeForTag } from "./tagsTypes"


const updateStatusClass = {
  IN_FLIGHT: styles.inFlight,
  FAILED: styles.failed,
  SUCCEEDED: styles.success,
};

const updateStatusIcon = {
  IN_FLIGHT: <ProcessingIcon />,
  FAILED: <FailedIcon />,
  SUCCEEDED: <SuccessIcon />,
  NOTE: <NoteIcon />,
};


function ModifyTagModal() {
  const { t } = useTranslations();

  const [createTageState, setCreateTagState] = useState(ProgressStepState.active);
  const [doneState, setDoneState] = useState(ProgressStepState.disabled);
  const [errMessage, setErrorMessage] = useState<any[]>([]);
  const [stepIndex, setStepIndex] = useState(0);
  const [response, setResponse] = useState<NewAddressResponse>();
  const [newAddressError, setNewAddressError] = useState("");
  const [tagName, setTagName] = useState<string>("");
  const [tagId, setTagId] = useState<number>(0);
  const [nodeId, setNodeId] = useState<number>(0);
  const [channelId, setChannelId] = useState<number>(0);
  const tagColorOptions: SelectOptions[] = [
    { value: "", label: "Select your color" },
    { value: "Yellow", label: "Yellow" },
    { value: "Red", label: "Red" },
    { value: "Orange", label: "Orange" },
  ];

  const [selectedTagColor, setTagColor] = useState<string>(tagColorOptions[0].value as string);

  const { data: categoriesResponse }  = useGetCategoriesQuery<{
    data: Array<Category>;
    isLoading: boolean;
    isFetching: boolean;
    isUninitialized: boolean;
    isSuccess: boolean;
  }>();
  let tagCategorieOptions: SelectOptions[] = [{ value: 0, label: "Select your Category"}];
  if (categoriesResponse !== undefined) {
    tagCategorieOptions = categoriesResponse.map((tagCategorieOptions: Category) => {
      return { value: tagCategorieOptions.categoryId, label: tagCategorieOptions.name};
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
  let channelsNodesOptions: SelectOptions[] = [{ value: 0, label: "Channel or Node"}];

  const [selectedTarget, setTarget] = useState<number>(channelsNodesOptions[0].value as number);
  const [selectedType, setType] = useState<string>("");
  let channelsOptions: SelectOptions[] = [];
  let nodesOptions: SelectOptions[] = [];
  if (channelsNodesResponse !== undefined) {
    channelsOptions = channelsNodesResponse.nodes.map((nodesOptions: NodeForTag) => {
      return { value: nodesOptions.nodeId, label: nodesOptions.alias, type: 'node' };
    });
    nodesOptions = channelsNodesResponse.channels.map((channelsOptions: ChannelForTag) => {
      return { value: channelsOptions.channelId, label: channelsOptions.shortChannelId, type: 'channel' };
    });
  }
  channelsNodesOptions = nodesOptions.concat(channelsOptions)

  const [addTagMutation, addTagResponse] = useAddTagMutation();
  const [addChannelsGroupsMutation, addChannelsGroupsResponse] = useAddChannelsGroupsMutation();


  useEffect(() => {
    let tag
    const message = clone(errMessage) || [];
    if (addTagResponse.isSuccess) {
      tag = addTagResponse.data.tagId ? addTagResponse.data.tagId : 0;
      setTagId(tag)
      if (addTagResponse.status != 'fulfilled') {
        setDoneState(ProgressStepState.error);
        message.push(
          <span key={0} className={classNames(styles.modifyTagsStatusMessage)}>
            {"Error saving the tag"}
          </span>
        );
        setErrorMessage(message)
      } else {
        addChannelsGroupsMutation({
          tagId: tag,
          nodeId,
          channelId,
          categoryId: selectedTagCategory

        })
        setDoneState(ProgressStepState.completed);
      }
     }
    if (addTagResponse.isError) {
      setDoneState(ProgressStepState.error);
      message.push(
        <span key={0} className={classNames(styles.modifyTagsStatusMessage)}>
          {"Error saving the tag"}
        </span>
      );
      setErrorMessage(message)

    }
  }, [addTagResponse]);

  function handleTarget(value: number, type: string) {
    setTarget(value);
    setType(type);
    let selectedChannelNode;
    if (type == 'channel') {
      selectedChannelNode = channelsNodesResponse.channels.find((channelsOptions: ChannelForTag) => {
        if (channelsOptions.channelId == value) {
          return channelsOptions
        }
      });
      setChannelId(value)
      setNodeId(selectedChannelNode?.nodeId as number | 0)
    } else {
      setNodeId(value)
    }
  }


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
                    setTagColor(selectOptions?.value as string);
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
                  label={t.tagTarget}
                  onChange={(newValue: unknown, _: ActionMeta<unknown>) => {
                    const selectOptions = newValue as SelectOptions;
                    handleTarget(selectOptions?.value as number, selectOptions?.type as string);
                  }}
                  options={channelsNodesOptions}
                  value={channelsNodesOptions.find((option) => option.value === selectedTarget && option.type == selectedType)}
                />
              </div>
            </FormRow>
            <ButtonWrapper
              className={styles.customButtonWrapperStyles}
              rightChildren={
                <Button
                  text={t.newTag}
                  onClick={() => {
                    setStepIndex(1);
                    setCreateTagState(ProgressStepState.completed);
                    setDoneState(ProgressStepState.active);
                    addTagMutation({
                      style: selectedTagColor,
                      name: tagName,
                      categoryId: selectedTagCategory
                    });
                  }}
                  buttonColor={buttonColor.subtle}
                />
              }
            />

          </div>
       </ProgressTabContainer>

       <ProgressTabContainer>
          <div
            className={classNames(
              styles.modifyTagsResultIconWrapper,
              { [styles.failed]: !addTagResponse.data },
              updateStatusClass[addTagResponse?.status == "fulfilled" ? "SUCCEEDED" : "FAILED"]
            )}
          >
            {" "}
            {updateStatusIcon[addTagResponse.status == "fulfilled" ? "SUCCEEDED" : "FAILED"]}
          </div>
          <div className={errMessage.length ? styles.errorBox : styles.successeBox}>
            <div>
              <div className={errMessage.length ? styles.errorIcon : styles.successIcon}>
                {updateStatusIcon["NOTE"]}
              </div>
              <div className={errMessage.length ? styles.errorNote : styles.successNote}>
                {errMessage.length ? t.openCloseChannel.error : t.openCloseChannel.note}
              </div>
            </div>
            <div className={errMessage.length ? styles.errorMessage : styles.successMessage}>
              {errMessage.length ? errMessage : t.tagsModal.addConfirmedMessage}
            </div>
          </div>
         <ButtonWrapper
           className={styles.customButtonWrapperStyles}
           rightChildren={
             <Button
               text={t.newTag}
               onClick={() => {
                setCreateTagState(ProgressStepState.active);
                setDoneState(ProgressStepState.disabled);
                setStepIndex(0);
                setErrorMessage([]);
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
