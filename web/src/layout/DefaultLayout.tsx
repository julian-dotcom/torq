import React from "react";
import { Outlet } from "react-router-dom";
import { useAppSelector, useAppDispatch } from "../store/hooks";
import { selectHidden } from "../features/navigation/navSlice";
import styles from "./default-layout.module.scss";
import navStyles from "../features/navigation/nav.module.scss";

import Navigation from "../features/navigation/Navigation";
import TopNavigation from "../features/navigation/TopNavigation";
import classNames from "classnames";

function DefaultLayout() {
  const hidden = useAppSelector(selectHidden);
  const dispatch = useAppDispatch();
  return (
    <div className={classNames(styles.mainContentWrapper, { [navStyles.navCollapsed]: hidden })}>
      <TopNavigation />
      <div className={navStyles.navigationWrapper}>
        <Navigation />
      </div>
      <div className={styles.pageWrapper}>
        {/*<div className="dismiss-navigation-background" onClick={() => dispatch(toggleNav())} />*/}
        <Outlet />
      </div>
    </div>
  );
}

export default DefaultLayout;
