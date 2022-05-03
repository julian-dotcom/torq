import React from "react";
import { useAppDispatch } from "../../store/hooks";
import { toggleNav } from "./navSlice";
import MenuItem from "./MenuItem";
import { ReactComponent as DotIcon } from "../../icons/dot-solid.svg";
import { ReactComponent as TorqLogo } from "../../icons/torq-logo.svg";
import {
  ColumnTriple20Regular as TableIcon,
  ChevronDoubleLeft20Regular as CollapseIcon,
  VirtualNetwork20Regular as ChannelIcon,
  LockClosed20Regular as LogoutIcon,
  // Money20Regular as FeeIcon,
  // AddSquare20Regular as AddTable,
  // ArrowRepeatAll20Regular as RebalanceIcon,
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
        <div
          className="collapse icon-button"
          onClick={() => dispatch(toggleNav())}
        >
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
          <MenuItem text={"Tables"} icon={<TableIcon />} routeTo={"/"} />
          <MenuItem
            text={"Channels"}
            icon={<ChannelIcon />}
            routeTo={"/channel"}
          />
          {/*<MenuItem text={'Fees'} icon={<FeeIcon/>}  />*/}
          {/*<MenuItem text={'Rebalance'} icon={<RebalanceIcon/>}  />*/}
        </div>
      </div>

      <div className="bottom-wrapper">
        <MenuItem text={"Logout"} icon={<LogoutIcon />} routeTo={"/logout"} />
      </div>
    </div>
  );
}

export default Navigation;
