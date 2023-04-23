import React from "react";
import styles from "./note.module.scss";
import { Note24Regular as DefaultNoteIcon, Dismiss20Regular as DismissIcon } from "@fluentui/react-icons";
import classNames from "classnames";

export enum NoteType {
  info = "info",
  warning = "warning",
  error = "error",
  success = "success",
}

const noteTypeClasses = new Map<NoteType, string>([
  [NoteType.info, styles.info],
  [NoteType.success, styles.success],
  [NoteType.warning, styles.warning],
  [NoteType.error, styles.error],
]);

type NoteProps = {
  title: string;
  noteType?: NoteType;
  icon?: React.ReactNode;
  className?: string;
  children: React.ReactNode;
  dismissible?: boolean;
  intercomTarget?: string;
};

function Note(props: NoteProps) {
  const [dismissed, setDismissed] = React.useState(false);
  if (dismissed) return null;
  return (
    <div
      className={classNames(styles.noteWrapper, noteTypeClasses.get(props.noteType || NoteType.info), props.className)}
      data-intercom-target={props.intercomTarget}
    >
      <div className={styles.noteTitleWrapper}>
        <div className={styles.noteTitleIcon}>{props.icon || <DefaultNoteIcon />}</div>
        <div className={styles.noteTitle}>{props.title}</div>
        {props.dismissible && (
          <div className={styles.noteDismissIcon} onClick={() => setDismissed(true)}>
            {<DismissIcon />}
          </div>
        )}
      </div>
      <div className={styles.noteContent}>{props.children}</div>
    </div>
  );
}

export default Note;
