import React, { useRef, useEffect, useState, useCallback } from "react";
import {
  FaSortAmountDownAlt,
  FaSearch,
  FaChevronDown,
  FaChevronUp,
} from "react-icons/fa";
import { fetchApps, fetchLabels } from "./api/appStatusAPI";
import { userPreferences } from "./utils/storage";

export const AppStatusPanel = ({
  site,
  onClose,
  scrapeInterval,
  statusFilters,
  labelFilters,
  refreshLocations,
}) => {
  const panelRef = useRef(null);
  const groupDropdownRef = useRef(null);
  const sortDropdownRef = useRef(null);
  const resizeHandleRef = useRef(null);

  const [apps, setApps] = useState([]);
  const [filteredApps, setFilteredApps] = useState([]);
  const [searchTerm, setSearchTerm] = useState("");
  const [sortOrder, setSortOrder] = useState(userPreferences.loadSortOrder());
  const [showSortOptions, setShowSortOptions] = useState(false);

  // Updated group by state
  const [showGroupOptions, setShowGroupOptions] = useState(false);
  const [availableLabels, setAvailableLabels] = useState([]);
  const [selectedGroupLabel, setSelectedGroupLabel] = useState(
    userPreferences.loadGroupByLabel(),
  );
  const [groupLabelInput, setGroupLabelInput] = useState(
    userPreferences.loadGroupByLabel() || "",
  );
  const [filteredLabels, setFilteredLabels] = useState([]);
  const [expandedGroups, setExpandedGroups] = useState(new Set());
  const [groupedApps, setGroupedApps] = useState({});

  // Fetch apps for this location and refresh locations simultaneously
  const refreshApps = useCallback(async () => {
    try {
      // Execute both API calls simultaneously
      const [appsData] = await Promise.all([
        fetchApps(site.name, statusFilters, labelFilters),
        refreshLocations ? refreshLocations() : Promise.resolve(),
      ]);
      setApps(appsData);
    } catch (error) {
      console.error("Error fetching apps for location:", error);
    }
  }, [site.name, statusFilters, labelFilters, refreshLocations]);

  // Initial fetch when panel opens - synchronized refresh
  useEffect(() => {
    refreshApps();
  }, [refreshApps]);

  // Load available labels
  const loadAvailableLabels = async () => {
    try {
      const labels = await fetchLabels();
      setAvailableLabels(labels);
      return labels;
    } catch (error) {
      console.error("Error loading labels:", error);
      return [];
    }
  };

  // Set up periodic refresh while panel is open
  useEffect(() => {
    if (scrapeInterval) {
      const intervalId = setInterval(() => {
        refreshApps();
      }, scrapeInterval);

      return () => clearInterval(intervalId);
    }
  }, [scrapeInterval, refreshApps]);

  // Close dropdowns on outside click
  useEffect(() => {
    const handleOutsideClick = (e) => {
      if (
        sortDropdownRef.current &&
        !sortDropdownRef.current.contains(e.target)
      ) {
        setShowSortOptions(false);
      }
      if (
        groupDropdownRef.current &&
        !groupDropdownRef.current.contains(e.target)
      ) {
        setShowGroupOptions(false);
      }
    };
    document.addEventListener("mousedown", handleOutsideClick);
    return () => document.removeEventListener("mousedown", handleOutsideClick);
  }, []);

  // Filter, sort, and group apps
  useEffect(() => {
    let filtered = apps.filter((app) => app.location === site.name);

    if (searchTerm) {
      const lower = searchTerm.toLowerCase();
      filtered = filtered.filter((app) =>
        app.name.toLowerCase().startsWith(lower),
      );
    }

    const sortMethods = {
      "name-asc": (a, b) => a.name.localeCompare(b.name),
      "name-desc": (a, b) => b.name.localeCompare(a.name),
      "status-up": (a, b) => a.status.localeCompare(b.status),
      "status-down": (a, b) => b.status.localeCompare(a.status),
    };

    filtered.sort(sortMethods[sortOrder] || (() => 0));
    setFilteredApps(filtered);

    // Group apps if a label is selected
    if (selectedGroupLabel) {
      const grouped = filtered.reduce((acc, app) => {
        // Find the label in the array format
        const label = app.labels?.find((l) => l.key === selectedGroupLabel);
        const labelValue = label?.value || "No Label";
        if (!acc[labelValue]) {
          acc[labelValue] = [];
        }
        acc[labelValue].push(app);
        return acc;
      }, {});
      setGroupedApps(grouped);
    }
  }, [apps, site, searchTerm, sortOrder, selectedGroupLabel]);

  const toggleGroup = (groupName) => {
    setExpandedGroups((prev) => {
      const newSet = new Set(prev);
      if (newSet.has(groupName)) {
        newSet.delete(groupName);
      } else {
        newSet.add(groupName);
      }
      return newSet;
    });
  };

  const getGroupStatus = (apps) => {
    if (apps.some((app) => app.status === "down")) return "down";
    if (apps.some((app) => app.status === "unavailable")) return "unavailable";
    return "up";
  };

  const getStatusCounts = (apps) => {
    return apps.reduce(
      (acc, app) => {
        acc[app.status]++;
        return acc;
      },
      { up: 0, down: 0, unavailable: 0 },
    );
  };

  const renderSortOption = (label, value) => (
    <li
      className={sortOrder === value ? "selected" : ""}
      onClick={() => {
        setSortOrder(value);
        setShowSortOptions(false);
      }}
    >
      <span>{label}</span>
    </li>
  );

  // Handle group label input changes
  const handleGroupLabelInput = (e) => {
    const value = e.target.value;
    setGroupLabelInput(value);

    if (value.length > 0) {
      const filtered = availableLabels.filter(
        (label) => label && label.toLowerCase().includes(value.toLowerCase()),
      );
      setFilteredLabels(filtered);
    } else {
      setFilteredLabels(availableLabels);
    }
  };

  // Toggle group options
  const toggleGroupOptions = async () => {
    if (showGroupOptions) {
      setShowGroupOptions(false);
    } else {
      if (availableLabels.length === 0) {
        const labels = await loadAvailableLabels();
        setFilteredLabels(labels);
      } else {
        setFilteredLabels(availableLabels);
      }
      setShowGroupOptions(true);
    }
  };

  // Handle group label selection
  const selectGroupLabel = (label) => {
    setSelectedGroupLabel(label);
    setGroupLabelInput(label || ""); // Clear input when label is null/undefined
    setShowGroupOptions(false);
  };

  // Handle group label input key press
  const handleGroupLabelKeyPress = (e) => {
    if (e.key === "Enter" && groupLabelInput.trim()) {
      const matchingLabel = availableLabels.find(
        (label) => label.toLowerCase() === groupLabelInput.toLowerCase(),
      );
      if (matchingLabel) {
        selectGroupLabel(matchingLabel);
      }
    }
  };

  // Clear input when no group label is selected
  useEffect(() => {
    if (!selectedGroupLabel) {
      setGroupLabelInput("");
    }
  }, [selectedGroupLabel]);

  // Save sort order whenever it changes
  useEffect(() => {
    userPreferences.saveSortOrder(sortOrder);
  }, [sortOrder]);

  // Save group by label whenever it changes
  useEffect(() => {
    userPreferences.saveGroupByLabel(selectedGroupLabel);
  }, [selectedGroupLabel]);

  // Update filtered labels when available labels change and dropdown is open
  useEffect(() => {
    if (showGroupOptions && availableLabels.length > 0) {
      setFilteredLabels(availableLabels);
    }
  }, [availableLabels, showGroupOptions]);

  // Resizable panel
  useEffect(() => {
    const panel = panelRef.current;
    const handle = resizeHandleRef.current;

    if (!handle || !panel) return;

    let isResizing = false;
    let startX = 0;
    let startWidth = 0;

    const onMouseDown = (e) => {
      isResizing = true;
      startX = e.clientX;
      startWidth = panel.getBoundingClientRect().width;

      document.body.style.cursor = "col-resize";
      document.body.style.userSelect = "none";
    };

    const onMouseMove = (e) => {
      if (!isResizing) return;

      const delta = startX - e.clientX;
      const newWidth = Math.max(300, Math.min(800, startWidth + delta));
      panel.style.width = `${newWidth}px`;
    };

    const onMouseUp = () => {
      isResizing = false;
      document.body.style.cursor = "";
      document.body.style.userSelect = "";
    };

    handle.addEventListener("mousedown", onMouseDown);
    document.addEventListener("mousemove", onMouseMove);
    document.addEventListener("mouseup", onMouseUp);

    return () => {
      handle.removeEventListener("mousedown", onMouseDown);
      document.removeEventListener("mousemove", onMouseMove);
      document.removeEventListener("mouseup", onMouseUp);
    };
  }, []);

  return (
    <div className="status-panel" ref={panelRef}>
      <div className="resize-handle" ref={resizeHandleRef} />
      <button
        className="close-button"
        onClick={() => typeof onClose === "function" && onClose()}
        aria-label="Close status panel"
      >
        ×
      </button>

      <h2>{site.name}</h2>

      <div className="search-sort-container">
        <div className="search-bar">
          <FaSearch className="search-icon" aria-hidden="true" />
          <input
            type="text"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="search-input"
            placeholder="Search apps..."
            aria-label="Search applications"
          />
        </div>

        <div className="group-dropdown" ref={groupDropdownRef}>
          <div className="group-input-container">
            <input
              type="text"
              className="group-input"
              placeholder="Group by label..."
              value={groupLabelInput}
              onChange={handleGroupLabelInput}
              onFocus={async () => {
                if (availableLabels.length === 0) {
                  const labels = await loadAvailableLabels();
                  setFilteredLabels(labels);
                } else {
                  setFilteredLabels(availableLabels);
                }
                setShowGroupOptions(true);
              }}
              onKeyPress={handleGroupLabelKeyPress}
            />
            <div className="group-input-arrow" onClick={toggleGroupOptions}>
              {showGroupOptions ? "▲" : "▼"}
            </div>
          </div>
          {showGroupOptions && (
            <ul className="group-options">
              <li
                className={!selectedGroupLabel ? "selected" : ""}
                onClick={() => selectGroupLabel(null)}
              >
                <span>None</span>
              </li>
              {filteredLabels.map((label) => (
                <li
                  key={label}
                  className={selectedGroupLabel === label ? "selected" : ""}
                  onClick={() => selectGroupLabel(label)}
                >
                  <span>{label}</span>
                </li>
              ))}
            </ul>
          )}
        </div>

        <div className="sort-dropdown" ref={sortDropdownRef}>
          <button
            className="sort-icon-button"
            onClick={() => setShowSortOptions(!showSortOptions)}
            aria-label="Sort options"
          >
            <FaSortAmountDownAlt className="sort-icon" aria-hidden="true" />
          </button>
          {showSortOptions && (
            <ul className="sort-options">
              {renderSortOption("Name A-Z", "name-asc")}
              {renderSortOption("Name Z-A", "name-desc")}
              {renderSortOption("Status (Down-Unavailable-Up)", "status-up")}
              {renderSortOption("Status (Up-Unavailable-Down)", "status-down")}
            </ul>
          )}
        </div>
      </div>

      {selectedGroupLabel ? (
        <div className="grouped-apps">
          {Object.entries(groupedApps).map(([groupName, groupApps]) => {
            const isExpanded = expandedGroups.has(groupName);
            const groupStatus = getGroupStatus(groupApps);
            const statusCounts = getStatusCounts(groupApps);

            return (
              <div key={groupName} className="group-tab">
                <div
                  className={`group-header ${groupStatus}`}
                  onClick={() => toggleGroup(groupName)}
                >
                  <div className="group-info">
                    <div className={`status-line ${groupStatus}`} />
                    <span className="group-name">{groupName}</span>
                  </div>
                  <div className="group-stats">
                    <div className="status-dots">
                      <span className="status-dot up">{statusCounts.up}</span>
                      <span className="status-dot unavailable">
                        {statusCounts.unavailable}
                      </span>
                      <span className="status-dot down">
                        {statusCounts.down}
                      </span>
                    </div>
                    {isExpanded ? <FaChevronUp /> : <FaChevronDown />}
                  </div>
                </div>
                {isExpanded && (
                  <ul className="group-apps">
                    {groupApps.map((app) => {
                      const statusClass =
                        app.status === "up"
                          ? "status-up"
                          : app.status === "down"
                            ? "status-down"
                            : "status-unavailable";

                      const label =
                        app.status === "up"
                          ? "Up"
                          : app.status === "down"
                            ? "Down"
                            : "Unavailable";

                      return (
                        <li key={app.name}>
                          <div className="app-name">{app.name}</div>
                          <div className={`status-indicator ${statusClass}`}>
                            {label}
                          </div>
                        </li>
                      );
                    })}
                  </ul>
                )}
              </div>
            );
          })}
        </div>
      ) : (
        <ul>
          {filteredApps.map((app) => {
            const statusClass =
              app.status === "up"
                ? "status-up"
                : app.status === "down"
                  ? "status-down"
                  : "status-unavailable";

            const label =
              app.status === "up"
                ? "Up"
                : app.status === "down"
                  ? "Down"
                  : "Unavailable";

            return (
              <li key={app.name}>
                <div className="app-name">{app.name}</div>
                <div className={`status-indicator ${statusClass}`}>{label}</div>
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
};
