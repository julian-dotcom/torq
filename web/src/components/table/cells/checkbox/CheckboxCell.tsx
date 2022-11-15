import { CheckboxChecked16Regular, CheckboxUnchecked16Regular } from "@fluentui/react-icons";
import classNames from "classnames";
import cellStyles from "components/table/cells/cell.module.scss";
import styles from "./checkbox_cell.module.scss";

export interface CheckboxCellProps
  extends React.DetailedHTMLProps<React.InputHTMLAttributes<HTMLInputElement>, HTMLInputElement> {
  wrapperClassNames?: string;
  totalCell?: boolean;
}

const CheckboxCell = ({ wrapperClassNames, totalCell, ...rest }: CheckboxCellProps) => {
  return (
    <div
      className={classNames(
        cellStyles.cell,
        styles.checkboxCell,
        { [cellStyles.totalCell]: totalCell },
        wrapperClassNames
      )}
    >
      <label className={classNames(styles.wrapper, { [styles.checked]: rest.checked })}>
        <input {...rest} type="checkbox" />
        {rest.checked ? <CheckboxChecked16Regular /> : <CheckboxUnchecked16Regular />}
      </label>
    </div>
  );
};

export default CheckboxCell;
