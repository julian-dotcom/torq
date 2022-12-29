import useTranslations from "services/i18n/useTranslations";
import { useState } from "react";
import { Timer16Regular as TimeTriggerIcon } from "@fluentui/react-icons";
import WorkflowNodeWrapper, { WorkflowNodeProps } from "components/workflow/nodeWrapper/WorkflowNodeWrapper";
import Input from "components/forms/input/Input";
import { InputSizeVariant } from "components/forms/input/variants";
import Form from "components/forms/form/Form";
import Socket from "components/forms/socket/Socket";
import { NodeColorVariant } from "../nodeVariants";

type TimeTriggerNodeProps = Omit<WorkflowNodeProps, "colorVariant">;

type channelPolicy = {
  frequency: number | undefined;
  timeUnit: string | undefined;
};

export function TimeTriggerNode<T>({ ...wrapperProps }: TimeTriggerNodeProps) {
  const { t } = useTranslations();

  const [channelPolicy, setChannelPolicy] = useState<channelPolicy>({
    frequency: undefined,
    timeUnit: undefined,
  });

  function createChangeHandler(key: keyof channelPolicy) {
    return (e: React.ChangeEvent<HTMLInputElement>) => {
      setChannelPolicy((prev) => ({
        ...prev,
        [key]: e.target.value,
      }));
    };
  }

  return (
    <WorkflowNodeWrapper
      {...wrapperProps}
      heading={t.channelPolicyConfiguration}
      headerIcon={<TimeTriggerIcon />}
      colorVariant={NodeColorVariant.accent2}
    >
      <Form>
        <Socket label={"Channels"} id={"sss"} />
        <Input
          formatted={true}
          value={channelPolicy.frequency}
          thousandSeparator={","}
          suffix={" ppm"}
          onChange={createChangeHandler("frequency")}
          label={t.feeRate}
          sizeVariant={InputSizeVariant.small}
        />
      </Form>
    </WorkflowNodeWrapper>
  );
}
