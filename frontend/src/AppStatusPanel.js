import React, { useRef, useEffect, useState } from "react";
import { FaSortAmountDownAlt, FaSearch } from "react-icons/fa";
import { fetchApps } from "./api/appStatusAPI";

export const AppStatusPanel = ({
  site,
  onClose,
  scrapeInterval,
  statusFilter,
  labelFilters,
}) => {
  const panelRef = useRef(null);

  const [apps, setApps] = useState([]);
  const [filteredApps, setFilteredApps] = useState([]);
  const [searchTerm, setSearchTerm] = useState("");
  const [sortOrder, setSortOrder] = useState("name-asc");
  const [showSortOptions, setShowSortOptions] = useState(false);

  // Fetch apps for this location
  const refreshApps = async () => {
    try {
      const appsData = await fetchApps(site.name, statusFilter, labelFilters);
      setApps(appsData);
    } catch (error) {
      console.error("Error fetching apps for location:", error);
    }
  };

  // Initial fetch when panel opens
  useEffect(() => {
    refreshApps();
  }, [site.name, statusFilter, labelFilters]);

  // Set up periodic refresh while panel is open
  useEffect(() => {
    if (scrapeInterval) {
      const intervalId = setInterval(() => {
        refreshApps();
      }, scrapeInterval);

      return () => clearInterval(intervalId);
    }
  }, [scrapeInterval, site.name, statusFilter, labelFilters]);

  // Close sort dropdown on outside click
  useEffect(() => {
    const handleOutsideClick = (e) => {
      if (
        panelRef.current &&
        !panelRef.current.contains(e.target) &&
        !e.target.closest(".sort-dropdown")
      ) {
        setShowSortOptions(false);
      }
    };
    document.addEventListener("mousedown", handleOutsideClick);
    return () => document.removeEventListener("mousedown", handleOutsideClick);
  }, []);

  // Filter and sort apps
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
  }, [apps, site, searchTerm, sortOrder]);

  // Resizable panel
  useEffect(() => {
    const panel = panelRef.current;
    const handle = panel?.querySelector(".resize-handle");

    if (!handle) return;

    let isResizing = false;

    const onMouseMove = (e) => {
      if (isResizing) {
        const newWidth = window.innerWidth - e.clientX;
        panel.style.width = `${newWidth}px`;
      }
    };

    const onMouseUp = () => {
      isResizing = false;
      document.removeEventListener("mousemove", onMouseMove);
      document.removeEventListener("mouseup", onMouseUp);
    };

    const onMouseDown = () => {
      isResizing = true;
      document.addEventListener("mousemove", onMouseMove);
      document.addEventListener("mouseup", onMouseUp);
    };

    handle.addEventListener("mousedown", onMouseDown);
    return () => handle.removeEventListener("mousedown", onMouseDown);
  }, []);

  const renderSortOption = (label, value) => (
    <li
      className={sortOrder === value ? "selected" : ""}
      onClick={() => {
        setSortOrder(value);
        setShowSortOptions(false);
      }}
    >
      {label} {sortOrder === value && <span className="checkmark">✔</span>}
    </li>
  );

  return (
    <div className="status-panel" ref={panelRef}>
      <div className="resize-handle" />
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

        <div className="sort-dropdown">
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
              {renderSortOption("Status (Up-Unavailable-Down)", "status-up")}
              {renderSortOption("Status (Down-Unavailable-Up)", "status-down")}
            </ul>
          )}
        </div>
      </div>

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
    </div>
  );
};
