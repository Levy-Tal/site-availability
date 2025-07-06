import React, { useState, useEffect } from "react";
import { MapComponent } from "./map";
import { AppStatusPanel } from "./AppStatusPanel";
import Sidebar from "./Sidebar";
import { fetchLocations, fetchApps } from "./api/appStatusAPI";
import { fetchScrapeInterval } from "./api/scrapeIntervalAPI";
import { fetchDocs } from "./api/docsAPI";
import "./styles/main.css";

function App() {
  const [locations, setLocations] = useState([]);
  const [selectedSite, setSelectedSite] = useState(null);
  const [isPanelOpen, setIsPanelOpen] = useState(false);
  const [scrapeInterval, setScrapeInterval] = useState(null);
  const [docsInfo, setDocsInfo] = useState({ docs_title: "", docs_url: "" });
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false);
  const [statusFilters, setStatusFilters] = useState([]);
  const [labelFilters, setLabelFilters] = useState([]);

  // Fetch locations with their calculated status from the server
  const refreshLocations = async () => {
    try {
      const locationsData = await fetchLocations(statusFilters, labelFilters);
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
  }, [scrapeInterval, statusFilters, labelFilters]);

  // Refresh locations when filters change
  useEffect(() => {
    refreshLocations();
  }, [statusFilters, labelFilters]);

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

  const handleDocsClick = () => {
    // Use the URL from API response, or fallback to default if empty
    const docsUrl =
      docsInfo.docs_url || "https://levy-tal.github.io/site-availability/";
    window.open(docsUrl, "_blank");
  };

  const handleStatusFilterChange = (status) => {
    setStatusFilters((prevFilters) => {
      if (prevFilters.includes(status)) {
        // Remove the status if it's already selected
        return prevFilters.filter((filter) => filter !== status);
      } else {
        // Add the status if it's not selected
        return [...prevFilters, status];
      }
    });
  };

  const handleLabelFilterChange = (newLabelFilters) => {
    setLabelFilters(newLabelFilters);
  };

  const handleSidebarToggle = () => {
    setIsSidebarCollapsed(!isSidebarCollapsed);
  };

  return (
    <div
      className={`app-container ${
        isSidebarCollapsed
          ? "app-container--with-sidebar app-container--sidebar-collapsed"
          : "app-container--with-sidebar"
      }`}
    >
      <Sidebar
        onStatusFilterChange={handleStatusFilterChange}
        onLabelFilterChange={handleLabelFilterChange}
        onDocsClick={handleDocsClick}
        isCollapsed={isSidebarCollapsed}
        onToggleCollapse={handleSidebarToggle}
        selectedStatusFilters={statusFilters}
        selectedLabels={labelFilters}
        docsTitle={docsInfo.docs_title || "Documentation"}
      />
      <MapComponent locations={locations} onSiteClick={handleSiteClick} />
      {isPanelOpen && selectedSite && (
        <AppStatusPanel
          site={selectedSite}
          onClose={handlePanelClose}
          scrapeInterval={scrapeInterval}
          statusFilters={statusFilters}
          labelFilters={labelFilters}
        />
      )}
    </div>
  );
}

export default App;
