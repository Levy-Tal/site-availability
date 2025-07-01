// frontend/src/api/appStatusAPI.js

// Fetch locations with their calculated status
export const fetchLocations = async (statusFilter = "", labelFilters = []) => {
  try {
    const params = new URLSearchParams();

    if (statusFilter) {
      params.append("status", statusFilter);
    }

    labelFilters.forEach((label) => {
      params.append(`labels.${label.key}`, label.value);
    });

    const queryString = params.toString();
    const url = queryString
      ? `/api/locations?${queryString}`
      : "/api/locations";

    const response = await fetch(url);
    if (!response.ok) {
      throw new Error(`HTTP error! Status: ${response.status}`);
    }
    const contentType = response.headers.get("content-type");
    if (!contentType || !contentType.includes("application/json")) {
      throw new Error("Received non-JSON response");
    }
    const data = await response.json();
    return data.locations || [];
  } catch (error) {
    console.error("Error fetching locations:", error);
    return [];
  }
};

// Fetch apps for a specific location or all apps
export const fetchApps = async (
  locationName = null,
  statusFilter = "",
  labelFilters = [],
) => {
  try {
    const params = new URLSearchParams();

    if (locationName) {
      params.append("location", locationName);
    }

    if (statusFilter) {
      params.append("status", statusFilter);
    }

    labelFilters.forEach((label) => {
      params.append(`labels.${label.key}`, label.value);
    });

    const queryString = params.toString();
    const url = queryString ? `/api/apps?${queryString}` : "/api/apps";

    const response = await fetch(url);
    if (!response.ok) {
      throw new Error(`HTTP error! Status: ${response.status}`);
    }
    const contentType = response.headers.get("content-type");
    if (!contentType || !contentType.includes("application/json")) {
      throw new Error("Received non-JSON response");
    }
    const data = await response.json();
    return data.apps || [];
  } catch (error) {
    console.error("Error fetching apps:", error);
    return [];
  }
};

// Fetch available labels
export const fetchLabels = async (key = null) => {
  try {
    const url = key ? `/api/labels?${encodeURIComponent(key)}` : "/api/labels";
    const response = await fetch(url);
    if (!response.ok) {
      throw new Error(`HTTP error! Status: ${response.status}`);
    }
    const contentType = response.headers.get("content-type");
    if (!contentType || !contentType.includes("application/json")) {
      throw new Error("Received non-JSON response");
    }
    const data = await response.json();
    return data.labels || [];
  } catch (error) {
    console.error("Error fetching labels:", error);
    return [];
  }
};
