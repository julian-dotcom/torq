import { FormErrors } from "components/errors/errors";
import Input, { FormattedInputProps, InputProps } from "components/forms/input/Input";
import useTranslations from "services/i18n/useTranslations";

export type ValidatedInputProps = {
  errors?: FormErrors;
};

const InputWithValidation = (props: (InputProps | FormattedInputProps) & ValidatedInputProps) => {
  const { t } = useTranslations();

  const inputProps = { ...props };
  // there is a field error for this field
  if (props.errors && props.errors.fields && props.name && props.errors.fields[props.name]) {
    const codeOrDescription = props.errors.fields[props.name][0];
    inputProps.errorText = t.errors[codeOrDescription.code] ?? codeOrDescription.description;
  }
  return <Input {...inputProps} />;
};

export default InputWithValidation;
