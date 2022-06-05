import React from "react";
import { useAppDispatch } from "../../store/hooks";
import { toggleNav } from "./navSlice";
import MenuItem from "./MenuItem";
import { ReactComponent as TorqLogo } from "../../icons/torq-logo.svg";
import {
  ColumnTriple20Regular as TableIcon,
  Navigation20Regular as CollapseIcon,
  Gauge20Regular as DashboardIcon,
  LockClosed20Regular as LogoutIcon,
  Settings20Regular as SettingsIcon,
} from "@fluentui/react-icons";
import "./navigation.scss";

function Navigation() {
  const dispatch = useAppDispatch();

  return (
    <div className="navigation">
      <div className="logo-wrapper">
        <div className="logo">
          <TorqLogo />
        </div>
        <div className="collapse icon-button" onClick={() => dispatch(toggleNav())}>
          <CollapseIcon />
        </div>
      </div>

      {/*<MenuItem text={'My Routing Node'} />*/}

      <div className="menu-items">
        {/*<MenuItem text={'Top revenue today'} icon={<DotIcon/>} selected={true} routeTo={'/a'} />*/}
        {/*<MenuItem text={'Source channels'} icon={<DotIcon/>} routeTo={'/a'} />*/}
        {/*<MenuItem text={'Destination channels'} icon={<DotIcon/>} routeTo={'/a'} />*/}

        <div className="wrapper">
          {/*actions={<AddTable/>}*/}
          <MenuItem text={"Dashboard"} icon={<DashboardIcon />} routeTo={"/"} />
          <MenuItem text={"Channels"} icon={<TableIcon />} routeTo={"/channels"} />
          {/*<MenuItem text={'Fees'} icon={<FeeIcon/>}  />*/}
          {/*<MenuItem text={'Rebalance'} icon={<RebalanceIcon/>}  />*/}
        </div>
      </div>

      <div className="bottom-wrapper">
        <MenuItem text={"Settings"} icon={<SettingsIcon />} routeTo={"/settings"} />
        <MenuItem text={"Logout"} icon={<LogoutIcon />} routeTo={"/logout"} />
      </div>
    </div>
  );
}

export default Navigation;
