import React, { useRef, useEffect, useState } from "react";
import { FaSortAmountDownAlt, FaSearch } from "react-icons/fa";

export const AppStatusPanel = ({ site, apps, onClose }) => {
  const panelRef = useRef(null);
  const [filteredApps, setFilteredApps] = useState([]);
  const [searchTerm, setSearchTerm] = useState("");
  const [sortOrder, setSortOrder] = useState("name-asc");
  const [showSortOptions, setShowSortOptions] = useState(false);

  const handleSortChange = (order) => {
    setSortOrder(order);
    setShowSortOptions(false);
  };

  const handleOutsideClick = (e) => {
    if (
      panelRef.current &&
      !panelRef.current.contains(e.target) &&
      !e.target.closest(".sort-dropdown")
    ) {
      setShowSortOptions(false);
    }
  };

  useEffect(() => {
    document.addEventListener("mousedown", handleOutsideClick);
    return () => {
      document.removeEventListener("mousedown", handleOutsideClick);
    };
  }, []);

  useEffect(() => {
    let filtered = apps.filter((app) => app.location === site.name);

    if (searchTerm) {
      filtered = filtered.filter((app) =>
        app.name.toLowerCase().startsWith(searchTerm.toLowerCase())
      );
    }

    filtered.sort((a, b) => {
      if (sortOrder === "name-asc") {
        return a.name.localeCompare(b.name);
      } else if (sortOrder === "name-desc") {
        return b.name.localeCompare(a.name);
      } else if (sortOrder === "status-up") {
        return a.status.localeCompare(b.status);
      } else if (sortOrder === "status-down") {
        return b.status.localeCompare(a.status);
      }
      return 0;
    });

    setFilteredApps(filtered);
  }, [site, apps, searchTerm, sortOrder]);

  useEffect(() => {
    const panel = panelRef.current;
    let isResizing = false;

    const handleMouseDown = (e) => {
      isResizing = true;
      document.addEventListener("mousemove", handleMouseMove);
      document.addEventListener("mouseup", handleMouseUp);
    };

    const handleMouseMove = (e) => {
      if (isResizing) {
        const newWidth = window.innerWidth - e.clientX;
        panel.style.width = `${newWidth}px`;
      }
    };

    const handleMouseUp = () => {
      isResizing = false;
      document.removeEventListener("mousemove", handleMouseMove);
      document.removeEventListener("mouseup", handleMouseUp);
    };

    const resizeHandle = panel.querySelector(".resize-handle");
    resizeHandle.addEventListener("mousedown", handleMouseDown);

    return () => {
      resizeHandle.removeEventListener("mousedown", handleMouseDown);
    };
  }, []);

  return (
    <div className="status-panel" ref={panelRef}>
      <div className="resize-handle"></div>
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
              <li
                className={sortOrder === "name-asc" ? "selected" : ""}
                onClick={() => handleSortChange("name-asc")}
              >
                Name A-Z {sortOrder === "name-asc" && <span className="checkmark">✔</span>}
              </li>
              <li
                className={sortOrder === "name-desc" ? "selected" : ""}
                onClick={() => handleSortChange("name-desc")}
              >
                Name Z-A {sortOrder === "name-desc" && <span className="checkmark">✔</span>}
              </li>
              <li
                className={sortOrder === "status-up" ? "selected" : ""}
                onClick={() => handleSortChange("status-up")}
              >
                Status (Up-Unavailable-Down){" "}
                {sortOrder === "status-up" && <span className="checkmark">✔</span>}
              </li>
              <li
                className={sortOrder === "status-down" ? "selected" : ""}
                onClick={() => handleSortChange("status-down")}
              >
                Status (Down-Unavailable-Up){" "}
                {sortOrder === "status-down" && <span className="checkmark">✔</span>}
              </li>
            </ul>
          )}
        </div>
      </div>
      <ul>
        {filteredApps.map((app) => (
          <li key={app.name}>
            <div className="app-name">{app.name}</div>
            <div
              className={`status-indicator ${
                app.status === "up"
                  ? "status-up"
                  : app.status === "down"
                  ? "status-down"
                  : "status-unavailable"
              }`}
            >
              {app.status === "up"
                ? "Up"
                : app.status === "down"
                ? "Down"
                : "Unavailable"}
            </div>
          </li>
        ))}
      </ul>
    </div>
  );
};
