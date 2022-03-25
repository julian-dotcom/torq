import React from 'react';
import { useSelector, useDispatch } from "react-redux";
import MenuItem from './MenuItem'
import {ReactComponent as DotIcon} from '../../icons/dot-solid.svg'
import {ReactComponent as TorqLogo} from '../../icons/torq-logo.svg'
import {
  Money20Regular as FeeIcon,
  ColumnTriple20Regular as TableIcon,
  AddSquare20Regular as AddTable,
  ArrowRepeatAll20Regular as RebalanceIcon,
  ChevronDoubleLeft20Regular as CollapseIcon,
} from "@fluentui/react-icons";
import './navigation.scss'


function Navigation() {
  const dispatch = useDispatch();
  const navHidden: number = useSelector((state:{navHidden:number}) => {return state.navHidden});

  const toggleNav = () => {
    dispatch({type: 'toggleNav'})
  }

  return (
<div className="navigation">
  <div className="logo-wrapper">
    <div className="logo"><TorqLogo/></div>
    <div className="collapse icon-button" onClick={toggleNav}>
      <CollapseIcon/>
    </div>
  </div>

  <MenuItem text={'My Routing Node'} />

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

