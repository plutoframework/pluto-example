export function downloadJson(value: unknown, name: string): void {
  const url = URL.createObjectURL(
    new Blob([JSON.stringify(value, null, 2)], { type: "application/json" }),
  );

  const link = document.createElement("a");
  link.href = url;
  link.download = name;
  link.click();
  // Leave the URL alive for the download to start, then release its backing data.
  setTimeout(() => URL.revokeObjectURL(url), 1000);
}
