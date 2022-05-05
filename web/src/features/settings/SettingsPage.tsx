import Page from "layout/Page"

interface settingsProps {
  name: string
}

function Settings(props: settingsProps) {
  return (
    <Page>
      <h1>Hello I am the settings page, {props.name}</h1>
    </Page>
  )
}

export default Settings
