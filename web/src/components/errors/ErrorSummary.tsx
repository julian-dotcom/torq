import Note, { NoteType } from "features/note/Note";
import { ErrorCircle20Regular as ErrorIcon } from "@fluentui/react-icons";
import { FormErrors, replaceMessageMergeTags } from "./errors";
import useTranslations from "services/i18n/useTranslations";

type errorSummaryType = {
  title?: string;
  errors: FormErrors;
};

const ErrorSummary = ({ errors, title }: errorSummaryType) => {
  const { t } = useTranslations();
  const serverErrors = [] as string[];
  if (errors && errors.server) {
    for (const error of errors.server) {
      if (error.code) {
        const translatedError = t.errors[error.code];
        const mergedError = replaceMessageMergeTags(translatedError, error.attributes);
        serverErrors.push(mergedError);
        continue;
      }
      // A bit of a hack to only show the top most error rather than all the wrapped errors
      serverErrors.push(error.description ?? "");
    }
  }
  return (
    <>
      {(errors.server || errors.fields) && (
        <Note icon={<ErrorIcon />} title={title ?? "Error"} noteType={NoteType.error}>
          {serverErrors &&
            serverErrors.map((error, index) => {
              return <p key={index}>{error}</p>;
            })}
          {errors.fields && <p>See form error{Object.keys(errors.fields).length > 1 && <>s</>}</p>}
        </Note>
      )}
    </>
  );
};

export default ErrorSummary;
