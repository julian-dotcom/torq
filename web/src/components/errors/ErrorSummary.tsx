import Note, { NoteType } from "features/note/Note";
import { ErrorCircle20Regular as ErrorIcon } from "@fluentui/react-icons";
import { FormErrors } from "./errors";
import useTranslations from "services/i18n/useTranslations";

type errorSummaryType = {
  title?: string;
  errors: FormErrors;
};

const ErrorSummary = ({ errors, title }: errorSummaryType) => {
  const { t } = useTranslations();
  if (errors && errors.server) {
    // Bit of a hack to only show the top most error rather than all the wrapped errors
    errors.server = errors.server.map((error) => {
      return { code: error.code, description: (error.description ?? "").split(":")[0] };
    });
  }
  return (
    (errors.server || errors.fields) && (
      <Note icon={<ErrorIcon />} title={title ?? "Error"} noteType={NoteType.error}>
        {errors.server &&
          errors.server.map((error, index) => {
            if (error.code) {
              return <p key={index}>{t.errors[error.code]}</p>;
            }
            return <p key={index}>{error.description}</p>;
          })}
        {errors.fields && errors.fields.keys() && <p>See above for form error(s)</p>}
      </Note>
    )
  );
};

export default ErrorSummary;
