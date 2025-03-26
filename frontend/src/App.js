import React, { useState, useEffect } from "react";
import { MapComponent } from "./map";
import { AppStatusPanel } from "./AppStatusPanel";
import { fetchAppStatuses } from "./fetchAppStatuses";
import "./styles/main.css";

function App() {
  const [locations, setLocations] = useState([]);
  const [apps, setApps] = useState([]);
  const [selectedSite, setSelectedSite] = useState(null);
  const [isPanelOpen, setIsPanelOpen] = useState(false);

  useEffect(() => {
    const getData = async () => {
      const data = await fetchAppStatuses();
      setLocations(data.locations);
      setApps(data.apps);
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
      {isPanelOpen && selectedSite && <AppStatusPanel site={selectedSite} apps={apps} />}
    </div>
  );
}

export default App;
