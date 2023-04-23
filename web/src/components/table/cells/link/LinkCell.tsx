import { Link16Regular as LinkIcon } from "@fluentui/react-icons";
import classNames from "classnames";
import { ColorVariant, ExternalLinkButton, SizeVariant } from "components/buttons/Button";
import styles from "components/table/cells/cell.module.scss";
import { userEvents } from "utils/userEvents";

interface LinkCell {
  text: string;
  link: string;
  className?: string;
  totalCell?: boolean;
}

function LinkCell(props: LinkCell) {
  const { track } = userEvents();
  return (
    <div className={classNames(styles.cell, styles.numericCell, styles.linkCell, props.className)}>
      {!props.totalCell && (
        <div className={classNames(styles.current, styles.text, styles.link)}>
          <ExternalLinkButton
            // Link to external site
            href={props.link}
            target="_blank"
            onClick={() => {
              track("Link Cell Clicked", { href: props.link });
            }}
            buttonSize={SizeVariant.tiny}
            buttonColor={ColorVariant.success}
            icon={<LinkIcon />}
            intercomTarget={"link-cell-button"}
          >
            {props.text}
          </ExternalLinkButton>
        </div>
      )}
    </div>
  );
}
export default LinkCell;
