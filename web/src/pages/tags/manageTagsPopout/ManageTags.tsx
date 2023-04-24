import mixpanel from "mixpanel-browser";
import { Tag24Regular as TagHeaderIcon } from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import PopoutPageTemplate from "features/templates/popoutPageTemplate/PopoutPageTemplate";
import { useState } from "react";
import { useNavigate } from "react-router";
import {
  useGetTagsQuery,
  useTagChannelMutation,
  useTagNodeMutation,
  useUntagChannelMutation,
  useUntagNodeMutation,
} from "../tagsApi";
import { Select } from "components/forms/forms";
import { useParams } from "react-router-dom";
import { SelectOptionType } from "components/forms/select/Select";
import { TagSize, TagColor } from "components/tags/Tag";
import RemoveTag from "components/tags/RemoveTag";
import Button, { ColorVariant, ButtonPosition } from "components/buttons/Button";
import { AddSquare20Regular as AddIcon } from "@fluentui/react-icons";
import styles from "./manageTags.module.scss";

function isOption(result: unknown): result is SelectOptionType {
  return result !== null && typeof result === "object" && "value" in result && "label" in result;
}

export default function MangeTagsPopout() {
  const { t } = useTranslations();
  const navigate = useNavigate();
  const tagsResponse = useGetTagsQuery();
  const { channelId: channelIdParam } = useParams<{ channelId: string }>();
  const channelId = channelIdParam ? parseInt(channelIdParam) : undefined;
  const { nodeId: nodeIdParam } = useParams<{ nodeId: string }>();
  const nodeId = nodeIdParam ? parseInt(nodeIdParam) : undefined;
  const [tagNode] = useTagNodeMutation();
  const [tagChannel] = useTagChannelMutation();
  const [untagChannel] = useUntagChannelMutation();
  const [untagNode] = useUntagNodeMutation();

  const [selectedTagState, setSelectedTagState] = useState("");

  const existingChannelTags = tagsResponse.data?.filter((t) => {
    if (!t.channels) return false;
    return t.channels.some((c) => c.channelId === channelId);
  });
  const existingNodeTags = tagsResponse.data?.filter((t) => {
    if (!t.nodes) return false;
    return t.nodes.some((n) => n.nodeId === nodeId);
  });

  // Loops through all tags and create options for the select component
  const tagOptions = tagsResponse.data?.map((tag) => ({
    value: tag.tagId,
    label: tag.name,
  }));

  const closeAndReset = () => {
    navigate(-1);
  };

  function handleTagSelect(tagOption: unknown) {
    if (!isOption(tagOption)) {
      return;
    }
    setSelectedTagState(tagOption.value);
  }

  function addTag(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedTagState) {
      return;
    }
    if (channelId !== undefined) {
      tagChannel({
        tagId: parseInt(selectedTagState),
        channelId: channelId,
      });
    }
    if (nodeId !== undefined) {
      tagNode({
        tagId: parseInt(selectedTagState),
        nodeId: nodeId,
      });
    }
  }

  const title = channelId ? t.channel : t.node;

  const removeTag = (tagId: number, tagName: string) => {
    if (channelId !== undefined) {
      mixpanel.track("Untag", {
        tagId: tagId,
        tagName: tagName,
        channelId: channelId,
        tagType: "channel",
      });
      untagChannel({ tagId: tagId, channelId: channelId });
    }
    if (nodeId !== undefined) {
      mixpanel.track("Untag", {
        tagId: tagId,
        tagName: tagName,
        nodeId: nodeId,
        tagType: "node",
      });
      untagNode({ tagId: tagId, nodeId: nodeId });
    }
  };

  return (
    <PopoutPageTemplate title={`${t.tag} ${title}`} show={true} icon={<TagHeaderIcon />} onClose={closeAndReset}>
      <div>
        {existingChannelTags && existingChannelTags.length > 0 && (
          <>
            <h4>Remove existing tag</h4>
            <div className={styles.tagWrapper}>
              {existingChannelTags.map((ct) => (
                <RemoveTag
                  key={"channel" + ct.tagId}
                  label={ct.name}
                  colorVariant={TagColor.primary}
                  sizeVariant={TagSize.normal}
                  onClick={() => removeTag(ct.tagId ?? 0, ct.name)}
                />
              ))}
            </div>
          </>
        )}
      </div>
      <div>
        {existingNodeTags && existingNodeTags.length > 0 && (
          <>
            <h4>Remove existing tag</h4>

            <div className={styles.tagWrapper}>
              {existingNodeTags.map((nt) => (
                <RemoveTag
                  key={"node" + nt.tagId}
                  label={nt.name}
                  colorVariant={TagColor.primary}
                  sizeVariant={TagSize.normal}
                  onClick={() => removeTag(nt.tagId ?? 0, nt.name)}
                />
              ))}
            </div>
          </>
        )}
      </div>
      <div className={styles.addTagWrapper}>
        <h4>Add new tag</h4>
        <form onSubmit={addTag}>
          <Select
            intercomTarget={"add-tag-select-tag"}
            label={t.tag}
            autoFocus={true}
            options={tagOptions}
            onChange={handleTagSelect}
            placeholder={"Select Tag"}
          />
          <Button
            type={"submit"}
            icon={<AddIcon />}
            buttonColor={ColorVariant.success}
            buttonPosition={ButtonPosition.left}
          >
            Add Tag
          </Button>
        </form>
      </div>
    </PopoutPageTemplate>
  );
}
