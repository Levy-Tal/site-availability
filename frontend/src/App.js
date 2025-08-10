import React, { useState, useEffect, useCallback } from "react";
import { MapComponent } from "./map";
import { AppStatusPanel } from "./AppStatusPanel";
import Sidebar from "./Sidebar";
import { AuthProvider, useAuth } from "./contexts/AuthContext";
import SimpleLogin from "./components/SimpleLogin";
import SimpleUserModal from "./components/SimpleUserModal";
import Spinner from "./components/Spinner";
import { fetchLocations } from "./api/appStatusAPI";
import { fetchScrapeInterval } from "./api/scrapeIntervalAPI";
import { fetchDocs } from "./api/docsAPI";
import { userPreferences } from "./utils/storage";
import "./styles/main.css";

function AppContent() {
  const { isInitialized, needsAuthentication, user } = useAuth();
  const [isUserModalOpen, setIsUserModalOpen] = useState(false);

  // Show login if authentication is required
  if (isInitialized && needsAuthentication) {
    return <SimpleLogin />;
  }

  // Show loading if not initialized
  if (!isInitialized) {
    return (
      <div
        style={{
          position: "fixed",
          top: 0,
          left: 0,
          width: "100vw",
          height: "100vh",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          zIndex: 2000,
        }}
      >
        <Spinner size="lg" label="Initializing" />
      </div>
    );
  }

  // Render original app with minimal changes
  return (
    <>
      <OriginalApp
        user={user}
        onUserInfoClick={() => setIsUserModalOpen(true)}
      />
      <SimpleUserModal
        isOpen={isUserModalOpen}
        onClose={() => setIsUserModalOpen(false)}
      />
    </>
  );
}

function OriginalApp({ user, onUserInfoClick }) {
  const [locations, setLocations] = useState([]);
  const [selectedSite, setSelectedSite] = useState(null);
  const [isPanelOpen, setIsPanelOpen] = useState(false);
  const [scrapeInterval, setScrapeInterval] = useState(null);
  const [docsInfo, setDocsInfo] = useState({ docs_title: "", docs_url: "" });
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(
    userPreferences.loadSidebarCollapsed(),
  );
  const [statusFilters, setStatusFilters] = useState(
    userPreferences.loadStatusFilters(),
  );
  const [labelFilters, setLabelFilters] = useState(
    userPreferences.loadLabelFilters(),
  );

  // Fetch locations with their calculated status from the server
  const refreshLocations = useCallback(async () => {
    try {
      const locationsData = await fetchLocations(statusFilters, labelFilters);
      setLocations(locationsData);
    } catch (error) {
      console.error("Error refreshing locations:", error);
    }
  }, [statusFilters, labelFilters]);

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
    if (scrapeInterval && !isPanelOpen) {
      // Only set up periodic refresh of locations when panel is closed
      // When panel is open, it will coordinate the refresh
      const intervalId = setInterval(() => {
        refreshLocations();
      }, scrapeInterval);

      // Clean up interval on component unmount or when scrapeInterval changes
      return () => clearInterval(intervalId);
    }
  }, [scrapeInterval, statusFilters, labelFilters, isPanelOpen]);

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

  // Save sidebar collapsed state whenever it changes
  useEffect(() => {
    userPreferences.saveSidebarCollapsed(isSidebarCollapsed);
  }, [isSidebarCollapsed]);

  // Save status filters whenever they change
  useEffect(() => {
    userPreferences.saveStatusFilters(statusFilters);
  }, [statusFilters]);

  // Save label filters whenever they change
  useEffect(() => {
    userPreferences.saveLabelFilters(labelFilters);
  }, [labelFilters]);

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
        onUserInfoClick={user ? onUserInfoClick : undefined}
        isCollapsed={isSidebarCollapsed}
        onToggleCollapse={handleSidebarToggle}
        selectedStatusFilters={statusFilters}
        selectedLabels={labelFilters}
        docsTitle={docsInfo.docs_title || "Documentation"}
        user={user}
        locations={locations}
      />
      <MapComponent locations={locations} onSiteClick={handleSiteClick} />
      {isPanelOpen && selectedSite && (
        <AppStatusPanel
          site={selectedSite}
          onClose={handlePanelClose}
          scrapeInterval={scrapeInterval}
          statusFilters={statusFilters}
          labelFilters={labelFilters}
          refreshLocations={refreshLocations}
        />
      )}
    </div>
  );
}

function App() {
  return (
    <AuthProvider>
      <AppContent />
    </AuthProvider>
  );
}

export default App;
