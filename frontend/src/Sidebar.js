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
import { userPreferences } from "./utils/storage";

const Sidebar = ({
  onStatusFilterChange,
  onLabelFilterChange,
  onDocsClick,
  isCollapsed,
  onToggleCollapse,
  selectedStatusFilters,
  selectedLabels,
  docsTitle,
}) => {
  useEffect(() => {
    const savedLabelFilters = userPreferences.loadLabelFilters();

    if (JSON.stringify(savedLabelFilters) !== JSON.stringify(selectedLabels)) {
      onLabelFilterChange(savedLabelFilters);
    }
  }, []);

  const sidebarRef = useRef(null);
  const keyDropdownRef = useRef(null);
  const valueDropdownRef = useRef(null);
  const [labelKeys, setLabelKeys] = useState([]);
  const [keyInput, setKeyInput] = useState("");
  const [valueInput, setValueInput] = useState("");
  const [selectedKey, setSelectedKey] = useState("");
  const [keySuggestions, setKeySuggestions] = useState([]);
  const [valueSuggestions, setValueSuggestions] = useState([]);
  const [showKeySuggestions, setShowKeySuggestions] = useState(false);
  const [showValueSuggestions, setShowValueSuggestions] = useState(false);

  useEffect(() => {
    const handleOutsideClick = (e) => {
      if (
        showKeySuggestions &&
        keyDropdownRef.current &&
        !keyDropdownRef.current.contains(e.target)
      ) {
        setShowKeySuggestions(false);
      }

      if (
        showValueSuggestions &&
        valueDropdownRef.current &&
        !valueDropdownRef.current.contains(e.target)
      ) {
        setShowValueSuggestions(false);
      }
    };

    document.addEventListener("mousedown", handleOutsideClick);
    return () => document.removeEventListener("mousedown", handleOutsideClick);
  }, [showKeySuggestions, showValueSuggestions]);

  const handleKeyInputChange = async (e) => {
    const value = e.target.value;
    setKeyInput(value);

    if (value.trim() === "") {
      setSelectedKey("");
      setValueInput("");
      setValueSuggestions([]);
      setShowValueSuggestions(false);
    }

    if (value.length > 0) {
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

  const toggleKeySuggestions = async () => {
    if (showKeySuggestions) {
      setShowKeySuggestions(false);
    } else {
      if (labelKeys.length === 0) {
        await handleKeyInputFocus();
      } else {
        setKeySuggestions(labelKeys);
        setShowKeySuggestions(true);
      }
    }
  };

  const toggleValueSuggestions = async () => {
    if (!selectedKey) return;

    if (showValueSuggestions) {
      setShowValueSuggestions(false);
    } else {
      if (valueSuggestions.length === 0) {
        await handleValueInputFocus();
      } else {
        setShowValueSuggestions(true);
      }
    }
  };

  const handleValueInputChange = async (e) => {
    const value = e.target.value;
    setValueInput(value);

    if (selectedKey && value.length > 0) {
      const filteredValues = valueSuggestions.filter(
        (val) => val && val.toLowerCase().includes(value.toLowerCase()),
      );
      setShowValueSuggestions(filteredValues.length > 0);
    } else if (selectedKey) {
      setShowValueSuggestions(valueSuggestions.length > 0);
    }
  };

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

  const handleValueInputFocus = async () => {
    if (selectedKey) {
      try {
        const response = await fetchLabels(selectedKey);
        let values = [];
        if (Array.isArray(response)) {
          values = response.filter(
            (value) => value && typeof value === "string",
          );
        } else if (response && Array.isArray(response.labels)) {
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

  const selectKey = (key) => {
    setKeyInput(key);
    setSelectedKey(key);
    setShowKeySuggestions(false);
    setValueInput("");
    setValueSuggestions([]);
    setShowValueSuggestions(false);
  };

  const selectValue = (value) => {
    setValueInput(value);
    setShowValueSuggestions(false);
    addLabelFilter(value, selectedKey);
  };

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
        {/* Header with logo and collapse button */}
        <div className="sidebar__header">
          {!isCollapsed && (
            <>
              <img src="/logo.png" alt="Logo" className="sidebar__logo-image" />
              <div className="sidebar__logo">
                <div className="sidebar__logo-text">
                  <div>Site</div>
                  <div>Availability</div>
                </div>
              </div>
            </>
          )}
          <div className="sidebar__collapse-button" onClick={onToggleCollapse}>
            {isCollapsed ? <FaChevronRight /> : <FaChevronLeft />}
          </div>
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
            {!isCollapsed && <span>{docsTitle}</span>}
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
              <div
                className="sidebar__label-input-container"
                ref={keyDropdownRef}
              >
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

              <div
                className="sidebar__label-input-container"
                ref={valueDropdownRef}
              >
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
