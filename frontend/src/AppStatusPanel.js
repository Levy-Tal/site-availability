import React, { useRef, useEffect, useState } from "react";

export const AppStatusPanel = ({ site, apps }) => {
  const panelRef = useRef(null);
  const [filteredApps, setFilteredApps] = useState([]);

  useEffect(() => {
    setFilteredApps([]); // Clear apps before setting new ones
    setTimeout(() => {
      setFilteredApps(apps.filter((app) => app.location === site.name));
    }, 0);
  }, [site, apps]);

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
    <div className="status-panel" ref={panelRef} style={{ fontFamily: "'Roboto', sans-serif", fontSize: "14px", color: "#333" }}>
      <div className="resize-handle"></div>
      <h2 style={{ fontFamily: "'Roboto', sans-serif", fontWeight: "bold", color: "#000", fontSize: "18px" }}>
        {site.name}
      </h2>
      <ul>
        {filteredApps.map((app) => (
          <li key={app.name} style={{ marginBottom: "10px" }}>
            <div
              className="app-name"
              style={{
                fontFamily: "'Roboto', sans-serif",
                fontWeight: "bold",
                color: "#333",
              }}
            >
              {app.name}
            </div>
            <div
              className={`status-indicator ${
                app.status === "up"
                  ? "status-up"
                  : app.status === "down"
                  ? "status-down"
                  : "status-unavailable"
              }`}
              style={{
                fontFamily: "'Roboto', sans-serif",
                fontWeight: "bold",
                fontSize: "12px",
                color: "#fff", // Ensure text color is white
              }}
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
