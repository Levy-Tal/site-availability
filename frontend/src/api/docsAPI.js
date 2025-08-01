export const fetchDocs = async () => {
  try {
    const response = await fetch("/api/docs", {
      credentials: "include",
    });
    if (!response.ok) {
      throw new Error(`HTTP error! Status: ${response.status}`);
    }
    const contentType = response.headers.get("content-type");
    if (!contentType || !contentType.includes("application/json")) {
      throw new Error("Received non-JSON response");
    }
    return await response.json();
  } catch (error) {
    console.error("Error fetching docs:", error);
    throw error;
  }
};
