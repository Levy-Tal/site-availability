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
  const [statusFilter, setStatusFilter] = useState("");
  const [labelFilters, setLabelFilters] = useState([]);

  // Fetch locations with their calculated status from the server
  const refreshLocations = async () => {
    try {
      const locationsData = await fetchLocations(statusFilter, labelFilters);
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
  }, [scrapeInterval, statusFilter, labelFilters]);

  // Refresh locations when filters change
  useEffect(() => {
    refreshLocations();
  }, [statusFilter, labelFilters]);

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
    if (docsInfo.docs_url) {
      window.open(docsInfo.docs_url, "_blank");
    }
  };

  const handleStatusFilterChange = (newStatusFilter) => {
    setStatusFilter(newStatusFilter);
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
        selectedStatusFilter={statusFilter}
        selectedLabels={labelFilters}
      />
      <MapComponent locations={locations} onSiteClick={handleSiteClick} />
      {isPanelOpen && selectedSite && (
        <AppStatusPanel
          site={selectedSite}
          onClose={handlePanelClose}
          scrapeInterval={scrapeInterval}
          statusFilter={statusFilter}
          labelFilters={labelFilters}
        />
      )}
    </div>
  );
}

export default App;
