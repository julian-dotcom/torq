import useTranslations from "services/i18n/useTranslations";
import { useState } from "react";
import { MoneySettings20Regular as ChannelPolicyIcon } from "@fluentui/react-icons";
import WorkflowNodeWrapper, { WorkflowNodeProps } from "components/workflow/nodeWrapper/WorkflowNodeWrapper";
import Input from "components/forms/input/Input";
import { InputSizeVariant } from "components/forms/input/variants";
import Form from "components/forms/form/Form";
import Socket from "components/forms/socket/Socket";
import { NodeColorVariant } from "../nodeVariants";

type ChannelPolicyNodeProps = Omit<WorkflowNodeProps, "colorVariant">;

type channelPolicy = {
  feeRate: number | undefined;
  baseFee: number | undefined;
  minHTLCAmount: number | undefined;
  maxHTLCAmount: number | undefined;
};

export function ChannelPolicyNode<T>({ ...wrapperProps }: ChannelPolicyNodeProps) {
  const { t } = useTranslations();

  const [channelPolicy, setChannelPolicy] = useState<channelPolicy>({
    feeRate: undefined,
    baseFee: undefined,
    minHTLCAmount: undefined,
    maxHTLCAmount: undefined,
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
      headerIcon={<ChannelPolicyIcon />}
      colorVariant={NodeColorVariant.accent1}
    >
      <Form>
        <Socket label={"Channels"} id={"sss"} />
        <Input
          formatted={true}
          value={channelPolicy.feeRate}
          thousandSeparator={","}
          suffix={" ppm"}
          onChange={createChangeHandler("feeRate")}
          label={t.feeRate}
          sizeVariant={InputSizeVariant.small}
        />
        <Input
          formatted={true}
          value={channelPolicy.baseFee}
          thousandSeparator={","}
          suffix={" sat"}
          onChange={createChangeHandler("baseFee")}
          label={t.baseFee}
          sizeVariant={InputSizeVariant.small}
        />
        <Input
          formatted={true}
          value={channelPolicy.minHTLCAmount}
          thousandSeparator={","}
          suffix={" sat"}
          onChange={createChangeHandler("minHTLCAmount")}
          label={t.minHTLCAmount}
          sizeVariant={InputSizeVariant.small}
        />
        <Input
          formatted={true}
          value={channelPolicy.maxHTLCAmount}
          thousandSeparator={","}
          suffix={" sat"}
          onChange={createChangeHandler("maxHTLCAmount")}
          label={t.maxHTLCAmount}
          sizeVariant={InputSizeVariant.small}
        />
      </Form>
    </WorkflowNodeWrapper>
  );
}
