import React from "react";
import styles from "./note.module.scss";
import { Note24Regular as DefaultNoteIcon } from "@fluentui/react-icons";
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
};

function Note(props: NoteProps) {
  return (
    <div
      className={classNames(styles.noteWrapper, noteTypeClasses.get(props.noteType || NoteType.info), props.className)}
    >
      <div className={styles.noteTitleWrapper}>
        <div className={styles.noteTitleIcon}>{props.icon || <DefaultNoteIcon />}</div>
        <div className={styles.noteTitle}>{props.title}</div>
      </div>
      <div className={styles.noteContent}>{props.children}</div>
    </div>
  );
}

export default Note;
