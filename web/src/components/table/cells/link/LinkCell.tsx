import { Link16Regular as LinkIcon } from "@fluentui/react-icons";
import classNames from "classnames";
import mixpanel from "mixpanel-browser";
import { ColorVariant, LinkButton, SizeVariant } from "components/buttons/Button";
import styles from "components/table/cells/cell.module.scss";

interface LinkCell {
  text: string;
  link: string;
  className?: string;
  totalCell?: boolean;
}

function LinkCell(props: LinkCell) {
  return (
    <div className={classNames(styles.cell, styles.numericCell, styles.linkCell, props.className)}>
      {!props.totalCell && (
        <div className={classNames(styles.current, styles.text, styles.link)}>
          <LinkButton
            rel="noreferrer"
            to={props.link}
            target="_blank"
            onClick={() => {
              mixpanel.track("Link Cell Clicked", { href: props.link });
            }}
            buttonSize={SizeVariant.tiny}
            buttonColor={ColorVariant.success}
            icon={<LinkIcon />}
          >
            {props.text}
          </LinkButton>
        </div>
      )}
    </div>
  );
}
export default LinkCell;
