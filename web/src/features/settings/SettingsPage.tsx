import { Save20Regular as SaveIcon } from "@fluentui/react-icons";
import Page from "layout/Page";
import Box from "./Box";
import style from "./settings.module.css";
import Select, { SelectOption } from "../forms/Select";
import SubmitButton from "../forms/SubmitButton";
import React from "react";
import { defaultStaticRanges } from "../timeIntervalSelect/customRanges";

interface settingsProps {
  name: string;
}

function Settings(props: settingsProps) {
  const defaultDateRangeLabels: {
    label: string;
    code: string;
  }[] = defaultStaticRanges;

  const defaultDateRangeOptions: SelectOption[] = defaultDateRangeLabels.map(
    dsr => ({ value: dsr.code, label: dsr.label })
  );

  const handleDefaultDateRangeChange = (combiner: any) => {
    console.log(combiner.value);
  };

  let preferredTimezoneOptions: SelectOption[] = [];
  for (let i = -11; i <= 12; i++) {
    let label = "UTC";
    if (i < 0) {
      label += " " + i;
    }
    if (i > 0) {
      label += " +" + i;
    }
    preferredTimezoneOptions.push({ label: label, value: i.toString() });
  }

  const submitPreferences = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    console.log("submitted");
  };

  return (
    <Page>
      <React.Fragment>
        <div>
          <div className={style.center}>
            <Box minWidth={440} title="Date & time settings">
              <form onSubmit={submitPreferences}>
                <Select
                  label="Default date range"
                  onChange={handleDefaultDateRangeChange}
                  options={defaultDateRangeOptions}
                  value={defaultDateRangeOptions.find(
                    dd => dd.value === "Last 7 Days"
                  )}
                />

                <Select
                  label="Preferred timezone"
                  onChange={handleDefaultDateRangeChange}
                  options={preferredTimezoneOptions}
                  value={preferredTimezoneOptions.find(tz => tz.value === "0")}
                />

                <Select
                  label="Week starts on"
                  onChange={handleDefaultDateRangeChange}
                  options={defaultDateRangeOptions}
                  value={defaultDateRangeOptions.find(
                    dd => dd.value === "Last 7 Days"
                  )}
                />
                <SubmitButton>
                  <React.Fragment>
                    <SaveIcon />
                    Save
                  </React.Fragment>
                </SubmitButton>
              </form>
            </Box>
          </div>
        </div>
      </React.Fragment>
    </Page>
  );
}

export default Settings;
