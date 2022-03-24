import React from 'react';
import TableControls from "../components/table/TableControls";
import Table from "../components/table/Table";
import './table-page.scss'

function TablePage() {
    return (
      <div className="table-page-wrapper">
        <div className="table-controls-wrapper">
          <TableControls/>
        </div>
        <Table/>
      </div>
    );
}

export default TablePage;
