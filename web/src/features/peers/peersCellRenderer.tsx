import { ColumnMetaData } from "features/table/types";
import { Peer } from "features/peers/peersTypes";
import DefaultCellRenderer from "features/table/DefaultCellRenderer";
import CellWrapper from "components/table/cells/cellWrapper/CellWrapper";
import Button, { ColorVariant, SizeVariant } from "../../components/buttons/Button";
import mixpanel from "mixpanel-browser";
import useTranslations from "services/i18n/useTranslations";
import styles from "features/peers/peers.module.scss";
export default function peerCellRenderer(
  row: Peer,
  rowIndex: number,
  column: ColumnMetaData<Peer>,
  columnIndex: number,
  isTotalsRow?: boolean,
  maxRow?: Peer
): JSX.Element {
  const { t } = useTranslations();

  if (column.key === "actions") {
    return (
      <CellWrapper cellWrapperClassName={styles.actionsWrapper} key={"connect-button-" + row.nodeId}>
        <Button
          // disabled={!(row.tagId !== undefined && row.tagId >= 0)}
          buttonSize={SizeVariant.small}
          onClick={() => {
            mixpanel.track("Connect Node", {
              tagId: row.nodeId,
            });
            //deleteTagMutation(row.tagId || 0);
          }}
          buttonColor={ColorVariant.success}
        >
          {t.peersPage.connect}
        </Button>
        <Button
          // disabled={!(row.tagId !== undefined && row.tagId >= 0)}
          buttonSize={SizeVariant.small}
          onClick={() => {
            mixpanel.track("Disconnect Node", {
              tagId: row.nodeId,
            });
            //deleteTagMutation(row.tagId || 0);
          }}
          buttonColor={ColorVariant.error}
        >
          {t.peersPage.disconnect}
        </Button>
      </CellWrapper>
    );
  }

  return DefaultCellRenderer(row, rowIndex, column, columnIndex, false, maxRow);
}
