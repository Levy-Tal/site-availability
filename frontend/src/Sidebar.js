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
  selectedStatusFilters,
  selectedLabels,
}) => {
  const sidebarRef = useRef(null);
  const [labelKeys, setLabelKeys] = useState([]);
  const [keyInput, setKeyInput] = useState("");
  const [valueInput, setValueInput] = useState("");
  const [selectedKey, setSelectedKey] = useState("");
  const [keySuggestions, setKeySuggestions] = useState([]);
  const [valueSuggestions, setValueSuggestions] = useState([]);
  const [showKeySuggestions, setShowKeySuggestions] = useState(false);
  const [showValueSuggestions, setShowValueSuggestions] = useState(false);

  // Load labels on component mount - now done dynamically on user interaction
  useEffect(() => {
    // Labels are now loaded dynamically when user focuses on input
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

  // Handle key input changes
  const handleKeyInputChange = async (e) => {
    const value = e.target.value;
    setKeyInput(value);

    // If key input is cleared, reset selected key and disable value input
    if (value.trim() === "") {
      setSelectedKey("");
      setValueInput("");
      setValueSuggestions([]);
      setShowValueSuggestions(false);
    }

    if (value.length > 0) {
      // Filter existing key suggestions
      const filteredKeys = labelKeys.filter(
        (key) => key && key.toLowerCase().includes(value.toLowerCase()),
      );
      setKeySuggestions(filteredKeys);
      setShowKeySuggestions(filteredKeys.length > 0);
    } else {
      setKeySuggestions(labelKeys);
      setShowKeySuggestions(labelKeys.length > 0);
    }
  };

  // Toggle key suggestions
  const toggleKeySuggestions = async () => {
    if (showKeySuggestions) {
      setShowKeySuggestions(false);
    } else {
      if (labelKeys.length === 0) {
        // Load keys if not already loaded
        await handleKeyInputFocus();
      } else {
        setKeySuggestions(labelKeys);
        setShowKeySuggestions(true);
      }
    }
  };

  // Toggle value suggestions
  const toggleValueSuggestions = async () => {
    if (!selectedKey) return; // Can't toggle if no key selected

    if (showValueSuggestions) {
      setShowValueSuggestions(false);
    } else {
      if (valueSuggestions.length === 0) {
        // Load values if not already loaded
        await handleValueInputFocus();
      } else {
        setShowValueSuggestions(true);
      }
    }
  };

  // Handle value input changes
  const handleValueInputChange = async (e) => {
    const value = e.target.value;
    setValueInput(value);

    if (selectedKey && value.length > 0) {
      // Filter existing value suggestions
      const filteredValues = valueSuggestions.filter(
        (val) => val && val.toLowerCase().includes(value.toLowerCase()),
      );
      setShowValueSuggestions(filteredValues.length > 0);
    } else if (selectedKey) {
      setShowValueSuggestions(valueSuggestions.length > 0);
    }
  };

  // Handle key input focus
  const handleKeyInputFocus = async () => {
    try {
      const response = await fetchLabels();
      const keys = response.filter((key) => key && typeof key === "string");
      setLabelKeys(keys);
      setKeySuggestions(keys);
      setShowKeySuggestions(keys.length > 0);
    } catch (error) {
      console.error("Error loading keys on focus:", error);
    }
  };

  // Handle value input focus
  const handleValueInputFocus = async () => {
    if (selectedKey) {
      try {
        const response = await fetchLabels(selectedKey);
        // The API may return different formats for values
        let values = [];
        if (Array.isArray(response)) {
          // If response is an array of strings (value names)
          values = response.filter(
            (value) => value && typeof value === "string",
          );
        } else if (response && Array.isArray(response.labels)) {
          // If response contains a labels array with objects
          values = response.labels
            .filter((label) => label && label.value)
            .map((label) => label.value);
        }
        setValueSuggestions(values);
        setShowValueSuggestions(values.length > 0);
      } catch (error) {
        console.error("Error loading values on focus:", error);
      }
    }
  };

  // Handle key selection
  const selectKey = (key) => {
    setKeyInput(key);
    setSelectedKey(key);
    setShowKeySuggestions(false);
    setValueInput("");
    setValueSuggestions([]);
    setShowValueSuggestions(false);
  };

  // Handle value selection
  const selectValue = (value) => {
    setValueInput(value);
    setShowValueSuggestions(false);
    // Automatically add the label when value is selected - pass both key and value directly
    addLabelFilter(value, selectedKey);
  };

  // Handle key press events
  const handleKeyInputKeyPress = (e) => {
    if (e.key === "Enter" && keyInput.trim()) {
      selectKey(keyInput.trim());
    }
  };

  const handleValueInputKeyPress = (e) => {
    if (e.key === "Enter" && valueInput.trim() && selectedKey) {
      addLabelFilter();
    }
  };

  const addLabelFilter = (directValue = null, directKey = null) => {
    const value = directValue || valueInput.trim();
    const key = directKey || selectedKey;

    if (key && value) {
      const newLabel = { key: key.trim(), value: value.trim() };
      const exists = selectedLabels.some(
        (label) => label.key === newLabel.key && label.value === newLabel.value,
      );
      if (!exists) {
        onLabelFilterChange([...selectedLabels, newLabel]);
      }
      // Reset inputs
      setKeyInput("");
      setValueInput("");
      setSelectedKey("");
      setShowKeySuggestions(false);
      setShowValueSuggestions(false);
    }
  };

  const removeLabelFilter = (indexToRemove) => {
    const updatedLabels = selectedLabels.filter(
      (_, index) => index !== indexToRemove,
    );
    onLabelFilterChange(updatedLabels);
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
            <div className="sidebar__filters-title"></div>

            {/* Status Filter */}
            <div className="sidebar__filter-group">
              <div className="sidebar__filter-header">STATUS</div>
              <div className="sidebar__filter-options">
                <label className="sidebar__filter-option">
                  <input
                    type="checkbox"
                    value="up"
                    checked={selectedStatusFilters.includes("up")}
                    onChange={() => onStatusFilterChange("up")}
                  />
                  <span className="sidebar__checkbox"></span>
                  <span className="sidebar__status-circle sidebar__status-circle--up"></span>
                  Up
                </label>
                <label className="sidebar__filter-option">
                  <input
                    type="checkbox"
                    value="down"
                    checked={selectedStatusFilters.includes("down")}
                    onChange={() => onStatusFilterChange("down")}
                  />
                  <span className="sidebar__checkbox"></span>
                  <span className="sidebar__status-circle sidebar__status-circle--down"></span>
                  Down
                </label>
                <label className="sidebar__filter-option">
                  <input
                    type="checkbox"
                    value="unavailable"
                    checked={selectedStatusFilters.includes("unavailable")}
                    onChange={() => onStatusFilterChange("unavailable")}
                  />
                  <span className="sidebar__checkbox"></span>
                  <span className="sidebar__status-circle sidebar__status-circle--unavailable"></span>
                  Unavailable
                </label>
              </div>
            </div>

            {/* Labels Filter */}
            <div className="sidebar__filter-group">
              <div className="sidebar__filter-header">LABELS</div>
              <div className="sidebar__label-input-container">
                <input
                  type="text"
                  placeholder="key"
                  value={keyInput}
                  onChange={handleKeyInputChange}
                  onFocus={handleKeyInputFocus}
                  onKeyPress={handleKeyInputKeyPress}
                  className="sidebar__label-input"
                />
                <div
                  className="sidebar__input-arrow"
                  onClick={toggleKeySuggestions}
                >
                  {showKeySuggestions ? "▲" : "▼"}
                </div>
                {showKeySuggestions && (
                  <div className="sidebar__suggestions">
                    {keySuggestions.map((key, index) => (
                      <div
                        key={index}
                        className="sidebar__suggestion"
                        onClick={() => selectKey(key)}
                      >
                        {key}
                      </div>
                    ))}
                  </div>
                )}
              </div>

              <div className="sidebar__label-input-container">
                <input
                  type="text"
                  placeholder="value"
                  value={valueInput}
                  onChange={handleValueInputChange}
                  onFocus={handleValueInputFocus}
                  onKeyPress={handleValueInputKeyPress}
                  className={`sidebar__label-input ${
                    !selectedKey ? "sidebar__label-input--disabled" : ""
                  }`}
                  disabled={!selectedKey}
                />
                <div
                  className={`sidebar__input-arrow ${
                    !selectedKey ? "sidebar__input-arrow--disabled" : ""
                  }`}
                  onClick={toggleValueSuggestions}
                >
                  {showValueSuggestions ? "▲" : "▼"}
                </div>
                {showValueSuggestions && selectedKey && (
                  <div className="sidebar__suggestions">
                    {valueSuggestions.map((value, index) => (
                      <div
                        key={index}
                        className="sidebar__suggestion"
                        onClick={() => selectValue(value)}
                      >
                        {value}
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
