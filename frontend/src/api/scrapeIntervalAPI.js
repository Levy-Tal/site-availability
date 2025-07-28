export const fetchScrapeInterval = async () => {
  try {
    const response = await fetch("/api/scrape-interval");
    if (!response.ok) {
      throw new Error(`HTTP error! Status: ${response.status}`);
    }
    const contentType = response.headers.get("content-type");
    if (!contentType || !contentType.includes("application/json")) {
      throw new Error("Received non-JSON response");
    }
    return await response.json();
  } catch (error) {
    console.error("Error fetching scrape interval:", error);
    throw error;
  }
};
