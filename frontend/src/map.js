import React, { useState, useEffect, useRef } from "react";
import { ComposableMap, Geographies, Geography, Marker, ZoomableGroup } from "react-simple-maps";

const INITIAL_ZOOM = 1;
const MIN_SCALE = 250;


export const MapComponent = ({ locations, onSiteClick, apps }) => {
  const geoUrl = "/data/countries-50m.json";
  const [zoom, setZoom] = useState(INITIAL_ZOOM);
  const [center, setCenter] = useState([0, 0]);
  const [scale, setScale] = useState(MIN_SCALE);
  const [loading, setLoading] = useState(true);
  const [baseSize, setBaseSize] = useState(0);
  const hasInitialized = useRef(false);

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

  const calculateMapSettings = (bounds, width, height) => {
    const centerLon = (bounds.minLon + bounds.maxLon) / 2;
    const centerLat = (bounds.minLat + bounds.maxLat) / 2;
    const lonRange = bounds.maxLon - bounds.minLon;
  
    let scaleFactor, zoomFactor, baseSize;
  
    if (lonRange > 200) { //size of the world
      scaleFactor = 300;
      zoomFactor = 1;
      baseSize = 1;
    } else if (lonRange > 50) { // size of continent 
      scaleFactor = 400;
      zoomFactor = 3;
      baseSize = 1;
    } else if (lonRange > 1) { //size of big countries
      scaleFactor = 500;
      zoomFactor = 3;
      baseSize = 1;
    } else if (lonRange > 0.5) { //size of small countries
      scaleFactor = 1500;
      zoomFactor = 5;
      baseSize = 0;
    } else if (lonRange > 0.1) { //size of small countries
      scaleFactor = 3000;
      zoomFactor = 5;
      baseSize = 0;
    } else { //size of small countries
      scaleFactor = 20000;
      zoomFactor = 5;
      baseSize = 0;
    }
  
    console.log("width:", width);
    console.log("lonRange:", lonRange);
    console.log("scaleFactor:", scaleFactor);
    console.log("zoomFactor:", zoomFactor);
  
    return { center: [centerLon, centerLat], zoom: zoomFactor, scale: scaleFactor ,baseSize: baseSize};
  };
  

  useEffect(() => {
    if (locations.length > 0 && !hasInitialized.current) {
      const width = window.innerWidth;
      const height = window.innerHeight;
      const bounds = calculateBounds(locations);
      const base = 0;
      if (bounds) {
        const { center, zoom, scale ,baseSize} = calculateMapSettings(bounds, width, height , base);
        setCenter(center);
        setZoom(zoom);
        setScale(scale);
        setBaseSize(baseSize);
        hasInitialized.current = true;
        setLoading(false); // Show the map only after initialization
      }
    }
  }, [locations]);

  if (loading) {
    return <div style={{ width: "100vw", height: "100vh", display: "flex", justifyContent: "center", alignItems: "center", fontSize: "20px" }}>Loading map...</div>;
  }

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
            const anyAppUnavailable = appsInSite.some((app) => app.status === "unavailable");

            const color = allAppsUp
              ? "#4CAF50" // Green
              : anyAppDown
              ? "#F44336" // Red
              : anyAppUnavailable
              ? "#FF9800" // Orange
              : "#D6D6DA"; // Default

            return (
              <Marker key={site.name} coordinates={[site.longitude, site.latitude]} onClick={() => onSiteClick(site)}>
                <text
                  x={0}
                  y={-(2 + baseSize)}
                  fill={color}
                  fontSize={3 + baseSize}
                  textAnchor="middle"
                  fontWeight="bold"
                  style={{
                    cursor: "pointer",
                  }}
                >
                  {site.name}
                </text>
                <circle r={1+baseSize} fill={color} opacity={0.8} style={{ transition: "all 0.3s ease" }} />
              </Marker>
            );
          })}
        </ZoomableGroup>
      </ComposableMap>
    </div>
  );
};
