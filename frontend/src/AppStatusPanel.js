import React from "react";

export const AppStatusPanel = ({ site, apps }) => {
  const appsInSite = apps.filter((app) => app.location === site.name);

  return (
    <div className="status-panel">
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
