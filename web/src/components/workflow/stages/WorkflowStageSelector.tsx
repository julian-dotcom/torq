import classNames from "classnames";
import { Add16Regular as NewStageIcon, Delete16Regular as DeleteIcon } from "@fluentui/react-icons";
import { useDeleteStageMutation } from "pages/WorkflowPage/workflowApi";
import styles from "./workflow_stages.module.scss";
import { ReactComponent as StageArrowBack } from "pages/WorkflowPage/stageArrowBack.svg";
import { ReactComponent as StageArrowFront } from "pages/WorkflowPage/stageArrowFront.svg";
import { useAddNodeMutation } from "pages/WorkflowPage/workflowApi";
import { WorkflowNodeType } from "pages/WorkflowPage/constants";
import useTranslations from "services/i18n/useTranslations";

type StageSelectorProps = {
  stageNumbers: Array<number>;
  selectedStage: number;
  setSelectedStage: (stage: number) => void;
  workflowVersionId: number;
  workflowId: number;
  version: number;
};

export function StageSelector({
  stageNumbers,
  selectedStage,
  setSelectedStage,
  workflowVersionId,
  workflowId,
  version,
}: StageSelectorProps) {
  const { t } = useTranslations();

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
      name: `${t.stage} ${nextStage}`,
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
    <button className={classNames(styles.stageContainer, styles.addStageButton)} onClick={handleAddStage}>
      <StageArrowBack />
      <div className={styles.stage}>
        <NewStageIcon />
      </div>
      <StageArrowFront />
    </button>
  );
}
