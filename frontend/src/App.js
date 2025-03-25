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
    // Mock the fetchAppStatuses until the API is available , , 34.788745

    const mockData = {
      locations: [
        { name: "Hadera", Latitude: 32.446235, Longitude: 34.903852 },
        { name: "Jerusalem", Latitude: 31.782904, Longitude: 35.214774 },
        { name: "Beer Sheva", Latitude: 31.245381, Longitude: 34.788745 },
      ],
      apps: [
        { name: "App1", location: "Hadera", status: "up" },
        { name: "App1", location: "Hadera", status: "up" },
        { name: "App1", location: "Hadera", status: "up" },
        { name: "App1", location: "Hadera", status: "up" },
        { name: "App1", location: "Hadera", status: "up" },
        { name: "App1", location: "Hadera", status: "down" },
        { name: "App1", location: "Hadera", status: "up" },
        { name: "App1", location: "Hadera", status: "up" },
        { name: "App1", location: "Hadera", status: "up" },
        { name: "App1", location: "Hadera", status: "up" },
        { name: "App2", location: "Jerusalem", status: "down" },
        { name: "App3", location: "Beer Sheva", status: "up" },
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
      <MapComponent locations={locations} onSiteClick={handleSiteClick} apps={apps}/>
      
      {/* Render the panel only if it's open */}
      {isPanelOpen && selectedSite && (
        <AppStatusPanel site={selectedSite} apps={apps} />
      )}
    </div>
  );
}

export default App;
