import { Navigation20Regular as NavigationIcon } from "@fluentui/react-icons";
import Button from "features/buttons/Button";
import { useAppDispatch } from "store/hooks";
import { toggleNav } from "features/navigation/navSlice";
import { ReactElement } from "react";
import styles from "./page.module.scss";

interface pageProps {
  children: ReactElement;
  head?: ReactElement;
}

function Page(props: pageProps) {
  const dispatch = useAppDispatch();
  return (
    <div className={styles.page}>
      <div className={styles.headContainer}>
        <Button
          icon={<NavigationIcon />}
          text={"Menu"}
          onClick={() => dispatch(toggleNav())}
          className={"show-nav-btn collapse-tablet"}
        />
        {props.head}
      </div>
      {props.children}
    </div>
  );
}

export default Page;
