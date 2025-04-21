import React, { useState, useEffect } from "react";
import { MapComponent } from "./map";
import { AppStatusPanel } from "./AppStatusPanel";
import { fetchAppStatuses } from "./api/appStatusAPI";
import { fetchScrapeInterval } from "./api/scrapeIntervalAPI";
import "./styles/main.css";

function App() {
  const [locations, setLocations] = useState([]);
  const [apps, setApps] = useState([]);
  const [selectedSite, setSelectedSite] = useState(null);
  const [isPanelOpen, setIsPanelOpen] = useState(false);
  const [scrapeInterval, setScrapeInterval] = useState(null);

  // Fetch app statuses from the server
  const refreshAppStatuses = async () => {
    try {
      const data = await fetchAppStatuses();
      setLocations(data.locations || []); // Fallback to an empty array if locations are null
      setApps(data.apps || []); // Fallback to an empty array if apps are null
    } catch (error) {
      console.error("Error refreshing app statuses:", error);
    }
  };

  useEffect(() => {
    // Fetch scrape interval and app statuses on initial load
    const initialize = async () => {
      try {
        const intervalData = await fetchScrapeInterval();
        setScrapeInterval(intervalData.scrape_interval_ms); // Set scrape interval in ms
        await refreshAppStatuses(); // Fetch initial app statuses
      } catch (error) {
        console.error("Error initializing app:", error);
      }
    };
    initialize();
  }, []);

  useEffect(() => {
    if (scrapeInterval) {
      // Set up periodic refresh of app statuses
      const intervalId = setInterval(() => {
        refreshAppStatuses();
      }, scrapeInterval);

      // Clean up interval on component unmount or when scrapeInterval changes
      return () => clearInterval(intervalId);
    }
  }, [scrapeInterval]);

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
