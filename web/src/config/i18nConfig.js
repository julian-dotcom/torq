import { getStaticEndpoint } from 'utils/apiUrlBuilder'

export const defaultLang = "en";

// When adding a language also add it to web/src/features/settings/SettingsPage.tsx
export const supportedLangs = {
  en: "English",
  nl: "Nederlands",
};

export const langUrl = getStaticEndpoint() + "/locales/{lang}.json";
