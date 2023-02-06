import { langUrl } from "config/i18nConfig";
const noCacheHeader = new Headers();
noCacheHeader.append("pragma", "no-cache");
noCacheHeader.append("cache-control", "no-cache");

export function fetchTranslations(lang) {
  return new Promise((resolve) => {
    fetch(langUrl.replace("{lang}", lang), { method: "get", headers: noCacheHeader })
      .then((response) => response.json())
      .then((data) => resolve(data));
  });
}
