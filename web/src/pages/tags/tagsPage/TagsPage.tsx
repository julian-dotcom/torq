import Table from "features/table/Table";
import { Link, useNavigate } from "react-router-dom";
import { Tag20Regular as NewTagIcon } from "@fluentui/react-icons";
import TablePageTemplate, {
  TableControlSection,
  TableControlsButtonGroup,
  TableControlsTabsGroup,
} from "features/templates/tablePageTemplate/TablePageTemplate";
import Button, { ColorVariant } from "components/buttons/Button";
import { useLocation } from "react-router";
import useTranslations from "services/i18n/useTranslations";
import { useGetTagsQuery } from "../tagsApi";
import { DefaultTagsView } from "./tagsDefaults";
import tagsCellRenderer from "./tagsCellRenderer";
import mixpanel from "mixpanel-browser";

function TagsPage() {
  const { t } = useTranslations();
  const navigate = useNavigate();
  const location = useLocation();

  const tagsResponse = useGetTagsQuery();

  const tableControls = (
    <TableControlSection>
      <TableControlsButtonGroup>
        <TableControlsTabsGroup>
          <Button
            buttonColor={ColorVariant.primary}
            icon={<NewTagIcon />}
            hideMobileText={true}
            onClick={() => {
              mixpanel.track("Navigate to Create Tag");
              navigate("/create-tag", { state: { background: location } });
            }}
          >
            {t.tagsModal.createTag}
          </Button>
        </TableControlsTabsGroup>
      </TableControlsButtonGroup>
    </TableControlSection>
  );

  const breadcrumbs = [
    <span key="b1">{t.manage}</span>,
    <Link key="b2" to={`/${t.manage}/${t.tags}`}>
      {t.tags}
    </Link>,
  ];

  return (
    <TablePageTemplate title={t.tags} breadcrumbs={breadcrumbs} tableControls={tableControls}>
      <Table
        cellRenderer={tagsCellRenderer}
        data={tagsResponse?.data || []}
        activeColumns={DefaultTagsView.view.columns}
        isLoading={tagsResponse.isLoading || tagsResponse.isFetching || tagsResponse.isUninitialized}
      />
    </TablePageTemplate>
  );
}

export default TagsPage;
