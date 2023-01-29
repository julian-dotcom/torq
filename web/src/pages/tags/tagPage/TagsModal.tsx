import { Tag24Regular as TagHeaderIcon } from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import { useParams } from "react-router-dom";
import { TagColor } from "components/tags/Tag";
import { useGetTagQuery, useAddTagMutation, useSetTagMutation } from "../tagsApi";
import { ExpandedTag, TagResponse } from "../tagsTypes";
import PopoutPageTemplate from "features/templates/popoutPageTemplate/PopoutPageTemplate";
import { Form, Select, Input, InputRow } from "components/forms/forms";
import styles from "./modifyTagModal.module.scss";
import classNames from "classnames";
import { useGetCategoriesQuery } from "pages/categoriesPage/categoriesApi";
import Button, { ButtonWrapper, ColorVariant } from "components/buttons/Button";
import { useNavigate } from "react-router";
import { useEffect, useState } from "react";
import TargetsSection from "features/targetsSection/TargetsSection";
import mixpanel from "mixpanel-browser";

const inputNames = {
  name: "name",
  style: "style",
  categoryId: "categoryId",
};

export default function TagsModal() {
  const { t } = useTranslations();
  const [addTagMutation] = useAddTagMutation();
  const [setTagMutation] = useSetTagMutation();
  const [isLocked, setIsLocked] = useState(true);
  const { tag } = useGetOrCreateTag();
  const [tagColor, setTagColor] = useState(tag.style);
  const [categoryId, setCategoryId] = useState(tag.categoryId);
  useEffect(() => {
    setIsLocked((tag?.tagId || 0) <= -1);
    if (tag?.tagId) {
      setTagColor(tag.style);
      setCategoryId(tag.categoryId);
    }
  }, [tag]);

  const navigate = useNavigate();

  const { categoryOptions } = useCategoryOptions(tag.categoryId || 0);
  const { tagColorOptions } = useTagColorOptions(tag.style);

  const closeAndReset = () => {
    navigate(-1);
  };

  function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const data = new Map(new FormData(event.currentTarget).entries());
    const categoryValue = data.get(inputNames.categoryId);
    const category = categoryValue ? parseInt(categoryValue.toString()) : undefined;

    const tagRequest: ExpandedTag = {
      ...tag,
      name: data.get("name") as string,
      style: tagColor,
      categoryId: category,
    };

    if (tag.tagId) {
      setTagMutation(tagRequest);
      mixpanel.track("Update Tag", {
        tagId: tag.tagId,
        tagName: tag.name,
        tagStyle: tag.style,
        categoryId: tag.categoryId,
        categoryName: tag.categoryName,
        categoryStyle: tag.categoryStyle,
      });
    } else {
      addTagMutation(tagRequest).then((result) => {
        if (hasTagsData(result) && result.data?.tagId !== undefined) {
          navigate(-1);
        }
        mixpanel.track("Create Tag", tagRequest);
      });
    }
  }

  return (
    <PopoutPageTemplate title={t.tags} show={true} icon={<TagHeaderIcon />} onClose={closeAndReset}>
      <div className={styles.activeColumns}>
        <Form onSubmit={handleSubmit} name={"tagForm"}>
          <div className={styles.updateChannelTableDouble}>
            <div className={styles.input}>
              <Input
                autoFocus={true}
                disabled={isLocked}
                label={"Name"}
                name={inputNames.name}
                className={classNames(styles.simple)}
                defaultValue={tag.name}
              />
            </div>
          </div>
          <InputRow>
            <Select
              key={"input-row-item-a"}
              label={t.tagsModal.color}
              name={inputNames.style}
              isDisabled={isLocked}
              options={tagColorOptions}
              value={tagColorOptions.find((option) => option.value === tagColor)}
              onChange={(newValue) => {
                if (isColorValue(newValue)) {
                  setTagColor(newValue.value);
                }
              }}
            />
            <Select
              key={"input-row-item-b"}
              isDisabled={isLocked}
              placeholder={t.noCategory}
              name={inputNames.categoryId}
              label={t.tagsModal.category}
              options={categoryOptions}
              value={categoryOptions.find((option) => option.value === categoryId)}
              onChange={(newValue) => {
                if (isCateogryValue(newValue)) {
                  setCategoryId(newValue.value);
                }
              }}
            />
          </InputRow>

          <ButtonWrapper
            className={styles.customButtonWrapperStyles}
            rightChildren={
              <Button disabled={isLocked} type={"submit"} buttonColor={ColorVariant.primary}>
                {tag.tagId ? t.tagsModal.update : t.tagsModal.create}
              </Button>
            }
          />
        </Form>

        {tag.tagId && <TargetsSection tagId={tag.tagId} tagName={tag.name} channels={tag.channels} nodes={tag.nodes} />}
      </div>
    </PopoutPageTemplate>
  );
}

function useGetOrCreateTag() {
  const { tagId } = useParams<{ tagId: string }>();

  const newTag: TagResponse = {
    name: "",
    style: TagColor.primary,
    channels: [],
    nodes: [],
  };

  const { data, isLoading } = tagId
    ? useGetTagQuery(parseInt(tagId), { skip: isNaN(parseInt(tagId)) })
    : { data: newTag, isLoading: false };

  return {
    tag: data ? data : newTag,
    tagIsLoading: isLoading,
  };
}

function useTagColorOptions(tag: string) {
  const { t } = useTranslations();
  const tagColorOptions = [
    { value: TagColor.primary, label: t.tagsModal.tagColors[TagColor.primary] },
    { value: TagColor.success, label: t.tagsModal.tagColors[TagColor.success] },
    { value: TagColor.warning, label: t.tagsModal.tagColors[TagColor.warning] },
    { value: TagColor.accent1, label: t.tagsModal.tagColors[TagColor.accent1] },
    { value: TagColor.accent2, label: t.tagsModal.tagColors[TagColor.accent2] },
    { value: TagColor.accent3, label: t.tagsModal.tagColors[TagColor.accent3] },
    { value: TagColor.error, label: t.tagsModal.tagColors[TagColor.error] },
  ];

  const selectedTag = tagColorOptions.find((option) => option.value === tag);
  return { tagColorOptions, selectedTag };
}

function hasTagsData(result: unknown): result is { data: ExpandedTag } {
  return result !== null && Object.hasOwn(result || {}, "data");
}

function isCateogryValue(value: unknown): value is { value: number } {
  return value !== null && typeof value === "object" && "value" in value;
}

function isColorValue(value: unknown): value is { value: TagColor } {
  return value !== null && typeof value === "object" && "value" in value;
}

function useCategoryOptions(categoryId: number) {
  const { t } = useTranslations();
  const { data, isLoading } = useGetCategoriesQuery();
  const categories = {
    categoryOptions: data
      ? data.map((category) => {
          return { value: category.categoryId, label: t.tagsModal.tagCategories[category.name] };
        })
      : [],
    categoriesIsLoading: isLoading,
  };
  const selectedCategory = categories.categoryOptions.find((option) => option.value === categoryId);
  return {
    categoryOptions: categories.categoryOptions,
    selectedCategory,
    categoriesIsLoading: categories.categoriesIsLoading,
  };
}
