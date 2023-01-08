export const aaa = 1;
// import { Tag24Regular as TagHeaderIcon } from "@fluentui/react-icons"; // ArrowRouting20Regular as ChannelsIcon
// import Button, { ColorVariant, ButtonWrapper } from "components/buttons/Button";
// import ProgressTabs from "features/progressTabs/ProgressTab";
// import PopoutPageTemplate from "features/templates/popoutPageTemplate/PopoutPageTemplate";
// import { ChangeEvent, useState, useEffect } from "react";
// import { useNavigate } from "react-router";
// import styles from "./modifyTagModal.module.scss";
// import useTranslations from "services/i18n/useTranslations";
// import Select, { SelectOptions } from "features/forms/Select";
// import { ActionMeta } from "react-select";
// // import clone from "clone";
// import classNames from "classnames";
// import FormRow from "features/forms/FormWrappers";
// import Input from "components/forms/input/Input";
// import { useGetCategoriesQuery } from "pages/categoriesPage/categoriesApi";
// import { Category } from "pages/categoriesPage/categoriesTypes";
// import { useAddTagMutation, useGetTagQuery, useSetTagMutation, useGetCorridorByReferenceQuery } from "./tagsApi";
// import { Tag, Corridor } from "./tagsTypes"; // CorridorFields, ChannelGroup
// import { TagColor } from "components/tags/Tag";
// // import { useLocation } from "react-router";
// // import { UPDATE_TAG } from "constants/routes";
// // import Target from "components/targets/Target";
// // import TargetsSection from "./targetsSection/TargetsSection";
//
// function ModifyTagModal() {
//   // const location = useLocation();
//   const { t } = useTranslations();
//
//   // const [deleteChannelGroupByTagMutation, deleteChannelGroupByTagResponse] = useDeleteChannelGroupByTagMutation();
//
//   // const [errMessage, setErrorMessage] = useState<ReactNode[]>([]);
//   const [stepIndex, setStepIndex] = useState(0);
//   const [tagName, setTagName] = useState<string>("");
//   const [tagId, setTagId] = useState<number>(0);
//   // const [nodeId] = useState<number>(0);
//   // const [channelId, setChannelId] = useState<number>(0);
//
//   // const [targetNodes, setTargetNodes] = useState<ReactNode[]>([]);
//   // const [targetChannels, setTargetChannels] = useState<ReactNode[]>([]);
//   const [colorName] = useState<string>("N/A"); // setColorName
//   const tagColorOptions: SelectOptions[] = [
//     { value: "", label: "Select your color" },
//     { value: TagColor.primary, label: TagColor.primary.charAt(0).toUpperCase() + TagColor.primary.slice(1) },
//     { value: TagColor.success, label: TagColor.success.charAt(0).toUpperCase() + TagColor.success.slice(1) },
//     { value: TagColor.warning, label: TagColor.warning.charAt(0).toUpperCase() + TagColor.warning.slice(1) },
//     { value: TagColor.error, label: TagColor.error.charAt(0).toUpperCase() + TagColor.error.slice(1) },
//     { value: TagColor.accent1, label: TagColor.accent1.charAt(0).toUpperCase() + TagColor.accent1.slice(1) },
//     { value: TagColor.accent2, label: TagColor.accent2.charAt(0).toUpperCase() + TagColor.accent2.slice(1) },
//     { value: TagColor.accent3, label: TagColor.accent3.charAt(0).toUpperCase() + TagColor.accent3.slice(1) },
//   ];
//
//   const queryParameters = new URLSearchParams(window.location.search);
//
//   const urlTagId = queryParameters.get("tagId");
//
//   useEffect(() => {
//     setTagId(Number(urlTagId));
//   }, [urlTagId]);
//
//   const [selectedTagColor, setTagColor] = useState<string>(tagColorOptions[0].value as string);
//
//   // const { data: categoriesResponse } = useGetCategoriesQuery<{
//   //   data: Array<Category>;
//   //   isLoading: boolean;
//   //   isFetching: boolean;
//   //   isUninitialized: boolean;
//   //   isSuccess: boolean;
//   // }>();
//   // let tagCategorieOptions: SelectOptions[] = [{ value: 0, label: "Select your Category" }];
//   // if (categoriesResponse !== undefined) {
//   //   tagCategorieOptions = categoriesResponse.map((tagCategorieOptions: Category) => {
//   //     return { value: tagCategorieOptions.categoryId, label: tagCategorieOptions.name };
//   //   });
//   // }
//   // const [selectedTagCategory, setTagCategory] = useState<number>(tagCategorieOptions[0].value as number);
//   let categoryName = "N/A";
//
//   const { data } = useGetTagQuery<{
//     data: Tag;
//     isLoading: boolean;
//     isFetching: boolean;
//     isUninitialized: boolean;
//     isSuccess: boolean;
//   }>(tagId);
//   //
//   // const { data: corridorsResponse } = useGetCorridorByReferenceQuery<{
//   //   data: Corridor;
//   //   isLoading: boolean;
//   //   isFetching: boolean;
//   //   isUninitialized: boolean;
//   //   isSuccess: boolean;
//   // }>(tagId);
//
//   const [addTagMutation] = useAddTagMutation(); // addTagResponse
//   const [setTagMutation] = useSetTagMutation(); // setTagResponse
//
//   // function handleDeleteTarget(corridorId: number) {
//   //   deleteChannelGroupByTagMutation(corridorId);
//   // }
//
//   // useEffect(() => {
//   //   let tag = 0;
//   //   const message = clone(errMessage) || [];
//   //
//   //   if (tagResponse) {
//   //     setTagName(tagResponse.name);
//   //     setTagColor(tagResponse.style);
//   //     setTagCategory(tagResponse?.categoryId || 0);
//   //
//   //     categoriesResponse?.forEach((tagCategorieOptions: Category) => {
//   //       if (tagResponse?.categoryId == tagCategorieOptions.categoryId) {
//   //         categoryName = tagCategorieOptions.name;
//   //       }
//   //     });
//   //     tagColorOptions.forEach((colors) => {
//   //       if (tagResponse.style == (colors.value as string)) {
//   //         setColorName(colors?.label || "N/A");
//   //       }
//   //     });
//   //   }
//   //
//   //   if (corridorsResponse) {
//   //     // setTotalNodes(corridorsResponse.totalNodes);
//   //     // setTotalChannels(corridorsResponse.totalChannels);
//   //     const listNodes: ReactNode[] = [];
//   //     const listChannels: ReactNode[] = [];
//   //     corridorsResponse.corridors?.map((c: CorridorFields) => {
//   //       if (c.shortChannelId) {
//   //         listChannels.push(
//   //           <Target
//   //             onDeleteTarget={() => handleDeleteTarget(c.corridorId)}
//   //             key={"channel-target-" + c.corridorId}
//   //             icon={<ChannelsIcon />}
//   //             details={c.alias}
//   //             title={c.shortChannelId}
//   //           />
//   //         );
//   //       } else {
//   //         listNodes.push(
//   //           <Target
//   //             onDeleteTarget={() => handleDeleteTarget(c.corridorId)}
//   //             key={"channel-target-" + c.corridorId}
//   //             icon={<ChannelsIcon />}
//   //             details={"-"}
//   //             title={c.alias}
//   //           />
//   //         );
//   //       }
//   //     });
//   //     setTargetNodes(listNodes);
//   //     setTargetChannels(listChannels);
//   //   }
//   //
//   //   if (
//   //     (addTagResponse.isSuccess || setTagResponse.isSuccess) &&
//   //     !addChannelsGroupsResponse?.data &&
//   //     addChannelsGroupsResponse.status != "pending"
//   //   ) {
//   //     if (addTagResponse.isSuccess && !addTagResponse.isLoading) {
//   //       tag = addTagResponse.data.tagId ? addTagResponse.data.tagId : 0;
//   //     }
//   //     if (setTagResponse.isSuccess && !setTagResponse.isLoading) {
//   //       tag = tagId;
//   //     }
//   //     setTagId(tag);
//   //     if (
//   //       (addTagResponse.isSuccess && addTagResponse.status != "fulfilled") ||
//   //       (setTagResponse.isSuccess && setTagResponse.status != "fulfilled")
//   //     ) {
//   //       message.push(
//   //         <span key={0} className={classNames(styles.modifyTagsStatusMessage)}>
//   //           {"Error saving the tag"}
//   //         </span>
//   //       );
//   //       setErrorMessage(message);
//   //     }
//   //
//   //     if (addTagResponse.isSuccess && addTagResponse.status == "fulfilled") {
//   //       location.pathname = "/manage/tags";
//   //       navigate(`${UPDATE_TAG}?tagId=${tag}`, { state: { background: location }, replace: true });
//   //     }
//   //   }
//   //   if (addTagResponse.isError || setTagResponse.isError) {
//   //     message.push(
//   //       <span key={0} className={classNames(styles.modifyTagsStatusMessage)}>
//   //         {"Error saving the tag"}
//   //       </span>
//   //     );
//   //     setErrorMessage(message);
//   //   }
//   // }, [addTagResponse, setTagResponse, tagResponse, corridorsResponse, deleteChannelGroupByTagResponse]);
//
//   // function handleTarget(value: number, type: string) {
//   //   // setTarget(value);
//   //   // setType(type);
//   //   let selectedChannelNode;
//   //   if (type == "channel") {
//   //     selectedChannelNode = channelsNodesResponse.channels.find((channelsOptions: ChannelForTag) => {
//   //       if (channelsOptions.channelId == value) {
//   //         return channelsOptions;
//   //       }
//   //     });
//   //     setChannelId(value);
//   //     setNodeId(selectedChannelNode?.nodeId as number | 0);
//   //   } else {
//   //     setNodeId(value);
//   //   }
//   // }
//
//   const closeAndReset = () => {
//     setStepIndex(0);
//     setTagName("");
//     setTagColor(tagColorOptions[0].value as string);
//     setTagCategory(tagCategorieOptions[0].value as number);
//     // setTarget(channelsNodesOptions[0].value as number);
//     navigate(-1);
//   };
//
//   const navigate = useNavigate();
//
//   return (
//     <PopoutPageTemplate title={t.tags} show={true} onClose={() => closeAndReset()} icon={<TagHeaderIcon />}>
//       <ProgressTabs showTabIndex={stepIndex}>
//         <div className={styles.activeColumns}>
//           <FormRow>
//             <div className={styles.updateChannelTableDouble}>
//               <div className={styles.input}>
//                 <Input
//                   disabled={!!(tagId && tagId < 0)}
//                   label={"Name"}
//                   className={classNames(styles.simple, { [styles.lockedInput]: tagId < 0 })}
//                   value={tagName}
//                   onChange={(e: ChangeEvent<HTMLInputElement>) => {
//                     setTagName(e.target.value);
//                   }}
//                 />
//               </div>
//             </div>
//           </FormRow>
//           <FormRow>
//             {tagId >= 0 && (
//               <div className={styles.updateChannelTableDouble}>
//                 <Select
//                   label={t.tagsModal.color}
//                   onChange={(newValue: unknown, _: ActionMeta<unknown>) => {
//                     const selectOptions = newValue as SelectOptions;
//                     setTagColor(selectOptions?.value as string);
//                   }}
//                   options={tagColorOptions}
//                   value={tagColorOptions.find((option) => option.value === selectedTagColor)}
//                 />
//               </div>
//             )}
//             {tagId < 0 && (
//               <div className={styles.updateChannelTableDouble}>
//                 <div className={styles.input}>
//                   <Input
//                     disabled={!!(tagId && tagId < 0)}
//                     label={t.tagsModal.color}
//                     className={classNames(styles.simple, styles.lockedInput)}
//                     value={colorName}
//                   />
//                 </div>
//               </div>
//             )}
//             {tagId >= 0 && (
//               <div className={styles.updateChannelTableDouble}>
//                 <Select
//                   label={t.tagsModal.category}
//                   onChange={(newValue: unknown, _: ActionMeta<unknown>) => {
//                     const selectOptions = newValue as SelectOptions;
//                     setTagCategory(selectOptions?.value as number);
//                   }}
//                   options={tagCategorieOptions}
//                   value={tagCategorieOptions.find((option) => option.value === selectedTagCategory)}
//                 />
//               </div>
//             )}
//             {tagId < 0 && (
//               <div className={styles.updateChannelTableDouble}>
//                 <div className={styles.input}>
//                   <Input
//                     disabled={!!(tagId && tagId < 0)}
//                     label={t.tagsModal.category}
//                     className={classNames(styles.simple, styles.lockedInput)}
//                     value={categoryName}
//                   />
//                 </div>
//               </div>
//             )}
//           </FormRow>
//
//           <ButtonWrapper
//             className={styles.customButtonWrapperStyles}
//             rightChildren={
//               <Button
//                 className={styles.tagbutton}
//                 onClick={() => {
//                   const tagObj: Tag = {
//                     style: selectedTagColor,
//                     name: tagName,
//                   };
//                   if (selectedTagCategory != 0) {
//                     tagObj.categoryId = selectedTagCategory;
//                   }
//                   if (urlTagId) {
//                     tagObj.tagId = tagId;
//                     setTagMutation(tagObj);
//                   } else {
//                     addTagMutation(tagObj);
//                   }
//                 }}
//                 buttonColor={ColorVariant.primary}
//               >
//                 {t.submit}
//               </Button>
//             }
//           />
//
//           {/*<TargetsSection channels={targetChannels} nodes={targetNodes} />*/}
//         </div>
//       </ProgressTabs>
//     </PopoutPageTemplate>
//   );
// }
//
// export default ModifyTagModal;
