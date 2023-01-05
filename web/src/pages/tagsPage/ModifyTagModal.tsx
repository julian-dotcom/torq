import {
  Tag24Regular as TagHeaderIcon,
  ChevronDown20Regular as CollapsedIcon,
  LineHorizontal120Regular as ExpandedIcon,
  ArrowRouting20Regular as ChannelsIcon,
  Molecule20Regular as NodesIcon,
  TargetArrow20Regular as TargetIcon,
  Dismiss20Regular as DeleteIcon,
} from "@fluentui/react-icons";
import Button, { ColorVariant, ButtonWrapper } from "components/buttons/Button";
import ProgressTabs from "features/progressTabs/ProgressTab";
import PopoutPageTemplate from "features/templates/popoutPageTemplate/PopoutPageTemplate";
import { ChangeEvent, useState, useEffect, ReactNode } from "react";
import { useNavigate } from "react-router";
import styles from "./modifyTagModal.module.scss";
import useTranslations from "services/i18n/useTranslations";
import Select, { SelectOptions } from "features/forms/Select";
import { ActionMeta } from "react-select";
import clone from "clone";
import classNames from "classnames";
import Collapse from "features/collapse/Collapse";
import FormRow from "features/forms/FormWrappers";
import Input from "components/forms/input/Input";
import { useGetCategoriesQuery } from "pages/categoriesPage/categoriesApi";
import { Category } from "pages/categoriesPage/categoriesTypes";
import {
  useGetNodesChannelsQuery,
  useAddTagMutation,
  useAddChannelsGroupsMutation,
  useGetTagQuery,
  useSetTagMutation,
  useGetCorridorByReferenceQuery,
  useDeleteChannelGroupByTagMutation
} from "./tagsApi";
import { ChannelNode, ChannelForTag, NodeForTag, Tag, Corridor, CorridorFields, ChannelGroup } from "./tagsTypes"
import TextCell from "components/table/cells/text/TextCell"
import { TagColor } from "components/tags/Tag"

