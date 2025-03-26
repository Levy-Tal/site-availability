export const fetchAppStatuses = async () => {
    try {
      const response = await fetch("/api/status");
      if (!response.ok) {
        throw new Error(`HTTP error! Status: ${response.status}`);
      }
      const contentType = response.headers.get("content-type");
      if (!contentType || !contentType.includes("application/json")) {
        throw new Error("Received non-JSON response");
      }
      return await response.json();
    } catch (error) {
      console.error("Error fetching app statuses:", error);
      return { locations: [], apps: [] }; // Return empty data instead of null
    }
  };