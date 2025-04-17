import React, { useRef, useEffect, useState } from "react";

export const AppStatusPanel = ({ site, apps }) => {
  const panelRef = useRef(null);
  const [filteredApps, setFilteredApps] = useState([]);

  useEffect(() => {
    setFilteredApps([]);
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
    <div className="status-panel" ref={panelRef}>
      <div className="resize-handle"></div>
      <h2>{site.name}</h2>
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
