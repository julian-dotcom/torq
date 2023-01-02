import styles from "./workflow_link.module.scss";

type WorkflowLinkProps = {
  workflowId: number;
  version: number;
  workflowVersionId: number;
};

export function WorkflowLink(props: WorkflowLinkProps) {
  return <svg className={styles.workflowLink}></svg>;
}
