import { channel } from "features/channels/channelsTypes";
import { Forward } from "features/forwards/forwardsTypes";
import { ChannelClosed } from "features/channelsClosed/channelsClosedTypes";
import { ChannelPending } from "features/channelsPending/channelsPendingTypes";
export function createCsvFile<T extends ChannelClosed | ChannelPending | channel | Forward>(
  data: Array<T>,
  filename: string
): void {
  // If there is no data, do not download anything
  if (data.length === 0) {
    return;
  }

  // Extract headers from the first object in the array
  const headers = Object.keys(data[0]);

  // Convert objects to arrays of values
  const rows = data.map((obj: T) =>
    headers.map((header) => {
      if (header === "tags" && obj[header]) {
        return `"${obj[header].map((tag) => tag.name).join(",")}"`;
      }
      return obj[header as keyof T] ?? "";
    })
  );

  // Combine headers and rows into a single array
  const csvData = [headers, ...rows];

  // Convert the array to CSV format
  const csv = csvData.map((row) => row.join(",")).join("\n");

  // Create a blob from the CSV string
  const blob = new Blob([csv], { type: "text/csv;charset=utf-8;" });

  // Create a URL for the blob
  const url = URL.createObjectURL(blob);

  // Create a link element to trigger the download
  const link = document.createElement("a");
  link.setAttribute("href", url);
  link.setAttribute("download", filename);

  // Append the link to the body and trigger the download
  document.body.appendChild(link);
  link.click();

  // Clean up the link and URL
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}
