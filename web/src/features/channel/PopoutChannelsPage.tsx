import { MoneyHand24Regular as TransactionIconModal } from "@fluentui/react-icons";
import PopoutPageTemplate from "features/templates/popoutPageTemplate/PopoutPageTemplate";
import ChannelPage from "features/channel/ChannelPage";

type PopoutChannelsPageProps = {
  show: boolean;
  modalCloseHandler: () => void;
};

function PopoutChannelsPage(props: PopoutChannelsPageProps) {
  return (
    <PopoutPageTemplate
      title={"Channel"}
      show={props.show}
      onClose={props.modalCloseHandler}
      icon={<TransactionIconModal />}
    >
      <ChannelPage />
    </PopoutPageTemplate>
  );
}

export default PopoutChannelsPage;
