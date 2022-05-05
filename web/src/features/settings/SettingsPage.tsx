import Page from "layout/Page";
import Box from "./Box";

interface settingsProps {
  name: string;
}

function Settings(props: settingsProps) {
  return (
    <Page>
      <Box maxWidth={400} title="Date & time settings">
        <h1>Hello I am the settings page, {props.name}</h1>
      </Box>
    </Page>
  );
}

export default Settings;
