import Note, { NoteType } from "features/note/Note";
import { ErrorCircle20Regular as ErrorIcon } from "@fluentui/react-icons";

type errorSummaryType = {
  title?: string;
  errors: Array<string>;
  hasFieldValidationError: boolean;
};

const ErrorSummary = ({ errors, title, hasFieldValidationError }: errorSummaryType) => {
  return (
    errors && (
      <Note icon={<ErrorIcon />} title={title ?? "Error"} noteType={NoteType.error}>
        {errors.map((error, index) => (
          <p key={index}>{error}</p>
        ))}
        {hasFieldValidationError && <p>See above for form error(s)</p>}
      </Note>
    )
  );
};

export default ErrorSummary;
