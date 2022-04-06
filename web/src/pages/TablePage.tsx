import './table-page.scss'
import { useEffect } from 'react';
import { format } from "date-fns";
import { BarLoader } from 'react-spinners';
import FadeIn from 'react-fade-in';
import styled from '@emotion/styled';

import TableControls from "../components/table/TableControls";
import Table from "../components/table/Table";
import { useAppDispatch, useAppSelector } from "../store/hooks";
import { fetchChannelsAsync, selectStatus, } from "../components/table/tableSlice";
import { selectTimeInterval } from "../components/timeIntervalSelect/timeIntervalSlice";

const CenterCenter = styled.div`
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100vh
`

function TablePage() {
  const dispatch = useAppDispatch();
  const currentPeriod = useAppSelector(selectTimeInterval);
  const status = useAppSelector(selectStatus);

  useEffect(() => {
    const from = format(new Date(currentPeriod.from), "yyyy-MM-dd");
    const to = format(new Date(currentPeriod.to), "yyyy-MM-dd");
    dispatch(fetchChannelsAsync({ from: from, to: to }));
  }, [currentPeriod])

  if (status === "loading") {
    return (
      <CenterCenter>
        <BarLoader color='#B8EDE3' loading={true} height={10} width={100} />
      </CenterCenter>
    )
  }

  return (
    <FadeIn>
      <div className="table-page-wrapper">
        <div className="table-controls-wrapper">
          <TableControls />
        </div>
        <Table />
      </div>
    </FadeIn>
  );
}

export default TablePage;
