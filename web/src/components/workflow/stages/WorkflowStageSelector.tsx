import classNames from "classnames";
import { Add16Regular as NewStageIcon, Delete16Regular as DeleteIcon } from "@fluentui/react-icons";
import { useDeleteStageMutation } from "pages/WorkflowPage/workflowApi";
import styles from "./workflow_stages.module.scss";
import { ReactComponent as StageArrowBack } from "pages/WorkflowPage/stageArrowBack.svg";
import { ReactComponent as StageArrowFront } from "pages/WorkflowPage/stageArrowFront.svg";
import { useAddNodeMutation } from "pages/WorkflowPage/workflowApi";
import { WorkflowNodeType } from "pages/WorkflowPage/constants";
import { WorkflowStages } from "pages/WorkflowPage/workflowTypes";
import useTranslations from "services/i18n/useTranslations";

type StageSelectorProps = {
  stages: WorkflowStages;
  selectedStage: number;
  setSelectedStage: (stage: number) => void;
  workflowVersionId: number;
  workflowId: number;
  version: number;
};

export function StageSelector({
  stages,
  selectedStage,
  setSelectedStage,
  workflowVersionId,
  workflowId,
  version,
}: StageSelectorProps) {
  const { t } = useTranslations();
  const stageNumbers = Object.keys(stages).map((s) => parseInt(s));

  return (
    <div className={styles.stagesWrapper}>
      {Object.keys(stages).map((stage, index) => {
        return (
          <SelectStageButton
            selectedStage={selectedStage}
            setSelectedStage={setSelectedStage}
            stage={parseInt(stage)}
            stageNumbers={stageNumbers}
            buttonIndex={index}
            workflowId={workflowId}
            version={version}
          />
        );
      })}
      <AddStageButton
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
};

function SelectStageButton(props: SelectStageButtonProps) {
  const { stage, stageNumbers, selectedStage, setSelectedStage, buttonIndex, workflowId, version } = props;
  const { t } = useTranslations();

  const [deleteStage] = useDeleteStageMutation();

  function handleDeleteStage(stage: number) {
    // Ask the user to confirm deletion of the stage
    if (!confirm(t.deleteStageConfirm)) {
      return;
    }
    deleteStage({ workflowId, version, stage }).then(() => {
      // On success, select the preceding stage
      const precedingStage = stageNumbers.slice(0, stageNumbers.indexOf(stage)).pop();
      setSelectedStage(precedingStage || 1);
    });
  }

  return (
    <button
      key={`stage-${stage}`}
      className={classNames(styles.stageContainer, { [styles.selected]: stage === selectedStage })}
      onClick={() => setSelectedStage(stage)}
    >
      {buttonIndex !== 0 && <StageArrowBack />}
      <div className={styles.stage}>
        {`${t.stage} ${stage}`}
        {buttonIndex !== 0 && (
          <div className={styles.deleteStage} onClick={() => handleDeleteStage(stage)}>
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
};

function AddStageButton(props: AddStageButtonProps) {
  const { t } = useTranslations();
  const [addNode] = useAddNodeMutation();
  const nextStage = Math.max(...props.stageNumbers) + 1;

  function handleAddStage() {
    addNode({
      type: WorkflowNodeType.StageTrigger,
      visibilitySettings: {
        xPosition: 0,
        yPosition: 0,
        collapsed: true,
      },
      workflowVersionId: props.workflowVersionId,
      stage: nextStage,
    }).then(() => {
      // On success, select the new stage
      props.setSelectedStage(nextStage);
    });
  }

  return (
    <button className={styles.addStageButton} onClick={handleAddStage}>
      <NewStageIcon />
    </button>
  );
}
