import React, { useState, useEffect } from "react";
import { MapComponent } from "./map";
import { AppStatusPanel } from "./AppStatusPanel";
import { fetchLocations, fetchApps } from "./api/appStatusAPI";
import { fetchScrapeInterval } from "./api/scrapeIntervalAPI";
import { fetchDocs } from "./api/docsAPI";
import { FaBook } from "react-icons/fa"; // Import the docs icon
import "./styles/main.css";

function App() {
  const [locations, setLocations] = useState([]);
  const [selectedSite, setSelectedSite] = useState(null);
  const [isPanelOpen, setIsPanelOpen] = useState(false);
  const [scrapeInterval, setScrapeInterval] = useState(null);
  const [docsInfo, setDocsInfo] = useState({ docs_title: "", docs_url: "" });

  // Fetch locations with their calculated status from the server
  const refreshLocations = async () => {
    try {
      const locationsData = await fetchLocations();
      setLocations(locationsData);
    } catch (error) {
      console.error("Error refreshing locations:", error);
    }
  };

  useEffect(() => {
    // Fetch scrape interval, locations, and docs info on initial load
    const initialize = async () => {
      try {
        const intervalData = await fetchScrapeInterval();
        setScrapeInterval(intervalData.scrape_interval_ms); // Set scrape interval in ms
        await refreshLocations(); // Fetch initial locations with status

        const docsData = await fetchDocs();
        setDocsInfo(docsData); // Set docs info
      } catch (error) {
        console.error("Error initializing app:", error);
      }
    };
    initialize();
  }, []);

  useEffect(() => {
    if (scrapeInterval) {
      // Set up periodic refresh of locations
      const intervalId = setInterval(() => {
        refreshLocations();
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

  const handlePanelClose = () => {
    setIsPanelOpen(false);
    setSelectedSite(null);
  };

  return (
    <div className="app-container">
      {docsInfo.docs_url && (
        <div
          className="docs-button"
          onClick={() => window.open(docsInfo.docs_url, "_blank")}
          title={docsInfo.docs_title}
        >
          <FaBook size={24} />
        </div>
      )}
      <MapComponent locations={locations} onSiteClick={handleSiteClick} />
      {isPanelOpen && selectedSite && (
        <AppStatusPanel
          site={selectedSite}
          onClose={handlePanelClose}
          scrapeInterval={scrapeInterval}
        />
      )}
    </div>
  );
}

export default App;
