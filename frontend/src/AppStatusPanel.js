import React, { useRef, useEffect } from "react";

export const AppStatusPanel = ({ site, apps }) => {
  const panelRef = useRef(null);

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

  const appsInSite = apps.filter((app) => app.location === site.name);

  return (
    <div className="status-panel" ref={panelRef} style={{ fontFamily: "'Roboto', sans-serif", fontSize: "14px", color: "#333" }}>
      <div className="resize-handle"></div>
      <h2 style={{ fontFamily: "'Roboto', sans-serif", fontWeight: "bold", color: "#000", fontSize: "18px" }}>
        {site.name}
      </h2>
      <ul>
        {appsInSite.map((app) => (
          <li key={app.name} style={{ marginBottom: "10px" }}>
            <div
              className="app-name"
              style={{
                fontFamily: "'Roboto', sans-serif",
                fontWeight: "bold",
                color: "#333",
                fontSize: "14px",
              }}
            >
              {app.name}
            </div>
            <div
              className={`status-indicator ${
                app.status === "up" ? "status-up" : "status-down"
              }`}
              style={{
                fontFamily: "'Roboto', sans-serif",
                fontWeight: "bold",
                fontSize: "12px",
                color: "#fff"
              }}
            >
              {app.status === "up" ? "Up" : "Down"}
            </div>
          </li>
        ))}
      </ul>
    </div>
  );
};
