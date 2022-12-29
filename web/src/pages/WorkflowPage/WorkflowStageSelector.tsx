import classNames from "classnames";
import { Add16Regular as NewStageIcon, Delete16Regular as DeleteIcon } from "@fluentui/react-icons";
import { useDeleteStageMutation } from "pages/WorkflowPage/workflowApi";
import styles from "./workflow_page.module.scss";
import { ReactComponent as StageArrowBack } from "./stageArrowBack.svg";
import { ReactComponent as StageArrowFront } from "./stageArrowFront.svg";
import { useAddNodeMutation } from "./workflowApi";
import { WorkflowNodeType } from "./constants";
import { WorkflowStages } from "./workflowTypes";
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
  const [deleteStage] = useDeleteStageMutation();
  const stageNumbers = Object.keys(stages).map((s) => parseInt(s));

  function handleDeleteStage(stage: number) {
    deleteStage({ workflowId, version, stage }).then(() => {
      // On success, select the preceding stage
      const precedingStage = stageNumbers.slice(0, stageNumbers.indexOf(stage)).pop();
      setSelectedStage(precedingStage || 1);
    });
  }

  const stageButtons = Object.keys(stages).map((stage, index) => {
    return (
      <button
        key={`stage-${stage}`}
        className={classNames(styles.stageContainer, { [styles.selected]: parseInt(stage) === selectedStage })}
        onClick={() => setSelectedStage(parseInt(stage))}
      >
        {index !== 0 && <StageArrowBack />}
        <div className={styles.stage}>
          {`${t.stage} ${stage}`}
          {index !== 0 && (
            <div className={styles.deleteStage} onClick={() => handleDeleteStage(parseInt(stage))}>
              <DeleteIcon />
            </div>
          )}
        </div>
        <StageArrowFront />
      </button>
    );
  });

  const [addNode] = useAddNodeMutation();

  function addStage() {
    addNode({
      type: WorkflowNodeType.StageTrigger,
      visibilitySettings: {
        xPosition: 0,
        yPosition: 0,
        collapsed: true,
      },
      workflowVersionId: workflowVersionId,
      stage: Math.max(...Object.keys(stages).map((stage) => parseInt(stage))) + 1,
    }).then(() => {
      // On success, select the new stage
      setSelectedStage(Math.max(...Object.keys(stages).map((stage) => parseInt(stage))) + 1);
    });
  }

  const addStageButton = (
    <button key={`stage-add-stage`} className={classNames(styles.stageContainer)} onClick={addStage}>
      <StageArrowBack />
      <div className={styles.stage}>
        <NewStageIcon />
      </div>
      <StageArrowFront />
    </button>
  );

  return (
    <div className={styles.stagesWrapper}>
      {stageButtons}
      {addStageButton}
    </div>
  );
}
