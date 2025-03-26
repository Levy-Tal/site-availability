import React, { useState, useEffect, useRef } from "react";
import { ComposableMap, Geographies, Geography, Marker, ZoomableGroup } from "react-simple-maps";

const INITIAL_ZOOM = 5;
const DEFAULT_SCALE = 1000;

export const MapComponent = ({ locations, onSiteClick, apps }) => {
  const geoUrl = "/data/countries-50m.json";
  const [zoom, setZoom] = useState(INITIAL_ZOOM);
  const [center, setCenter] = useState([0, 0]);
  const [scale, setScale] = useState(DEFAULT_SCALE);
  const hasInitialized = useRef(false);

  // Calculate bounding box
  const calculateBounds = (locations) => {
    if (!locations.length) return null;

    return locations.reduce(
      (acc, loc) => ({
        minLon: Math.min(acc.minLon, loc.longitude),
        minLat: Math.min(acc.minLat, loc.latitude),
        maxLon: Math.max(acc.maxLon, loc.longitude),
        maxLat: Math.max(acc.maxLat, loc.latitude),
      }),
      { minLon: Infinity, minLat: Infinity, maxLon: -Infinity, maxLat: -Infinity }
    );
  };

  // Calculate optimal zoom, center, and scale
  const calculateMapSettings = (bounds, width, height) => {
    const centerLon = (bounds.minLon + bounds.maxLon) / 2;
    const centerLat = (bounds.minLat + bounds.maxLat) / 2;
    const lonRange = bounds.maxLon - bounds.minLon;
    const zoomFactor = INITIAL_ZOOM;
    const scaleFactor = Math.max(width / lonRange, DEFAULT_SCALE);
    return { center: [centerLon, centerLat], zoom: zoomFactor, scale: scaleFactor };
  };

  useEffect(() => {
    if (locations.length > 0 && !hasInitialized.current) {
      const width = window.innerWidth;
      const height = window.innerHeight;
      const bounds = calculateBounds(locations);

      if (bounds) {
        const { center, zoom, scale } = calculateMapSettings(bounds, width, height);
        setCenter(center);
        setZoom(zoom);
        setScale(scale);
        hasInitialized.current = true; // Prevent re-calculating after first load
      }
    }
  }, [locations]);

  return (
    <div id="map-container" style={{ width: "100vw", height: "100vh", position: "relative", overflow: "hidden" }}>
      <ComposableMap projectionConfig={{ scale }}>
        <ZoomableGroup center={center} zoom={zoom}>
          <Geographies geography={geoUrl}>
            {({ geographies }) =>
              geographies.map((geo) => (
                <Geography
                  key={geo.rsmKey}
                  geography={geo}
                  style={{
                    default: { fill: "#D6D6DA", outline: "none" },
                    hover: { fill: "#D6D6DA", outline: "none" },
                    pressed: { fill: "#D6D6DA", outline: "none" },
                  }}
                />
              ))
            }
          </Geographies>
          {locations.map((site) => {
            const appsInSite = apps.filter((app) => app.location === site.name);
            const allAppsUp = appsInSite.every((app) => app.status === "up");
            const anyAppDown = appsInSite.some((app) => app.status === "down");

            const color = allAppsUp ? "#4CAF50" : anyAppDown ? "#F44336" : "#FF9800"; // Modern colors

            return (
              <Marker key={site.name} coordinates={[site.longitude, site.latitude]} onClick={() => onSiteClick(site)}>
                <text
                  x={0}
                  y={-2}
                  fill={color}
                  fontSize={3} // Retained original text size
                  textAnchor="middle"
                  fontWeight="bold"
                  style={{
                    textShadow: "0px 0px 2px rgba(0, 0, 0, 0.5)",
                    fontFamily: "'Roboto', sans-serif",
                  }}
                >
                  {site.name}
                </text>
                <circle r={1} fill={color} opacity={0.8} style={{ transition: "all 0.3s ease" }} />
              </Marker>
            );
          })}
        </ZoomableGroup>
      </ComposableMap>
    </div>
  );
};
