import React, { useState, useEffect, useRef } from "react";
import {
  FaMap,
  FaUser,
  FaBook,
  FaChevronLeft,
  FaChevronRight,
  FaTimes,
} from "react-icons/fa";
import { fetchLabels } from "./api/appStatusAPI";

const Sidebar = ({
  onStatusFilterChange,
  onLabelFilterChange,
  onDocsClick,
  isCollapsed,
  onToggleCollapse,
  selectedStatusFilter,
  selectedLabels,
}) => {
  const sidebarRef = useRef(null);
  const [labelSuggestions, setLabelSuggestions] = useState([]);
  const [labelKeys, setLabelKeys] = useState([]);
  const [labelInput, setLabelInput] = useState("");
  const [showLabelSuggestions, setShowLabelSuggestions] = useState(false);
  const [currentLabelKey, setCurrentLabelKey] = useState("");

  useEffect(() => {
    // Fetch available labels for autocomplete
    const loadLabels = async () => {
      try {
        const labels = await fetchLabels();
        // Filter out any invalid labels
        const validLabels = labels.filter(
          (label) => label && label.key && label.value,
        );
        const keys = [...new Set(validLabels.map((label) => label.key))].filter(
          (key) => key,
        );
        setLabelKeys(keys);
        setLabelSuggestions(validLabels);
      } catch (error) {
        console.error("Error loading labels:", error);
      }
    };
    loadLabels();
  }, []);

  // Resize functionality
  useEffect(() => {
    const sidebar = sidebarRef.current;
    const handle = sidebar?.querySelector(".sidebar__resize-handle");

    if (!handle) return;

    let isResizing = false;

    const onMouseMove = (e) => {
      if (isResizing) {
        e.preventDefault();
        const newWidth = e.clientX;
        if (newWidth >= 200 && newWidth <= 600) {
          sidebar.style.width = `${newWidth}px`;
          // Update the app container margin
          const appContainer = document.querySelector(
            ".app-container--with-sidebar",
          );
          if (appContainer && !isCollapsed) {
            appContainer.style.marginLeft = `${newWidth}px`;
          }
        }
      }
    };

    const onMouseUp = (e) => {
      if (isResizing) {
        e.preventDefault();
        isResizing = false;
        document.removeEventListener("mousemove", onMouseMove);
        document.removeEventListener("mouseup", onMouseUp);
        document.body.style.cursor = "default";
      }
    };

    const onMouseDown = (e) => {
      // Only start resizing if we clicked directly on the resize handle
      if (e.target === handle) {
        e.preventDefault();
        e.stopPropagation();
        isResizing = true;
        document.addEventListener("mousemove", onMouseMove);
        document.addEventListener("mouseup", onMouseUp);
        document.body.style.cursor = "col-resize";
      }
    };

    handle.addEventListener("mousedown", onMouseDown);
    return () => {
      handle.removeEventListener("mousedown", onMouseDown);
      document.removeEventListener("mousemove", onMouseMove);
      document.removeEventListener("mouseup", onMouseUp);
    };
  }, [isCollapsed]);

  const handleLabelInputChange = (e) => {
    const value = e.target.value;
    setLabelInput(value);

    if (value.includes("=")) {
      const [key, val] = value.split("=", 2);
      setCurrentLabelKey(key);

      if (key && val.length > 0) {
        // Filter suggestions for values of the specific key
        const filteredSuggestions = labelSuggestions
          .filter(
            (label) =>
              label &&
              label.key === key &&
              label.value &&
              label.value.toLowerCase().includes(val.toLowerCase()),
          )
          .map((label) => `${label.key}=${label.value}`);
        setShowLabelSuggestions(filteredSuggestions.length > 0);
      } else if (key && val.length === 0) {
        // Show all values for the key
        const filteredSuggestions = labelSuggestions
          .filter((label) => label && label.key === key)
          .map((label) => `${label.key}=${label.value}`);
        setShowLabelSuggestions(filteredSuggestions.length > 0);
      }
    } else {
      // Filter suggestions for keys
      const filteredKeys = labelKeys.filter(
        (key) => key && key.toLowerCase().includes(value.toLowerCase()),
      );
      setShowLabelSuggestions(filteredKeys.length > 0 && value.length > 0);
    }
  };

  const handleLabelInputKeyPress = (e) => {
    if (e.key === "Enter" && labelInput.trim()) {
      addLabelFilter(labelInput.trim());
    }
  };

  const addLabelFilter = (labelFilter) => {
    if (labelFilter.includes("=")) {
      const [key, value] = labelFilter.split("=", 2);
      if (key && value) {
        const newLabel = { key: key.trim(), value: value.trim() };
        const exists = selectedLabels.some(
          (label) =>
            label.key === newLabel.key && label.value === newLabel.value,
        );
        if (!exists) {
          onLabelFilterChange([...selectedLabels, newLabel]);
        }
        setLabelInput("");
        setShowLabelSuggestions(false);
      }
    }
  };

  const removeLabelFilter = (indexToRemove) => {
    const updatedLabels = selectedLabels.filter(
      (_, index) => index !== indexToRemove,
    );
    onLabelFilterChange(updatedLabels);
  };

  const getSuggestionsList = () => {
    if (labelInput.includes("=")) {
      const [key, val] = labelInput.split("=", 2);
      return labelSuggestions
        .filter(
          (label) =>
            label &&
            label.key === key &&
            label.value &&
            label.value.toLowerCase().includes(val.toLowerCase()),
        )
        .map((label) => `${label.key}=${label.value}`);
    } else {
      return labelKeys
        .filter(
          (key) => key && key.toLowerCase().includes(labelInput.toLowerCase()),
        )
        .map((key) => `${key}=`);
    }
  };

  const selectSuggestion = (suggestion) => {
    if (suggestion.endsWith("=")) {
      setLabelInput(suggestion);
      setShowLabelSuggestions(false);
    } else {
      addLabelFilter(suggestion);
    }
  };

  return (
    <div
      className={`sidebar ${isCollapsed ? "sidebar--collapsed" : ""}`}
      ref={sidebarRef}
    >
      <div className="sidebar__container">
        <div className="sidebar__resize-handle" />

        {/* Header with logo and collapse button */}
        <div className="sidebar__header">
          <div className="sidebar__collapse-button" onClick={onToggleCollapse}>
            {isCollapsed ? <FaChevronRight /> : <FaChevronLeft />}
          </div>
          {!isCollapsed && (
            <div className="sidebar__logo">
              <span className="sidebar__logo-text">Site Monitor</span>
            </div>
          )}
        </div>

        {/* Navigation items */}
        <div className="sidebar__nav">
          <div className="sidebar__nav-item sidebar__nav-item--active">
            <FaMap className="sidebar__nav-icon" />
            {!isCollapsed && <span>Map</span>}
          </div>
          <div className="sidebar__nav-item">
            <FaUser className="sidebar__nav-icon" />
            {!isCollapsed && <span>User Info</span>}
          </div>
          <div className="sidebar__nav-item" onClick={onDocsClick}>
            <FaBook className="sidebar__nav-icon" />
            {!isCollapsed && <span>Documentation</span>}
          </div>
        </div>

        {/* Filters section */}
        {!isCollapsed && (
          <div className="sidebar__filters">
            <div className="sidebar__filters-title">Filters</div>

            {/* Status Filter */}
            <div className="sidebar__filter-group">
              <div className="sidebar__filter-header">STATUS</div>
              <div className="sidebar__filter-options">
                <label className="sidebar__filter-option">
                  <input
                    type="radio"
                    name="status"
                    value=""
                    checked={selectedStatusFilter === ""}
                    onChange={(e) => onStatusFilterChange(e.target.value)}
                  />
                  <span className="sidebar__radio"></span>
                  All
                </label>
                <label className="sidebar__filter-option">
                  <input
                    type="radio"
                    name="status"
                    value="up"
                    checked={selectedStatusFilter === "up"}
                    onChange={(e) => onStatusFilterChange(e.target.value)}
                  />
                  <span className="sidebar__radio"></span>
                  UP
                </label>
                <label className="sidebar__filter-option">
                  <input
                    type="radio"
                    name="status"
                    value="down"
                    checked={selectedStatusFilter === "down"}
                    onChange={(e) => onStatusFilterChange(e.target.value)}
                  />
                  <span className="sidebar__radio"></span>
                  DOWN
                </label>
                <label className="sidebar__filter-option">
                  <input
                    type="radio"
                    name="status"
                    value="unavailable"
                    checked={selectedStatusFilter === "unavailable"}
                    onChange={(e) => onStatusFilterChange(e.target.value)}
                  />
                  <span className="sidebar__radio"></span>
                  UNAVAILABLE
                </label>
              </div>
            </div>

            {/* Labels Filter */}
            <div className="sidebar__filter-group">
              <div className="sidebar__filter-header">LABELS</div>
              <div className="sidebar__label-input-container">
                <input
                  type="text"
                  placeholder="key=value"
                  value={labelInput}
                  onChange={handleLabelInputChange}
                  onKeyPress={handleLabelInputKeyPress}
                  onFocus={() => setShowLabelSuggestions(true)}
                  className="sidebar__label-input"
                />
                {showLabelSuggestions && (
                  <div className="sidebar__suggestions">
                    {getSuggestionsList().map((suggestion, index) => (
                      <div
                        key={index}
                        className="sidebar__suggestion"
                        onClick={() => selectSuggestion(suggestion)}
                      >
                        {suggestion}
                      </div>
                    ))}
                  </div>
                )}
              </div>

              {/* Selected Labels */}
              <div className="sidebar__selected-labels">
                {selectedLabels.map((label, index) => (
                  <div key={index} className="sidebar__selected-label">
                    <span>
                      {label.key}={label.value}
                    </span>
                    <FaTimes
                      className="sidebar__remove-label"
                      onClick={() => removeLabelFilter(index)}
                    />
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default Sidebar;
