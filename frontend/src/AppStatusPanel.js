import React from "react";

export const AppStatusPanel = ({ site, apps }) => {
  const appsInSite = apps.filter((app) => app.location === site.name);

  return (
    <div className="status-panel">
      <h2>{site.name}</h2>
      <ul>
        {appsInSite.map((app) => (
          <li key={app.name}>
            <div>{app.name}</div>
            <div
              className={`status ${
                app.status === "up" ? "status-up" : "status-down"
              }`}
            >
              <span className={`status-dot ${app.status === "up" ? "green-dot" : "red-dot"}`}></span>
              {app.status === "up" ? "✔️ Up" : "❌ Down"}
            </div>
          </li>
        ))}
      </ul>
    </div>
  );
};
