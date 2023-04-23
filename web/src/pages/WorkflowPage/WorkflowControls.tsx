import {
  Play20Regular as ActivateIcon,
  Pause20Regular as DeactivateIcon,
  PuzzlePiece20Regular as NodesIcon,
} from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import {
  TableControlsButtonGroup,
  TableControlSection,
  TableControlsTabsGroup,
} from "features/templates/tablePageTemplate/TablePageTemplate";
import Button, { ColorVariant } from "components/buttons/Button";
import { useUpdateWorkflowMutation } from "./workflowApi";
import { Status } from "constants/backend";
import { userEvents } from "utils/userEvents";

type WorkflowControlsProps = {
  sidebarExpanded: boolean;
  setSidebarExpanded: (expanded: boolean) => void;
  workflowId: number;
  status: Status;
};

export default function WorkflowControls(props: WorkflowControlsProps) {
  const { t } = useTranslations();
  const { track } = userEvents();
  const [updateWorkflow] = useUpdateWorkflowMutation();

  return (
    <TableControlSection>
      <TableControlsButtonGroup>
        <TableControlsTabsGroup>
          <Button
            intercomTarget={"activate-workflow-button"}
            buttonColor={props.status === Status.Active ? ColorVariant.warning : ColorVariant.success}
            hideMobileText={true}
            icon={props.status === Status.Active ? <DeactivateIcon /> : <ActivateIcon />}
            onClick={() => {
              track("Workflow Toggle Status", {
                workflowStatus: props.status === Status.Active ? "Inactive" : "Active",
                workflowId: props.workflowId,
              });
              if (props.status === Status.Inactive && !confirm(t.workflowDetails.confirmWorkflowActivate)) {
                return;
              }
              if (props.status === Status.Active && !confirm(t.workflowDetails.confirmWorkflowDeactivate)) {
                return;
              }
              updateWorkflow({
                workflowId: props.workflowId,
                status: props.status === Status.Active ? Status.Inactive : Status.Active,
              });
            }}
          >
            {props.status === Status.Active ? t.deactivate : t.activate}
          </Button>
        </TableControlsTabsGroup>
      </TableControlsButtonGroup>
      <TableControlsButtonGroup>
        <Button
          intercomTarget={"activate-workflow-actions-button"}
          buttonColor={ColorVariant.primary}
          hideMobileText={true}
          id={"tableControlsButton"}
          icon={<NodesIcon />}
          onClick={() => {
            track("Workflow Toggle Sidebar");
            props.setSidebarExpanded(!props.sidebarExpanded);
          }}
        >
          {t.actions}
        </Button>
      </TableControlsButtonGroup>
    </TableControlSection>
  );
}
