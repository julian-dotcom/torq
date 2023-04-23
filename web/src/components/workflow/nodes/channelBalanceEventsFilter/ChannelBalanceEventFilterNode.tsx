import React, { useContext, useEffect, useState } from "react";
import { ArrowBounce20Regular as ChannelBalanceEventFilterIcon, Save16Regular as SaveIcon } from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import WorkflowNodeWrapper, { WorkflowNodeProps } from "components/workflow/nodeWrapper/WorkflowNodeWrapper";
import Form from "components/forms/form/Form";
import Socket from "components/forms/socket/Socket";
import { NodeColorVariant } from "components/workflow/nodes/nodeVariants";
import { SelectWorkflowNodeLinks, SelectWorkflowNodes, useUpdateNodeMutation } from "pages/WorkflowPage/workflowApi";
import Button, { ColorVariant, SizeVariant } from "components/buttons/Button";
import { useSelector } from "react-redux";
import FilterComponent from "features/sidebar/sections/filter/FilterComponent";
import { AndClause, deserialiseQuery, FilterInterface, OrClause } from "features/sidebar/sections/filter/filter";
import Spinny from "features/spinny/Spinny";
import { WorkflowContext } from "components/workflow/WorkflowContext";
import { Status } from "constants/backend";
import { toastCategory } from "features/toast/Toasts";
import ToastContext from "features/toast/context";
import { ColumnMetaData } from "features/table/types";
import { InputSizeVariant, RadioChips } from "components/forms/forms";

type FilterEventsNodeProps = Omit<WorkflowNodeProps, "colorVariant">;

export type event = {
  balanceDelta: number;
  balanceDeltaAbsolute: number;
  balanceUpdateEventOrigin: number;
};

export type ChannelBalanceEventFilterConfiguration = {
  ignoreWhenEventless: boolean;
  filterClauses: AndClause | OrClause;
};

export function ChannelBalanceEventFilterNode({ ...wrapperProps }: FilterEventsNodeProps) {
  const { t } = useTranslations();
  const toastRef = React.useContext(ToastContext);
  const { workflowStatus } = useContext(WorkflowContext);
  const editingDisabled = workflowStatus === Status.Active;

  const [updateNode] = useUpdateNodeMutation();

  const [configuration, setConfiguration] = useState<ChannelBalanceEventFilterConfiguration>({
    ignoreWhenEventless: (wrapperProps.parameters.ignoreWhenEventless || false) as boolean,
    filterClauses: deserialiseQuery(wrapperProps.parameters.filterClauses || { $and: [] }) as AndClause | OrClause,
  });

  const [dirty, setDirty] = useState(false);
  const [processing, setProcessing] = useState(false);
  useEffect(() => {
    if (
      Array.from(JSON.stringify(wrapperProps.parameters)).sort().join("") !==
      Array.from(JSON.stringify(configuration)).sort().join("")
    ) {
      setDirty(true);
    } else {
      setDirty(false);
    }
  }, [configuration, wrapperProps.parameters]);

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();

    if (editingDisabled) {
      toastRef?.current?.addToast(t.toast.cannotModifyWorkflowActive, toastCategory.warn);
      return;
    }

    setProcessing(true);
    updateNode({
      workflowVersionNodeId: wrapperProps.workflowVersionNodeId,
      parameters: configuration,
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
    setConfiguration((prev) => ({
      ...prev,
      filterClauses: filter,
    }));
  };

  const TypeLabels = new Map<string, string>([
    ["0", t.forwardEvent],
    ["1", t.invoiceEvent],
    ["2", t.paymentEvent],
  ]);

  const filters: ColumnMetaData<event>[] = [
    {
      heading: "Balance Delta",
      key: "balanceDelta",
      type: "NumericCell",
      valueType: "number",
    },
    {
      heading: "Balance Delta (absolute)",
      key: "balanceDeltaAbsolute",
      type: "NumericCell",
      valueType: "number",
    },
    {
      heading: "Event Type",
      type: "TextCell",
      key: "balanceUpdateEventOrigin",
      valueType: "enum",
      selectOptions: [
        { label: TypeLabels.get("0") || "", value: "0" },
        { label: TypeLabels.get("1") || "", value: "1" },
        { label: TypeLabels.get("2") || "", value: "2" },
      ],
    },
  ];

  const eventsFilterTemplate: FilterInterface = {
    funcName: "gte",
    category: "number",
    parameter: 1,
    key: "balanceDeltaAbsolute",
  };

  return (
    <WorkflowNodeWrapper
      {...wrapperProps}
      headerIcon={<ChannelBalanceEventFilterIcon />}
      colorVariant={NodeColorVariant.accent1}
      outputName={"channels"}
    >
      <Form onSubmit={handleSubmit} intercomTarget={"channel-balance-event-filter-node-form"}>
        <Socket
          collapsed={wrapperProps.visibilitySettings.collapsed}
          label={t.channels}
          selectedNodes={channels || []}
          workflowVersionId={wrapperProps.workflowVersionId}
          workflowVersionNodeId={wrapperProps.workflowVersionNodeId}
          inputName={"channels"}
          editingDisabled={editingDisabled}
        />
        <RadioChips
          label={t.ignoreWhenEventless}
          sizeVariant={InputSizeVariant.small}
          groupName={"ignore-switch-" + wrapperProps.workflowVersionNodeId}
          helpText={t.channelBalanceEventFilterNode.ignoreWhenEventlessHelpText}
          options={[
            {
              label: t.stop,
              id: "ignore-switch-true-" + wrapperProps.workflowVersionNodeId,
              checked: !configuration.ignoreWhenEventless,
              onChange: () =>
                setConfiguration((prev) => ({
                  ...prev,
                  ignoreWhenEventless: false,
                })),
            },
            {
              label: t.continue,
              id: "ignore-switch-false-" + wrapperProps.workflowVersionNodeId,
              checked: configuration.ignoreWhenEventless,
              onChange: () =>
                setConfiguration((prev) => ({
                  ...prev,
                  ignoreWhenEventless: true,
                })),
            },
          ]}
          editingDisabled={editingDisabled}
        />
        {configuration.filterClauses !== undefined && (
          <FilterComponent
            filters={configuration.filterClauses}
            columns={filters}
            defaultFilter={eventsFilterTemplate}
            child={false}
            onFilterUpdate={handleFilterUpdate}
            editingDisabled={editingDisabled}
          />
        )}
        <Button
          intercomTarget={"workflow-node-save"}
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
