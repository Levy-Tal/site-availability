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
    <div className="status-panel" ref={panelRef}>
      <div className="resize-handle"></div>
      <h2>{site.name}</h2>
      <ul>
        {appsInSite.map((app) => (
          <li key={app.name}>
            <div className="app-name">{app.name}</div>
            <div
              className={`status-indicator ${
                app.status === "up" ? "status-up" : "status-down"
              }`}
            >
              {app.status === "up" ? "Up" : "Down"}
            </div>
          </li>
        ))}
      </ul>
    </div>
  );
};
