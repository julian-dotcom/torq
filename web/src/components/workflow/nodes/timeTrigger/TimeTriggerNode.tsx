import { Save16Regular as SaveIcon } from "@fluentui/react-icons";
import { useEffect, useState } from "react";
import useTranslations from "services/i18n/useTranslations";
import { Timer16Regular as TimeTriggerIcon } from "@fluentui/react-icons";
import WorkflowNodeWrapper, { WorkflowNodeProps } from "components/workflow/nodeWrapper/WorkflowNodeWrapper";
import { NodeColorVariant } from "components/workflow/nodes/nodeVariants";
import { Form, Input, InputRow, InputSizeVariant, Select } from "components/forms/forms";
import { NumberFormatValues } from "react-number-format";
import Button, { ColorVariant, SizeVariant } from "components/buttons/Button";
import { useUpdateNodeMutation } from "pages/WorkflowPage/workflowApi";
import Spinny from "features/spinny/Spinny";

type TimeTriggerNodeProps = Omit<WorkflowNodeProps, "colorVariant">;

enum timeUnits {
  seconds = 1,
  minutes = 60,
  hours = 3600,
}

type timeUnitOption = { value: number; label: string };

// A function for checking that the option passed into the handle select function is a timeUnitOptions
function isTimeUnit(option: unknown): option is timeUnitOption {
  return (option as timeUnitOption).value !== undefined;
}

// function that converts between time units
function convertTimeUnits(from: timeUnits, to: timeUnits, value: number): number {
  return (value * from) / to;
}

const ONE_HOUR = 60 * 60; // 1 hour

type TimeTriggerParameters = {
  seconds: number;
  timeUnit: timeUnits;
};

// Function for checking if the parameters passed into the node are of type TimeTriggerParameters
function isTimeTriggerParameters(parameters: unknown): parameters is TimeTriggerParameters {
  const p = parameters as TimeTriggerParameters;
  return p.seconds !== undefined;
}

export function TimeTriggerNode({ ...wrapperProps }: TimeTriggerNodeProps) {
  const { t } = useTranslations();

  const [updateNode] = useUpdateNodeMutation();

  const parameters = isTimeTriggerParameters(wrapperProps.parameters)
    ? wrapperProps.parameters
    : { seconds: ONE_HOUR, timeUnit: timeUnits.hours };

  const [selectedTimeUnit, setSelectedTimeUnit] = useState<timeUnits>(parameters.timeUnit || timeUnits.seconds);
  const [frequency, setFrequency] = useState<number>(
    convertTimeUnits(timeUnits.seconds, selectedTimeUnit, parameters.seconds)
  );
  const [seconds, setSeconds] = useState<number>(ONE_HOUR);

  const timeUnitOptions = [
    { value: timeUnits.seconds, label: t.seconds },
    { value: timeUnits.minutes, label: t.minutes },
    { value: timeUnits.hours, label: t.hours },
  ];

  const selectedOption = timeUnitOptions.find((option) => option.value === selectedTimeUnit);

  const [dirty, setDirty] = useState(false);
  const [processing, setProcessing] = useState(false);
  useEffect(() => {
    // if the original parameters are different from the current parameters, set dirty to true
    if (parameters.seconds !== seconds || parameters.timeUnit !== selectedTimeUnit) {
      setDirty(true);
    } else {
      setDirty(false);
    }
  }, [selectedOption, seconds, frequency, selectedTimeUnit, wrapperProps.parameters]);

  useEffect(() => {
    setSeconds(convertTimeUnits(selectedTimeUnit, timeUnits.seconds, frequency));
  }, [frequency, selectedTimeUnit]);

  function handleFrequencyChange(values: NumberFormatValues) {
    const value = values.floatValue ? values.floatValue : 0;
    setFrequency(value);
  }

  function handleTimeUnitChange(newValue: unknown) {
    if (isTimeUnit(newValue)) {
      setFrequency(convertTimeUnits(selectedTimeUnit, newValue.value, frequency));
      setSelectedTimeUnit(newValue.value);
    }
  }

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setProcessing(true);
    updateNode({
      workflowVersionNodeId: wrapperProps.workflowVersionNodeId,
      parameters: {
        seconds: seconds,
        timeUnit: selectedTimeUnit,
      },
    }).finally(() => {
      setProcessing(false);
    });
  }

  return (
    <WorkflowNodeWrapper {...wrapperProps} headerIcon={<TimeTriggerIcon />} colorVariant={NodeColorVariant.accent2}>
      <Form onSubmit={handleSubmit}>
        <InputRow>
          <div style={{ flexGrow: 1 }}>
            <Input
              formatted={true}
              value={frequency}
              thousandSeparator={true}
              suffix={` ${selectedOption?.label}`}
              onValueChange={handleFrequencyChange}
              label={t.TriggerEvery}
              helpText={
                "Field sets the frequency of the trigger. For example 1 minute means the trigger will fire every minute."
              }
              sizeVariant={InputSizeVariant.small}
            />
          </div>
          <Select
            options={timeUnitOptions}
            onChange={handleTimeUnitChange}
            value={selectedOption}
            sizeVariant={InputSizeVariant.small}
          />
        </InputRow>
        <Button
          type="submit"
          buttonColor={ColorVariant.success}
          buttonSize={SizeVariant.small}
          icon={!processing ? <SaveIcon /> : <Spinny />}
          disabled={!dirty || processing}
        >
          {!processing ? t.save.toString() : t.saving.toString()}
        </Button>
      </Form>
    </WorkflowNodeWrapper>
  );
}
