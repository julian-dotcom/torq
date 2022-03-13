import React from 'react';
import './navigation.scss'
import MenuItem from './MenuItem'
import {ReactComponent as DotIcon} from '../../icons/dot-solid.svg'
import {ReactComponent as TorqLogo} from '../../icons/torq-logo.svg'
import {
  Money20Regular as FeeIcon,
  ColumnTriple20Regular as TableIcon,
  AddSquare20Regular as AddTable,
  ArrowRepeatAll20Regular as RebalanceIcon,
  ArrowExportRtl20Regular as CollapseIcon,
  TextBulletListCheckmark20Regular as SelectNodeIcon,
} from "@fluentui/react-icons";

function Navigation() {
  return (
<div className="navigation">
  <div className="logo-wrapper">
    <div className="logo"><TorqLogo/></div>
    <div className="collapse icon-button">
      <CollapseIcon/>
    </div>
  </div>

  <MenuItem text={'Routing Node 2'} actions={<SelectNodeIcon/>}/>

  <div className="menu-items">
    <MenuItem text={'Tables'} icon={<TableIcon/>} actions={<AddTable/>}>
      <MenuItem text={'Top revenue today'} icon={<DotIcon/>} selected={true} routeTo={'/a'} />
      <MenuItem text={'Source channels'} icon={<DotIcon/>} routeTo={'/a'} />
      <MenuItem text={'Destination channels'} icon={<DotIcon/>} routeTo={'/a'} />
    </MenuItem>

    <div className="wrapper">
      <MenuItem text={'Fees'} icon={<FeeIcon/>} routeTo={'/a'} />
      <MenuItem text={'Rebalance'} icon={<RebalanceIcon/>} routeTo={'/a'} />
    </div>

  </div>


</div>
  );
}

export default Navigation;


