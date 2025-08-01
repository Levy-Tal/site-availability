export const fetchLocations = async (statusFilters = [], labelFilters = []) => {
  try {
    const params = new URLSearchParams();

    if (Array.isArray(statusFilters) && statusFilters.length > 0) {
      statusFilters.forEach((status) => {
        params.append("status", status);
      });
    }

    labelFilters.forEach((label) => {
      params.append(`labels.${label.key}`, label.value);
    });

    const queryString = params.toString();
    const url = queryString
      ? `/api/locations?${queryString}`
      : "/api/locations";

    const response = await fetch(url, {
      credentials: "include",
    });
    if (!response.ok) {
      throw new Error(`HTTP error! Status: ${response.status}`);
    }
    const contentType = response.headers.get("content-type");
    if (!contentType || !contentType.includes("application/json")) {
      throw new Error("Received non-JSON response");
    }
    const data = await response.json();
    return Array.isArray(data) ? data : data.locations || [];
  } catch (error) {
    console.error("Error fetching locations:", error);
    return [];
  }
};

export const fetchApps = async (
  locationName = null,
  statusFilters = [],
  labelFilters = [],
) => {
  try {
    const params = new URLSearchParams();

    if (locationName) {
      params.append("location", locationName);
    }

    if (Array.isArray(statusFilters) && statusFilters.length > 0) {
      statusFilters.forEach((status) => {
        params.append("status", status);
      });
    }

    labelFilters.forEach((label) => {
      params.append(`labels.${label.key}`, label.value);
    });

    const queryString = params.toString();
    const url = queryString ? `/api/apps?${queryString}` : "/api/apps";

    const response = await fetch(url, {
      credentials: "include",
    });
    if (!response.ok) {
      throw new Error(`HTTP error! Status: ${response.status}`);
    }
    const contentType = response.headers.get("content-type");
    if (!contentType || !contentType.includes("application/json")) {
      throw new Error("Received non-JSON response");
    }
    const data = await response.json();
    return Array.isArray(data) ? data : data.apps || [];
  } catch (error) {
    console.error("Error fetching apps:", error);
    return [];
  }
};

export const fetchLabels = async (key = null) => {
  try {
    const url = key ? `/api/labels?${encodeURIComponent(key)}` : "/api/labels";
    const response = await fetch(url, {
      credentials: "include",
    });
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
