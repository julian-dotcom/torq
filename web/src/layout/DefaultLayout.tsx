import { Outlet } from "react-router-dom";
import { useAppSelector } from "../store/hooks";
import { selectHidden } from "../features/navigation/navSlice";
import styles from "./default-layout.module.scss";
import navStyles from "../features/navigation/nav.module.scss";

import Navigation from "../features/navigation/Navigation";
import TopNavigation from "../features/navigation/TopNavigation";
import classNames from "classnames";

function DefaultLayout() {
  const hidden = useAppSelector(selectHidden);
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
