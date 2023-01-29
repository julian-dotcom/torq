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
import mixpanel from "mixpanel-browser";

type WorkflowControlsProps = {
  sidebarExpanded: boolean;
  setSidebarExpanded: (expanded: boolean) => void;
  workflowId: number;
  status: Status;
};

export default function WorkflowControls(props: WorkflowControlsProps) {
  const { t } = useTranslations();

  const [updateWorkflow] = useUpdateWorkflowMutation();

  return (
    <TableControlSection>
      <TableControlsButtonGroup>
        <TableControlsTabsGroup>
          <Button
            buttonColor={props.status === Status.Active ? ColorVariant.warning : ColorVariant.success}
            hideMobileText={true}
            icon={props.status === Status.Active ? <DeactivateIcon /> : <ActivateIcon />}
            onClick={() => {
              mixpanel.track("Workflow Toggle Status", {
                status: props.status === Status.Active ? "Inactive" : "Active",
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
        <Button
          buttonColor={ColorVariant.primary}
          hideMobileText={true}
          id={"tableControlsButton"}
          icon={<NodesIcon />}
          onClick={() => {
            mixpanel.track("Workflow Toggle Sidebar");
            props.setSidebarExpanded(!props.sidebarExpanded);
          }}
        >
          {t.actions}
        </Button>
      </TableControlsButtonGroup>
    </TableControlSection>
  );
}
