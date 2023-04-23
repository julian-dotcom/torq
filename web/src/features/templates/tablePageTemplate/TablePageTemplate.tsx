import React from "react";
import styles from "./table-page-template.module.scss";

import classNames from "classnames";
import { FluentIconsProps } from "@fluentui/react-icons";
import PageTitle from "features/templates/PageTitle";
import Button, { ColorVariant } from "components/buttons/Button";

type TablePageTemplateProps = {
  title: string;
  titleContent?: React.ReactNode;
  sidebarExpanded?: boolean;
  sidebar?: React.ReactNode;
  pagination?: React.ReactNode;
  pageTotals?: React.ReactNode;
  tableControls?: React.ReactNode;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  breadcrumbs?: Array<any>;
  children?: React.ReactNode;
  onNameChange?: (title: string) => void;
  isDraft?: boolean;
};

export default function TablePageTemplate(props: TablePageTemplateProps) {
  return (
    <div className={classNames(styles.contentWrapper)}>
      <PageTitle
        breadcrumbs={props.breadcrumbs}
        title={props.title}
        onNameChange={props.onNameChange}
        isDraft={props.isDraft}
      >
        {props.titleContent}
      </PageTitle>

      {props.tableControls}

      <div className={styles.tableWrapper}>
        <div className={styles.tableContainer}>
          <div className={styles.tableExpander}>{props.children}</div>
        </div>
      </div>

      <div className={classNames(styles.paginationWrapper)}>{props.pagination}</div>

      <div className={classNames(styles.pageSidebarWrapper, { [styles.sidebarExpanded]: props.sidebarExpanded })}>
        {props.sidebar}
      </div>
    </div>
  );
}

type TableControlsButtonProps = {
  id?: string;
  text?: string;
  icon: React.FC<FluentIconsProps>;
  onClickHandler?: () => void;
  active?: boolean;
  intercomTarget?: string;
};

export function TableControlsButton(props: TableControlsButtonProps) {
  return (
    <div
      className={classNames(styles.tableControlsButtonWrapper, { [styles.active]: props.active })}
      data-intercom-target={props.intercomTarget}
    >
      <Button
        intercomTarget={"table-controls-button"}
        id={props.id}
        className={styles.tableControlsButtonIcon}
        onClick={props.onClickHandler}
        buttonColor={ColorVariant.primary}
        icon={<props.icon />}
      >
        {props.text}
      </Button>
    </div>
  );
}

export function TableControlSection(props: { intercomTarget?: string; children?: React.ReactNode }) {
  return (
    <div className={classNames(styles.tableControlsSection)} data-intercom-target={props.intercomTarget}>
      {props.children}
    </div>
  );
}

export function TableControlsButtonGroup(props: { intercomTarget?: string; children?: React.ReactNode }) {
  return (
    <div className={classNames(styles.tableControlsButtonGroup)} data-intercom-target={props.intercomTarget}>
      {props.children}
    </div>
  );
}

export function TableControlsTabsGroup(props: { children?: React.ReactNode }) {
  return <div className={classNames(styles.tableControlsTabsGroup)}>{props.children}</div>;
}
