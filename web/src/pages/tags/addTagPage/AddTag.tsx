import { Tag24Regular as TagHeaderIcon } from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import PopoutPageTemplate from "features/templates/popoutPageTemplate/PopoutPageTemplate";
import { useNavigate } from "react-router";
import { useGetTagsQuery, useTagChannelMutation, useTagNodeMutation } from "../tagsApi";
import { Select } from "components/forms/forms";
import { useParams } from "react-router-dom";
import { SelectOptionType } from "components/forms/select/Select";

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

  // Loops through all tags and create options for the select component
  const tagOptions = tagsResponse.data?.map((tag) => ({
    value: tag.tagId,
    label: tag.name,
  }));

  const closeAndReset = () => {
    navigate(-1);
  };

  function handleTagSelect(newValue: unknown) {
    if (!isOption(newValue)) {
      return;
    }
    if (channelId !== undefined) {
      tagChannel({
        tagId: parseInt(newValue.value),
        channelId: parseInt(channelId),
      });
    }
    if (nodeId !== undefined) {
      tagNode({
        tagId: parseInt(newValue.value),
        nodeId: parseInt(nodeId),
      });
    }
  }

  const title = channelId ? t.channel : t.node;

  return (
    <PopoutPageTemplate title={`${t.tag} ${title}`} show={true} icon={<TagHeaderIcon />} onClose={closeAndReset}>
      <div>
        <Select
          intercomTarget={"add-tag-select-tag"}
          label={t.tag}
          autoFocus={true}
          options={tagOptions}
          onChange={handleTagSelect}
          placeholder={"Select Tag"}
        />
      </div>
    </PopoutPageTemplate>
  );
}
