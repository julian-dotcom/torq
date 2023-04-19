import { Tag24Regular as TagHeaderIcon } from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import PopoutPageTemplate from "features/templates/popoutPageTemplate/PopoutPageTemplate";
import { useState } from "react";
import { useNavigate } from "react-router";
import { useGetTagsQuery, useTagChannelMutation, useTagNodeMutation } from "../tagsApi";
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

export default function AddTagModal() {
  const { t } = useTranslations();
  const navigate = useNavigate();
  const tagsResponse = useGetTagsQuery();
  const { channelId } = useParams<{ channelId: string }>();
  const { nodeId } = useParams<{ nodeId: string }>();
  const [tagNode] = useTagNodeMutation();
  const [tagChannel] = useTagChannelMutation();

  const [selectedTagState, setSelectedTagState] = useState("");

  const existingChannelTags = tagsResponse.data?.filter((t) => {
    if (!t.channels) return false;
    return t.channels.some((c) => c.channelId === parseInt(channelId ?? "0"));
  });
  const existingNodeTags = tagsResponse.data?.filter((t) => {
    if (!t.nodes) return false;
    return t.nodes.some((n) => n.nodeId === parseInt(nodeId ?? "0"));
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
        channelId: parseInt(channelId),
      });
    }
    if (nodeId !== undefined) {
      tagNode({
        tagId: parseInt(selectedTagState),
        nodeId: parseInt(nodeId),
      });
    }
  }

  const title = channelId ? t.channel : t.node;

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
                />
              ))}
            </div>
          </>
        )}
      </div>
      <div>
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
