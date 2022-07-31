
const nonSummableFields: Array<string> = ["alias", "pub_key", "color"]
const arrayAggKeys: Array<string> = ["channelDbId", "channel_point", "shortChannelId", "chan_id"]

export const groupByFn = (channels: Array<any>, by: string) => {

  if (by !== 'peers') {
    return channels
  }

  const summedPubKey: typeof channels = []

  for (const chan of channels) {
    const pub_key = String(chan["pub_key" as keyof typeof chan]);

    const summedChan = summedPubKey.find(sc => sc["pub_key" as keyof typeof sc] == pub_key)
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

        // If the previous result is not already an Array, create a new one
        if (!Array.isArray(summedChan[key as keyof typeof summedChan])) {
          (summedChan as { [key: string]: any })[key] = [summedChan[key as keyof typeof summedChan], value]
          continue
        }

        (summedChan as { [key: string]: any })[key] = [...summedChan[key as keyof typeof summedChan], value]
        continue
      }

      (summedChan as { [key: string]: any })[key] = summedChan[key as keyof typeof summedChan] as number + value
    }
  }

  return summedPubKey

}