function ModifyTagModal() {
  const { t } = useTranslations();

  const [errMessage, setErrorMessage] = useState<ReactNode[]>([]);
  const [stepIndex, setStepIndex] = useState(0);
  const [tagName, setTagName] = useState<string>("");
  const [tagId, setTagId] = useState<number>(0);
  const [nodeId, setNodeId] = useState<number>(0);
  const [channelId, setChannelId] = useState<number>(0);
  const [totalNodes, setTotalNodes] = useState<number>(0);
  const [totalChannels, setTotalChannels] = useState<number>(0);
  const [modalTitle, setModalTitle] = useState<string>(t.tagsModal.createTag);
  const [modalButton, setModalButton] = useState<string>(t.tagsModal.create);
  const [modalUpdateMode, setModalUpdateMode] = useState<boolean>(false);
  const [collapsedNodeState, setCollapsedNodeState] = useState<boolean>(true);
  const [collapsedChannelState, setCollapsedChannelState] = useState<boolean>(true);
  const [targetNodes, setTargetNodes] = useState<ReactNode[]>([]);
  const [targetChannels, setTargetChannels] = useState<ReactNode[]>([]);
  const [colorName, setColorName] = useState<string>("N/A");
  const tagColorOptions: SelectOptions[] = [
    { value: "", label: "Select your color" },
    { value: TagColor.primary, label: TagColor.primary.charAt(0).toUpperCase() + TagColor.primary.slice(1) },
    { value: TagColor.success, label: TagColor.success.charAt(0).toUpperCase() + TagColor.success.slice(1) },
    { value: TagColor.warning, label: TagColor.warning.charAt(0).toUpperCase() + TagColor.warning.slice(1) },
    { value: TagColor.error, label: TagColor.error.charAt(0).toUpperCase() + TagColor.error.slice(1) },
    { value: TagColor.accent1, label: TagColor.accent1.charAt(0).toUpperCase() + TagColor.accent1.slice(1) },
    { value: TagColor.accent3, label: TagColor.accent3.charAt(0).toUpperCase() + TagColor.accent3.slice(1) },
    { value: TagColor.custom, label: TagColor.custom.charAt(0).toUpperCase() + TagColor.custom.slice(1) },
  ];

  const queryParameters = new URLSearchParams(window.location.search)
  const urlTagId = queryParameters.get("tagId");

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
  let categoryName = "N/A";
  const { data: channelsNodesResponse } = useGetNodesChannelsQuery<{
    data: ChannelNode;
    isLoading: boolean;
    isFetching: boolean;
    isUninitialized: boolean;
    isSuccess: boolean;
  }>();

  const { data: tagResponse } = useGetTagQuery<{
    data: Tag;
    isLoading: boolean;
    isFetching: boolean;
    isUninitialized: boolean;
    isSuccess: boolean;
  }>(tagId);

    const { data: corridorsResponse } = useGetCorridorByReferenceQuery<{
      data: Corridor;
      isLoading: boolean;
      isFetching: boolean;
      isUninitialized: boolean;
      isSuccess: boolean;
    }>(tagId);

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
  channelsNodesOptions = nodesOptions.concat(channelsOptions);

  const [addTagMutation, addTagResponse] = useAddTagMutation();
  const [setTagMutation, setTagResponse] = useSetTagMutation();

  const [addChannelsGroupsMutation, addChannelsGroupsResponse] = useAddChannelsGroupsMutation();
  const [deleteChannelGroupByTagMutation, deleteChannelGroupByTagResponse] = useDeleteChannelGroupByTagMutation();

  function handleDeleteTarget(corridorId: number) {
    deleteChannelGroupByTagMutation(corridorId)
  }

  useEffect(() => {
    let tag = 0
    const message = clone(errMessage) || [];

    if (urlTagId) {
      setTagId(Number(urlTagId));
      setModalTitle(t.tagsModal.updateTag);
      setModalButton(t.tagsModal.update);
      setModalUpdateMode(true);
    }

    if (tagResponse) {
      setTagName(tagResponse.name);
      setTagColor(tagResponse.style);
      setTagCategory(tagResponse?.categoryId || 0);

      categoriesResponse?.forEach((tagCategorieOptions: Category) => {
        if (tagResponse?.categoryId == tagCategorieOptions.categoryId) {
          categoryName = tagCategorieOptions.name;
        }
      });
      tagColorOptions.forEach((colors) => {
        if (tagResponse.style == colors.value as string) {
          setColorName(colors?.label || "N/A")
        }
      });
    }

    if (corridorsResponse) {
      setTotalNodes(corridorsResponse.totalNodes);
      setTotalChannels(corridorsResponse.totalChannels);
      const listNodes: ReactNode[] = [];
      const listChannels: ReactNode[] = [];
      corridorsResponse.corridors?.map((c: CorridorFields) => {
        if (c.shortChannelId) {
          listChannels.push(
            <FormRow className={styles.targetRows}>
              <div className={styles.openChannelTableRow} >
                <div className={styles.label}>
                  <TextCell
                    className={classNames(styles.simple, styles.colapsedLabels, styles.textLeft)}
                    current={c.shortChannelId}
                    key={c.shortChannelId}
                  />
                  <div className={classNames(styles.deleteButtons)} onClick={() => {handleDeleteTarget(c.corridorId)}}>
                    <DeleteIcon/>
                  </div>
                </div>
              </div>
            </FormRow>
          );
        } else {
          listNodes.push(
            <FormRow className={styles.targetRows}>
              <div className={styles.openChannelTableRow} >
                <div className={styles.label}>
                  <TextCell
                    className={classNames(styles.simple, styles.colapsedLabels, styles.textLeft)}
                    current={c.alias}
                    key={c.alias}
                  />
                  <div className={classNames(styles.deleteButtons)} onClick={() => {handleDeleteTarget(c.corridorId)}}>
                    <DeleteIcon/>
                  </div>
                </div>
              </div>
            </FormRow>
          );
        }
      })
      setTargetNodes(listNodes);
      setTargetChannels(listChannels);
    }

    if ((addTagResponse.isSuccess || setTagResponse.isSuccess) && !addChannelsGroupsResponse?.data && addChannelsGroupsResponse.status != "pending") {
      if (addTagResponse.isSuccess && !addTagResponse.isLoading) {
        tag = addTagResponse.data.tagId ? addTagResponse.data.tagId : 0;
      }
      if (setTagResponse.isSuccess && !setTagResponse.isLoading) {
        tag = tagId;
      }
      setTagId(tag);
      if (addTagResponse.isSuccess && addTagResponse.status != 'fulfilled' ||
      setTagResponse.isSuccess && setTagResponse.status != 'fulfilled') {
        message.push(
          <span key={0} className={classNames(styles.modifyTagsStatusMessage)}>
            {"Error saving the tag"}
          </span>
        );
        setErrorMessage(message)
      }

      if (addTagResponse.isSuccess && addTagResponse.status == 'fulfilled' ||
      setTagResponse.isSuccess && setTagResponse.status == 'fulfilled') {
        if (nodeId && nodeId > 0 ) {
          const channelGoupObj: ChannelGroup = {
            tagId: tag,
            nodeId,
          }
          if (channelId && channelId > 0) {
            channelGoupObj.channelId = channelId
          }
          if (selectedTagCategory > 0) {
            channelGoupObj.categoryId = selectedTagCategory
          }
          addChannelsGroupsMutation(channelGoupObj);
        }
      }
     }
    if (addTagResponse.isError || setTagResponse.isError) {
      message.push(
        <span key={0} className={classNames(styles.modifyTagsStatusMessage)}>
          {"Error saving the tag"}
        </span>
      );
      setErrorMessage(message);

    }

  }, [addTagResponse, setTagResponse, tagResponse, corridorsResponse, deleteChannelGroupByTagResponse]);

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

  const handleCollapseNodeClick = () => {
    setCollapsedNodeState(!collapsedNodeState);
  };
  const handleCollapseChannelClick = () => {
    setCollapsedChannelState(!collapsedChannelState);
  };

  const closeAndReset = () => {
    setStepIndex(0);
    setTagName("");
    setTagColor(tagColorOptions[0].value as string);
    setTagCategory(tagCategorieOptions[0].value as number);
    setTarget(channelsNodesOptions[0].value as number);
    navigate(-1);
  };

  const navigate = useNavigate();

  return (
    <PopoutPageTemplate title={modalTitle} show={true} onClose={() => closeAndReset()} icon={<TagHeaderIcon />}>

      <ProgressTabs showTabIndex={stepIndex}>
          <div className={styles.activeColumns}>
            <FormRow>
              <div className={styles.updateChannelTableDouble}>
                <div className={styles.input}>
                  <Input
                    disabled={tagId && tagId < 0 ? true : false}
                    label={"Name"}
                    className={classNames(styles.simple, { [ styles.lockedInput]: tagId < 0 })}
                    value={tagName}
                    onChange={(e: ChangeEvent<HTMLInputElement>) => {
                      setTagName(e.target.value);
                    }}
                  />
                </div>
              </div>
            </FormRow>
            <FormRow>
                {tagId >= 0 && (
                  <div className={styles.updateChannelTableDouble}>
                    <Select
                      label={t.tagsModal.color}
                      onChange={(newValue: unknown, _: ActionMeta<unknown>) => {
                        const selectOptions = newValue as SelectOptions;
                        setTagColor(selectOptions?.value as string);
                      }}
                      options={tagColorOptions}
                      value={tagColorOptions.find((option) => option.value === selectedTagColor)}
                    />
                  </div>
                )}
                {tagId < 0 && (
                <div className={styles.updateChannelTableDouble}>
                  <div className={styles.input}>
                    <Input
                      disabled={tagId && tagId < 0 ? true : false}
                      label={t.tagsModal.color}
                      className={classNames(styles.simple, styles.lockedInput)}
                      value={colorName}
                    />
                  </div>
                </div>
                )}
                {tagId >= 0 && (
                  <div className={styles.updateChannelTableDouble}>
                    <Select
                      isDisabled={tagId && tagId < 0 ? true : false}
                      label={t.tagsModal.category}
                      onChange={(newValue: unknown, _: ActionMeta<unknown>) => {
                        const selectOptions = newValue as SelectOptions;
                        setTagCategory(selectOptions?.value as number);
                      }}
                      options={tagCategorieOptions}
                      value={tagCategorieOptions.find((option) => option.value === selectedTagCategory)}
                    />
                </div>
                )}
                {tagId < 0 && (
                <div className={styles.updateChannelTableDouble}>
                  <div className={styles.input}>
                    <Input
                      disabled={tagId && tagId < 0 ? true : false}
                      label={t.tagsModal.category}
                      className={classNames(styles.simple, styles.lockedInput)}
                      value={categoryName}
                    />
                  </div>
                </div>
                )}
            </FormRow>
            <FormRow>
            <div className={styles.openChannelTableRow}>
                <Select
                  label={t.tagsModal.target}
                  onChange={(newValue: unknown, _: ActionMeta<unknown>) => {
                    const selectOptions = newValue as SelectOptions;
                    handleTarget(selectOptions?.value as number, selectOptions?.type as string);
                  }}
                  options={channelsNodesOptions}
                  value={channelsNodesOptions.find((option) => option.value === selectedTarget && option.type == selectedType)}
                />
              </div>
            </FormRow>

            {tagId >= 0 && (
            <ButtonWrapper
            className={styles.customButtonWrapperStyles}
            rightChildren={
              <Button
                className={styles.tagbutton}
                onClick={() => {
                  setStepIndex(0);
                  const tagObj: Tag = {
                    style: selectedTagColor,
                    name: tagName,
                  }
                  if (selectedTagCategory != 0) {
                    tagObj.categoryId = selectedTagCategory
                  }
                  if (modalUpdateMode) {
                    tagObj.tagId = tagId
                    setTagMutation(tagObj);
                  } else {
                    addTagMutation(tagObj);
                  }
                }}
                buttonColor={ColorVariant.primary}>
                {modalButton}
              </Button>
            }
          />
            )}

            {modalUpdateMode && (targetNodes.length > 0 || targetChannels.length > 0) && (
              <div className={styles.target}><TargetIcon className={styles.targetIcon}/> <span  className={styles.targetText}>Applied to</span></div>
            )}

            <div className={styles.collapsGroup}>
              {modalUpdateMode && targetNodes.length > 0 && (
                <div
                  className={classNames(styles.header, { [styles.expanded]: !collapsedNodeState })}
                  onClick={handleCollapseNodeClick}
                >
                <div className={styles.title}><NodesIcon />{`${totalNodes} Node(s)`}</div>
                  <div className={classNames(styles.collapseIcon, { [styles.collapsed]: collapsedNodeState })}>
                    {collapsedNodeState ? <CollapsedIcon /> : <ExpandedIcon />}
                  </div>
                </div>
              )}
              <Collapse collapsed={collapsedNodeState} animate={modalUpdateMode}>
                <div>
                  {targetNodes}
                </div>
              </Collapse>
            </div>
            <div className={styles.collapsGroup}>
            {modalUpdateMode && targetChannels.length > 0 && (
              <div
                className={classNames(styles.header, { [styles.expanded]: !collapsedChannelState })}
                onClick={handleCollapseChannelClick}
              >
              <div className={styles.title}> <ChannelsIcon />{`${totalChannels} Channel(s)`}</div>
              <div className={classNames(styles.collapseIcon, { [styles.collapsed]: collapsedChannelState })}>
                {collapsedChannelState ? <CollapsedIcon /> : <ExpandedIcon />}
              </div>
              </div>
            )}
            <Collapse collapsed={collapsedChannelState} animate={modalUpdateMode}>
              <div>
                {targetChannels}
              </div>
            </Collapse>
            </div>
            {tagId < 0 && (
              <ButtonWrapper
                className={styles.customButtonWrapperStyles}
                rightChildren={
                  <Button
                    className={styles.tagbutton}
                    onClick={() => {
                      setStepIndex(0);
                      const tagObj: Tag = {
                        style: selectedTagColor,
                        name: tagName,
                      }
                      if (selectedTagCategory != 0) {
                        tagObj.categoryId = selectedTagCategory
                      }
                      if (modalUpdateMode) {
                        tagObj.tagId = tagId
                        setTagMutation(tagObj);
                      } else {
                        addTagMutation(tagObj);
                      }
                    }}
                  buttonColor={ColorVariant.primary}>
                  {modalButton}
                </Button>
                }
              />
            )}
          </div>

      </ProgressTabs>
    </PopoutPageTemplate>
  );
}

export default ModifyTagModal;
