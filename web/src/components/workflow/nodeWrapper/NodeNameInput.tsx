import { Save20Regular as SaveIcon } from "@fluentui/react-icons";
import Form from "components/forms/form/Form";
import Input from "components/forms/input/Input";
import { InputColorVaraint, InputSizeVariant } from "components/forms/input/variants";
import styles from "./workflow_nodes.module.scss";
import { useUpdateNodeMutation } from "pages/WorkflowPage/workflowApi";
import { useState } from "react";

type NodeNameInputProps = { nodeId: number; name: string; isVisible: boolean; setVisible: (visible: boolean) => void };

export default function NodeNameInput(props: NodeNameInputProps) {
  const [updateNode] = useUpdateNodeMutation();

  const { nodeId, name, isVisible, setVisible } = props;
  function handleDoubleClick() {
    setVisible(true);
  }

  const [localName, setLocalName] = useState(name);

  function handleFormSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setVisible(false);
    setLocalName(e.currentTarget.something.value);
    updateNode({ workflowVersionNodeId: nodeId, name: e.currentTarget.something.value });
  }

  function handleBlurNameInput() {
    setVisible(false);
  }

  return (
    <div className={styles.title} onDoubleClick={handleDoubleClick} data-intercom-target={"workflow-node-title"}>
      {isVisible ? (
        <Form
          onSubmit={handleFormSubmit}
          className={styles.nameForm}
          intercomTarget={"workflow-node-title-form"}
        >
          <Input
            intercomTarget={"workflow-node-title-input"}
            name={"something"}
            id={"something"}
            className={styles.input}
            autoFocus={true}
            defaultValue={localName}
            sizeVariant={InputSizeVariant.small}
            colorVariant={InputColorVaraint.accent1}
            onKeyDown={(e) => {
              if (e.key === "Escape") {
                handleBlurNameInput();
              }
            }}
          />
          <button type={"submit"}>
            <SaveIcon />
          </button>
        </Form>
      ) : (
        name
      )}
    </div>
  );
}
