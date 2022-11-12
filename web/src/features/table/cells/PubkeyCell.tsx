import classNames from "classnames";
import { Copy16Regular, Eye12Regular, Eye24Regular } from "@fluentui/react-icons";
import styles from "./cell.module.scss";
import { copyToClipboard } from "utils/copyToClipboard"

type Props = {
  pubKey: string
  className?: string
}

const PubkeyCell = ({ pubKey, className }: Props) => {
  return (
    <div className={classNames(styles.cell, styles.pubKeyCell, className)}>
      <button className={classNames(styles.action, styles.view)}>
        <Eye24Regular />
        <span className={classNames(styles.content)}>{pubKey}</span>
      </button>
      <button
        className={classNames(styles.action, styles.copy)}
        onClick={() => copyToClipboard(pubKey)}
      >
        <Copy16Regular />
        Copy
      </button>
      <button className={classNames(styles.action, styles.link)}>
        <Eye12Regular />
        Link
      </button>
    </div>
  )
}

export default PubkeyCell;