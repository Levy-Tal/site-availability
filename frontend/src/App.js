import React, { useState, useEffect } from "react";
import { MapComponent } from "./map";
import { AppStatusPanel } from "./AppStatusPanel";
import "./styles/main.css";

function App() {
  const [locations, setLocations] = useState([]);
  const [apps, setApps] = useState([]);
  const [selectedSite, setSelectedSite] = useState(null);
  const [isPanelOpen, setIsPanelOpen] = useState(false); // Track panel visibility

  useEffect(() => {
    // Mock the fetchAppStatuses until the API is available
    const mockData = {
      locations: [
        { name: "New York", Latitude: 40.7128, Longitude: -74.0060 },
        { name: "San Francisco", Latitude: 37.7749, Longitude: -122.4194 },
        { name: "London", Latitude: 51.5074, Longitude: -0.1278 },
      ],
      apps: [
        { name: "App1", location: "New York", status: "up" },
        { name: "App2", location: "San Francisco", status: "down" },
        { name: "App3", location: "London", status: "up" },
      ],
    };
    setLocations(mockData.locations);
    setApps(mockData.apps);
  }, []);

  const handleSiteClick = (site) => {
    if (selectedSite === site) {
      // If the same site is clicked, toggle the panel visibility
      setIsPanelOpen(!isPanelOpen);
    } else {
      // If a new site is clicked, open the panel
      setSelectedSite(site);
      setIsPanelOpen(true);
    }
  };

  return (
    <div className="app-container">
      {/* Ensure the map is visible */}
      <MapComponent locations={locations} onSiteClick={handleSiteClick} />
      
      {/* Render the panel only if it's open */}
      {isPanelOpen && selectedSite && (
        <AppStatusPanel site={selectedSite} apps={apps} />
      )}
    </div>
  );
}

export default App;
