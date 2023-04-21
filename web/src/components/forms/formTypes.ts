import { InputColorVaraint, InputSizeVariant } from "./input/variants";

export type BasicInputType = {
  label?: string;
  formatted?: never;
  sizeVariant?: InputSizeVariant;
  colorVariant?: InputColorVaraint;
  leftIcon?: React.ReactNode;
  errorText?: string;
  warningText?: string;
  helpText?: string;
  intercomTarget?: string;
};
