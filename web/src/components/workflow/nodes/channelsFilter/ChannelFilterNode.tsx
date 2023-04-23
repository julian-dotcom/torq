import React, { useContext, useEffect, useState } from "react";
import { Filter20Regular as FilterIcon, Save16Regular as SaveIcon } from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import WorkflowNodeWrapper, { WorkflowNodeProps } from "components/workflow/nodeWrapper/WorkflowNodeWrapper";
import Form from "components/forms/form/Form";
import Socket from "components/forms/socket/Socket";
import { NodeColorVariant } from "components/workflow/nodes/nodeVariants";
import { SelectWorkflowNodeLinks, SelectWorkflowNodes, useUpdateNodeMutation } from "pages/WorkflowPage/workflowApi";
import Button, { ColorVariant, SizeVariant } from "components/buttons/Button";
import { useSelector } from "react-redux";
import FilterComponent from "features/sidebar/sections/filter/FilterComponent";
import { AndClause, deserialiseQuery, OrClause } from "features/sidebar/sections/filter/filter";
import { ChannelsFilterTemplate } from "features/channels/channelsDefaults";
import { AllChannelsColumns } from "features/channels/channelsColumns.generated";
import Spinny from "features/spinny/Spinny";
import { WorkflowContext } from "components/workflow/WorkflowContext";
import { Status } from "constants/backend";
import { toastCategory } from "features/toast/Toasts";
import ToastContext from "features/toast/context";

type FilterChannelsNodeProps = Omit<WorkflowNodeProps, "colorVariant">;

export function ChannelFilterNode({ ...wrapperProps }: FilterChannelsNodeProps) {
  const { t } = useTranslations();
  const toastRef = React.useContext(ToastContext);
  const { workflowStatus } = useContext(WorkflowContext);
  const editingDisabled = workflowStatus === Status.Active;

  const [updateNode] = useUpdateNodeMutation();

  const [filterState, setFilterState] = useState(
    deserialiseQuery(wrapperProps.parameters || { $and: [] }) as AndClause | OrClause
  );

  const [dirty, setDirty] = useState(false);
  const [processing, setProcessing] = useState(false);
  useEffect(() => {
    if (
      Array.from(JSON.stringify(wrapperProps.parameters)).sort().join("") !==
      Array.from(JSON.stringify(filterState)).sort().join("")
    ) {
      setDirty(true);
    } else {
      setDirty(false);
    }
  }, [filterState, wrapperProps.parameters]);

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();

    if (editingDisabled) {
      toastRef?.current?.addToast(t.toast.cannotModifyWorkflowActive, toastCategory.warn);
      return;
    }

    setProcessing(true);
    updateNode({
      workflowVersionNodeId: wrapperProps.workflowVersionNodeId,
      parameters: filterState as unknown as Record<string, unknown>,
    }).finally(() => {
      setProcessing(false);
    });
  }

  const { childLinks } = useSelector(
    SelectWorkflowNodeLinks({
      version: wrapperProps.version,
      workflowId: wrapperProps.workflowId,
      nodeId: wrapperProps.workflowVersionNodeId,
      stage: wrapperProps.stage,
    })
  );

  const channelIds =
    childLinks
      ?.filter((n) => {
        return n.childInput === "channels";
      })
      ?.map((link) => link.parentWorkflowVersionNodeId) ?? [];

  const channels = useSelector(
    SelectWorkflowNodes({
      version: wrapperProps.version,
      workflowId: wrapperProps.workflowId,
      nodeIds: channelIds,
    })
  );

  const handleFilterUpdate = (filter: AndClause | OrClause) => {
    setFilterState(filter);
  };

  const filters = AllChannelsColumns.filter((column) => column.valueType !== "link");

  return (
    <WorkflowNodeWrapper
      {...wrapperProps}
      headerIcon={<FilterIcon />}
      colorVariant={NodeColorVariant.accent1}
      outputName={"channels"}
    >
      <Form onSubmit={handleSubmit} intercomTarget={"channel-filter-content"}>
        <Socket
          collapsed={wrapperProps.visibilitySettings.collapsed}
          label={t.channels}
          selectedNodes={channels || []}
          workflowVersionId={wrapperProps.workflowVersionId}
          workflowVersionNodeId={wrapperProps.workflowVersionNodeId}
          inputName={"channels"}
          editingDisabled={editingDisabled}
        />
        {filterState !== undefined && (
          <FilterComponent
            filters={filterState}
            columns={filters}
            defaultFilter={ChannelsFilterTemplate}
            child={false}
            onFilterUpdate={handleFilterUpdate}
            editingDisabled={editingDisabled}
          />
        )}
        <Button
          intercomTarget={"channel-filter-save-button"}
          type="submit"
          buttonColor={ColorVariant.success}
          buttonSize={SizeVariant.small}
          icon={!processing ? <SaveIcon /> : <Spinny />}
          disabled={!dirty || processing || editingDisabled}
        >
          {!processing ? t.save.toString() : t.saving.toString()}
        </Button>
      </Form>
    </WorkflowNodeWrapper>
  );
}
