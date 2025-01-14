import { GroupByOptions } from "features/viewManagement/types";

const nonSummableFields: Array<string> = ["alias", "pubKey", "color", "secondNodeId", "firstNodeId"];
const arrayAggKeys: Array<string> = [
  "channelId",
  "channelPoint",
  "shortChannelId",
  "lndShortChannelId",
  "tags",
  "channelTags",
  "peerTags",
];

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function useGroupBy<T>(data: Array<any>, by: GroupByOptions | undefined): Array<T> {
  if (by !== "peer") {
    return data;
  }

  const summedPubKey: typeof data = [];

  for (const chan of data) {
    const pub_key = String(chan["pubKey" as keyof typeof chan]);

    const summedChan = summedPubKey.find((sc) => sc["pubKey" as keyof typeof sc] == pub_key);
    if (!summedChan) {
      summedPubKey.push(chan);
      continue;
    }

    for (const key of Object.keys(chan)) {
      const value = chan[key as keyof typeof chan];

      if (nonSummableFields.includes(key)) {
        continue;
      }

      // Values fround in arrayAggKeys should be converted to an array of values
      if (arrayAggKeys.includes(key)) {
        let valueArr = [];
        if (Array.isArray(value)) {
          valueArr = [...value];
        } else {
          valueArr = [value];
        }

        // If the previous result is not already an Array, create a new one
        if (!Array.isArray(summedChan[key as keyof typeof summedChan])) {
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
          (summedChan as { [key: string]: any })[key] = [summedChan[key as keyof typeof summedChan], ...valueArr];
          continue;
        }

        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        (summedChan as { [key: string]: any })[key] = [...summedChan[key as keyof typeof summedChan], ...valueArr];
        continue;
      }

      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      (summedChan as { [key: string]: any })[key] = (summedChan[key as keyof typeof summedChan] as number) + value;
    }
  }

  return summedPubKey;
}
