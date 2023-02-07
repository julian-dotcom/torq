import classNames from "classnames";
import { Add16Regular as NewStageIcon, Delete16Regular as DeleteIcon } from "@fluentui/react-icons";
import { useDeleteStageMutation } from "pages/WorkflowPage/workflowApi";
import styles from "./workflow_stages.module.scss";
import { ReactComponent as StageArrowBack } from "pages/WorkflowPage/stageArrowBack.svg";
import { ReactComponent as StageArrowFront } from "pages/WorkflowPage/stageArrowFront.svg";
import { useAddNodeMutation } from "pages/WorkflowPage/workflowApi";
import { WorkflowNodeType } from "pages/WorkflowPage/constants";
import useTranslations from "services/i18n/useTranslations";
import mixpanel from "mixpanel-browser";
import React from "react";
import ToastContext from "features/toast/context";
import { toastCategory } from "features/toast/Toasts";

type StageSelectorProps = {
  stageNumbers: Array<number>;
  selectedStage: number;
  setSelectedStage: (stage: number) => void;
  workflowVersionId: number;
  workflowId: number;
  version: number;
  disabled: boolean;
};

export function StageSelector({
  stageNumbers,
  selectedStage,
  setSelectedStage,
  workflowVersionId,
  workflowId,
  version,
  disabled,
}: StageSelectorProps) {
  return (
    <div className={styles.stagesWrapper}>
      {stageNumbers.map((stage, index) => {
        return (
          <SelectStageButton
            key={`stage-${stage}`}
            selectedStage={selectedStage}
            setSelectedStage={setSelectedStage}
            stage={stage}
            stageNumbers={stageNumbers}
            buttonIndex={index}
            workflowId={workflowId}
            version={version}
            disabled={disabled}
          />
        );
      })}
      <AddStageButton
        disabled={disabled}
        setSelectedStage={setSelectedStage}
        workflowVersionId={workflowVersionId}
        selectedStage={selectedStage}
        stageNumbers={stageNumbers}
      />
    </div>
  );
}

type SelectStageButtonProps = {
  stage: number;
  stageNumbers: Array<number>;
  selectedStage: number;
  setSelectedStage: (stage: number) => void;
  buttonIndex: number;
  workflowId: number;
  version: number;
  disabled: boolean;
};

function SelectStageButton(props: SelectStageButtonProps) {
  const { stage, stageNumbers, selectedStage, setSelectedStage, buttonIndex, workflowId, version } = props;
  const { t } = useTranslations();
  const toastRef = React.useContext(ToastContext);
  const [deleteStage] = useDeleteStageMutation();

  function handleDeleteStage(stage: number) {
    if (props.disabled) {
      toastRef?.current?.addToast(t.toast.cannotModifyWorkflowActive, toastCategory.warn);
      return;
    }
    // Ask the user to confirm deletion of the stage
    if (!confirm(t.deleteStageConfirm)) {
      return;
    }
    mixpanel.track("Workflow Delete Stage", {
      workflowId: workflowId,
      workflowVersion: version,
      workflowStage: stage,
    });
    deleteStage({ workflowId, version, stage }).then(() => {
      // On success, select the preceding stage
      const precedingStage = stageNumbers.slice(0, stageNumbers.indexOf(stage)).pop();
      setSelectedStage(precedingStage || 1);
    });
  }

  // NB: The stage is the stage ID used on the nodes. The buttonIndex is used to display the stage number.
  //   This is because the user can delete a stage in between two stages, and then the stage numbers will not be consecutive.

  return (
    <button
      className={classNames(styles.stageContainer, { [styles.selected]: stage === selectedStage })}
      onClick={() => setSelectedStage(stage)}
    >
      {buttonIndex !== 0 && <StageArrowBack />}
      <div className={styles.stage}>
        {`${t.stage} ${buttonIndex + 1}`}
        {buttonIndex !== 0 && (
          <div className={classNames(styles.deleteStage)} onClick={() => handleDeleteStage(stage)}>
            <DeleteIcon />
          </div>
        )}
      </div>
      <StageArrowFront />
    </button>
  );
}

type AddStageButtonProps = {
  stageNumbers: Array<number>;
  selectedStage: number;
  setSelectedStage: (stage: number) => void;
  workflowVersionId: number;
  disabled: boolean;
};

function AddStageButton(props: AddStageButtonProps) {
  const { t } = useTranslations();
  const toastRef = React.useContext(ToastContext);
  const [addNode] = useAddNodeMutation();
  const nextStage = Math.max(...props.stageNumbers) + 1;

  function handleAddStage() {
    if (props.disabled) {
      toastRef?.current?.addToast(t.toast.cannotModifyWorkflowActive, toastCategory.warn);
      return;
    }

    mixpanel.track("Workflow Add Stage", {
      workflowVersionId: props.workflowVersionId,
      workflowCurrentStage: props.selectedStage,
      workflowNextStage: nextStage,
    });
    addNode({
      type: WorkflowNodeType.StageTrigger,
      name: `${t.stage} ${nextStage}`,
      visibilitySettings: {
        xPosition: 0,
        yPosition: 0,
        collapsed: false,
      },
      workflowVersionId: props.workflowVersionId,
      stage: nextStage,
    }).then(() => {
      // On success, select the new stage
      props.setSelectedStage(nextStage);
    });
  }

  return (
    <button
      className={classNames(styles.stageContainer, props.disabled ? styles.disabledStage : styles.addStageButton)}
      onClick={handleAddStage}
    >
      <StageArrowBack />
      <div className={classNames(styles.stage, props.disabled ? styles.disabled : "")}>
        <NewStageIcon />
      </div>
      <StageArrowFront />
    </button>
  );
}
