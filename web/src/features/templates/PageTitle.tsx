import React, { useState } from "react";
import classNames from "classnames";
import { Save28Regular as SaveTitleIcon, Edit28Regular as EditTitleIcon } from "@fluentui/react-icons";
import styles from "./templates.module.scss";
import Breadcrumbs from "features/breadcrumbs/Breadcrumbs";
import Button, { ColorVariant, SizeVariant } from "components/buttons/Button";

export type PageNameInputProps = React.DetailedHTMLProps<
  React.InputHTMLAttributes<HTMLInputElement>,
  HTMLInputElement
> & {
  title: string;
  className?: string;
  onSubmitHandler?: (name: string) => void;
  onHideInput: () => void;
  isDraft?: boolean;
};

function PageNameInput({ title, className, onSubmitHandler, onHideInput, ...inputProps }: PageNameInputProps) {
  const [localTitle, setLocalTitle] = useState(title);

  function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    onHideInput();
    onSubmitHandler && onSubmitHandler(localTitle);
  }

  return (
    <form className={styles.pageNameInputWrapper} onSubmit={handleSubmit}>
      <input
        {...inputProps}
        value={localTitle}
        className={classNames(className, styles.title, styles.pageNameInput)}
        onChange={(e) => setLocalTitle(e.target.value)}
      />
      <Button type={"submit"} buttonColor={ColorVariant.primary} buttonSize={SizeVariant.large}>
        <SaveTitleIcon />
      </Button>
    </form>
  );
}

type PageTitleProps = {
  title: string;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  breadcrumbs?: Array<any>;
  className?: string;
  children?: React.ReactNode;
  onNameChange?: (name: string) => void;
  isDraft?: boolean;
};

function PageTitle(props: PageTitleProps) {
  const [showNameInput, setShowNameInput] = React.useState(false);

  const title =
    props.onNameChange && props.isDraft ? (
      <span>{props.title}</span>
    ) : props.onNameChange && showNameInput ? (
      <PageNameInput
        title={props.title}
        onSubmitHandler={props.onNameChange}
        onHideInput={() => {
          setShowNameInput(false);
        }}
      />
    ) : (
      <>
        <span onDoubleClick={() => setShowNameInput(true)}>{props.title}</span>
        {props.onNameChange && (
          <Button
            buttonColor={ColorVariant.primary}
            buttonSize={SizeVariant.large}
            onClick={() => setShowNameInput(true)}
            className={styles.editNameButton}
          >
            <EditTitleIcon />
          </Button>
        )}
      </>
    );

  return (
    <div className={classNames(styles.pageTitleWrapper, props.className)}>
      <div className={styles.leftWrapper}>
        <Breadcrumbs breadcrumbs={props.breadcrumbs || []} />
        <h1 className={styles.titleContainer}>{title}</h1>
      </div>
      {props.children && <div className={styles.rightWrapper}>{props.children}</div>}
    </div>
  );
}

const memoizedPageTitle = React.memo(PageTitle);
export default memoizedPageTitle;
