import React, { useState, useEffect } from "react";
import { MapComponent } from "./map";
import { AppStatusPanel } from "./AppStatusPanel";
import { fetchAppStatuses } from "./api/appStatusAPI";
import "./styles/main.css";

function App() {
  const [locations, setLocations] = useState([]);
  const [apps, setApps] = useState([]);
  const [selectedSite, setSelectedSite] = useState(null);
  const [isPanelOpen, setIsPanelOpen] = useState(false);

  useEffect(() => {
    const getData = async () => {
      const data = await fetchAppStatuses();
      setLocations(data.locations || []); // Fallback to an empty array if locations are null
      setApps(data.apps || []); // Fallback to an empty array if apps are null
    };
    getData();
  }, []);

  const handleSiteClick = (site) => {
    if (selectedSite === site) {
      setIsPanelOpen(!isPanelOpen);
    } else {
      setSelectedSite(site);
      setIsPanelOpen(true);
    }
  };

  return (
    <div className="app-container">
      <MapComponent locations={locations} onSiteClick={handleSiteClick} apps={apps} />
      {isPanelOpen && selectedSite && (
        <AppStatusPanel
          site={selectedSite}
          apps={apps.filter((app) => app.location === selectedSite.name)} // Filter apps for the selected site
          onClose={() => setIsPanelOpen(false)}
        />
      )}
    </div>
  );
}

export default App;
