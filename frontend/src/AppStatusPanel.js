import React, { useRef, useEffect, useState, useCallback } from "react";
import {
  FaSortAmountDownAlt,
  FaSearch,
  FaChevronDown,
  FaChevronUp,
  FaCheck,
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
  const showLabelsDropdownRef = useRef(null);
  const resizeHandleRef = useRef(null);

  const [panelWidth, setPanelWidth] = useState(400);

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

  // Show labels state
  const [showLabelOptions, setShowLabelOptions] = useState(false);
  const [selectedShowLabels, setSelectedShowLabels] = useState(
    userPreferences.loadShowLabels(),
  );

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
      if (
        showLabelsDropdownRef.current &&
        !showLabelsDropdownRef.current.contains(e.target)
      ) {
        setShowLabelOptions(false);
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

  // Function to render app labels
  const renderAppLabels = (app, containerWidth = 400) => {
    if (!selectedShowLabels.length || !app.labels) return null;

    const labelsToShow = app.labels
      .filter((label) => selectedShowLabels.includes(label.key))
      .sort((a, b) => a.key.localeCompare(b.key)); // Sort labels by name

    if (!labelsToShow.length) return null;

    // Calculate how many labels can fit based on container width
    // Estimate label width: 10px per character + 16px padding + 4px gap
    const estimateLabelWidth = (label) =>
      (label.key.length + label.value.length + 2) * 7 + 20; // +2 for ": "

    const availableWidth = Math.max(containerWidth - 200, 100); // Reserve space for name and status
    let totalWidth = 0;
    let maxVisibleLabels = 0;

    for (let i = 0; i < labelsToShow.length; i++) {
      const labelWidth = estimateLabelWidth(labelsToShow[i]);
      if (totalWidth + labelWidth > availableWidth) {
        break;
      }
      totalWidth += labelWidth + 4; // +4 for gap
      maxVisibleLabels++;
    }

    // Always show at least 1 label if there are any, but collapse if no space
    if (maxVisibleLabels === 0 && labelsToShow.length > 0) {
      return (
        <div className="app-labels">
          <div
            className="app-label-overflow-container"
            onMouseEnter={(e) => {
              const tooltip =
                e.currentTarget.querySelector(".app-label-tooltip");
              if (tooltip) {
                const rect = e.currentTarget.getBoundingClientRect();
                const leftPos = rect.left + rect.width / 2;
                const topPos = rect.bottom + 6;

                // PORTAL APPROACH: Move tooltip to document body to escape ALL clipping
                const tooltipLeft = leftPos - 100; // Center minus half tooltip width

                // Move to document body to escape any container clipping
                document.body.appendChild(tooltip);

                tooltip.style.position = "fixed";
                tooltip.style.left = `${tooltipLeft}px`;
                tooltip.style.top = `${topPos}px`;
                tooltip.style.transform = "none";
                tooltip.style.marginTop = "0";
                tooltip.style.zIndex = "2147483647";

                tooltip.classList.add("show");
              }
            }}
            onMouseLeave={(e) => {
              // Find tooltip either in container or in document body
              let tooltip = e.currentTarget.querySelector(".app-label-tooltip");
              if (!tooltip) {
                // Tooltip might be in document body, find it by class
                const tooltips = document.body.querySelectorAll(
                  ".app-label-tooltip.show",
                );
                tooltip = tooltips[tooltips.length - 1]; // Get the last one (most recent)
              }

              if (tooltip) {
                tooltip.classList.remove("show");

                // Move tooltip back to its original container if it was moved to body
                if (tooltip.parentElement === document.body) {
                  e.currentTarget.appendChild(tooltip);
                  // Reset positioning
                  tooltip.style.position = "";
                  tooltip.style.left = "";
                  tooltip.style.top = "";
                  tooltip.style.transform = "";
                  tooltip.style.zIndex = "";
                }
              }
            }}
          >
            <span className="app-label-overflow">+{labelsToShow.length}</span>
            <div className="app-label-tooltip">
              {labelsToShow.map((label) => (
                <div key={label.key} className="tooltip-label">
                  {label.key}: {label.value}
                </div>
              ))}
            </div>
          </div>
        </div>
      );
    }

    const visibleLabels = labelsToShow.slice(0, maxVisibleLabels);
    const hiddenCount = labelsToShow.length - maxVisibleLabels;

    return (
      <div className="app-labels">
        {visibleLabels.map((label) => (
          <span key={label.key} className="app-label">
            {label.key}: {label.value}
          </span>
        ))}
        {hiddenCount > 0 && (
          <div
            className="app-label-overflow-container"
            onMouseEnter={(e) => {
              const tooltip =
                e.currentTarget.querySelector(".app-label-tooltip");
              if (tooltip) {
                const rect = e.currentTarget.getBoundingClientRect();
                const leftPos = rect.left + rect.width / 2;
                const topPos = rect.bottom + 6;

                // PORTAL APPROACH: Move tooltip to document body to escape ALL clipping
                const tooltipLeft = leftPos - 100; // Center minus half tooltip width

                // Move to document body to escape any container clipping
                document.body.appendChild(tooltip);

                tooltip.style.position = "fixed";
                tooltip.style.left = `${tooltipLeft}px`;
                tooltip.style.top = `${topPos}px`;
                tooltip.style.transform = "none";
                tooltip.style.marginTop = "0";
                tooltip.style.zIndex = "2147483647";

                tooltip.classList.add("show");
              }
            }}
            onMouseLeave={(e) => {
              // Find tooltip either in container or in document body
              let tooltip = e.currentTarget.querySelector(".app-label-tooltip");
              if (!tooltip) {
                // Tooltip might be in document body, find it by class
                const tooltips = document.body.querySelectorAll(
                  ".app-label-tooltip.show",
                );
                tooltip = tooltips[tooltips.length - 1]; // Get the last one (most recent)
              }

              if (tooltip) {
                tooltip.classList.remove("show");

                // Move tooltip back to its original container if it was moved to body
                if (tooltip.parentElement === document.body) {
                  e.currentTarget.appendChild(tooltip);
                  // Reset positioning
                  tooltip.style.position = "";
                  tooltip.style.left = "";
                  tooltip.style.top = "";
                  tooltip.style.transform = "";
                  tooltip.style.zIndex = "";
                }
              }
            }}
          >
            <span className="app-label-overflow">+{hiddenCount}</span>
            <div className="app-label-tooltip">
              {labelsToShow.slice(maxVisibleLabels).map((label) => (
                <div key={label.key} className="tooltip-label">
                  {label.key}: {label.value}
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    );
  };

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

  // Handle show labels dropdown toggle
  const toggleShowLabelOptions = async () => {
    if (showLabelOptions) {
      setShowLabelOptions(false);
    } else {
      if (availableLabels.length === 0) {
        await loadAvailableLabels();
      }
      setShowLabelOptions(true);
    }
  };

  // Handle label selection for show labels
  const toggleShowLabel = (labelKey) => {
    setSelectedShowLabels((prev) => {
      const newLabels = prev.includes(labelKey)
        ? prev.filter((key) => key !== labelKey)
        : [...prev, labelKey];
      return newLabels;
    });
  };

  // Handle toggle all labels
  const toggleAllLabels = () => {
    setSelectedShowLabels((prev) => {
      // If all labels are selected, deselect all; otherwise, select all
      const allSelected = availableLabels.every((label) =>
        prev.includes(label),
      );
      return allSelected ? [] : [...availableLabels];
    });
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

  // Save show labels whenever they change
  useEffect(() => {
    userPreferences.saveShowLabels(selectedShowLabels);
  }, [selectedShowLabels]);

  // Update filtered labels when available labels change and dropdown is open
  useEffect(() => {
    if (showGroupOptions && availableLabels.length > 0) {
      setFilteredLabels(availableLabels);
    }
  }, [availableLabels, showGroupOptions]);

  // Track panel width changes
  useEffect(() => {
    const panel = panelRef.current;
    if (!panel) return;

    // Set initial width
    const initialWidth = panel.getBoundingClientRect().width;
    setPanelWidth(initialWidth);

    // Use ResizeObserver to track width changes
    const resizeObserver = new ResizeObserver((entries) => {
      for (let entry of entries) {
        setPanelWidth(entry.contentRect.width);
      }
    });

    resizeObserver.observe(panel);

    return () => {
      resizeObserver.disconnect();
    };
  }, []);

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
      setPanelWidth(newWidth);
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

        <div className="show-labels-dropdown" ref={showLabelsDropdownRef}>
          <button
            className="sort-icon-button"
            onClick={toggleShowLabelOptions}
            aria-label="Show label options"
          >
            <FaCheck className="sort-icon" aria-hidden="true" />
          </button>
          {showLabelOptions && (
            <ul className="show-labels-options">
              <li key="all" className="all-option">
                <label className="checkbox-label">
                  <input
                    type="checkbox"
                    checked={
                      availableLabels.length > 0 &&
                      availableLabels.every((label) =>
                        selectedShowLabels.includes(label),
                      )
                    }
                    onChange={toggleAllLabels}
                  />
                  <span className="checkbox-custom"></span>
                  <span className="label-text label-text-all">All</span>
                </label>
              </li>
              {availableLabels
                .slice()
                .sort((a, b) => a.localeCompare(b))
                .map((label) => (
                  <li key={label}>
                    <label className="checkbox-label">
                      <input
                        type="checkbox"
                        checked={selectedShowLabels.includes(label)}
                        onChange={() => toggleShowLabel(label)}
                      />
                      <span className="checkbox-custom"></span>
                      <span className="label-text">{label}</span>
                    </label>
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
                          <div className="app-info">
                            <div className="app-name">{app.name}</div>
                            {renderAppLabels(app, panelWidth)}
                          </div>
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
                <div className="app-info">
                  <div className="app-name">{app.name}</div>
                  {renderAppLabels(app, panelWidth)}
                </div>
                <div className={`status-indicator ${statusClass}`}>{label}</div>
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
};
