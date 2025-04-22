import React, { useState, useEffect, useRef } from "react";
import { ComposableMap, Geographies, Geography, Marker, ZoomableGroup } from "react-simple-maps";

const INITIAL_ZOOM = 1;
const MIN_SCALE = 250;

export const MapComponent = ({ locations, onSiteClick, apps }) => {
  const geoUrl = "/data/countries-50m.json";
  const [zoom, setZoom] = useState(INITIAL_ZOOM);
  const [scale, setScale] = useState(MIN_SCALE);
  const [baseSize, setBaseSize] = useState(0);
  const [initialZoom, setInitialZoom] = useState(INITIAL_ZOOM);
  const [mapReady, setMapReady] = useState(false);
  const [currentCenter, setCurrentCenter] = useState(null);
  const [hoveredMarker, setHoveredMarker] = useState(null);
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
    console.log("lonRange", lonRange);
    if (lonRange > 100) {
      scaleFactor = 300;
      zoomFactor = 1;
      baseSize = 0.7;
    } else if (lonRange > 70) {
      scaleFactor = 350;
      zoomFactor = 2;
      baseSize = 0.5;
    } else if (lonRange > 40) {
      scaleFactor = 400;
      zoomFactor = 2;
      baseSize = 0.3;
    } else if (lonRange > 1) {
      scaleFactor = 500;
      zoomFactor = 3;
      baseSize = 0.2;
    } else if (lonRange > 0.5) {
      scaleFactor = 3000;
      zoomFactor = 2;
      baseSize = 0.3;
    } else if (lonRange > 0.1) {
      scaleFactor = 3000;
      zoomFactor = 5;
      baseSize = 0.15;
    } else {
      scaleFactor = 20000;
      zoomFactor = 5;
      baseSize = 0.15;
    }

    return { center: [centerLon, centerLat], zoom: zoomFactor, scale: scaleFactor, baseSize };
  };

  useEffect(() => {
    if (locations.length > 0 && !hasInitialized.current) {
      const width = window.innerWidth;
      const height = window.innerHeight;
      const bounds = calculateBounds(locations);
      if (bounds) {
        const { center, zoom, scale, baseSize } = calculateMapSettings(bounds, width, height);
        setCurrentCenter(center);
        setZoom(zoom);
        setInitialZoom(zoom);
        setScale(scale);
        setBaseSize(baseSize);
        hasInitialized.current = true;
        setMapReady(true);
      }
    }
  }, [locations]);

  const markerScaleFactor = initialZoom > 0 ? initialZoom / zoom : 1;

  const markerRefs = useRef([]);

  const bringToFront = (index) => {
    const markerGroup = markerRefs.current[index];
    if (markerGroup && markerGroup.parentNode) {
      markerGroup.parentNode.appendChild(markerGroup);
    }
  };

  if (!mapReady || !currentCenter) {
    return <div className="map-loading">Loading map...</div>;
  }

  return (
    <div id="map-container">
      <ComposableMap projectionConfig={{ scale }}>
        <ZoomableGroup
          zoom={zoom}
          center={currentCenter}
          maxZoom={8}
          onMoveEnd={({ coordinates, zoom: currentZoom }) => {
            setCurrentCenter(coordinates);
            setZoom(currentZoom);
          }}
        >
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

          {locations.map((site, index) => {
            const appsInSite = apps.filter((app) => app.location === site.name);
            const allAppsUp = appsInSite.length > 0 && appsInSite.every((app) => app.status === "up");
            const anyAppDown = appsInSite.some((app) => app.status === "down");
            const anyAppUnavailable = appsInSite.some((app) => app.status === "unavailable");

            const color = allAppsUp
              ? "#10B981"
              : anyAppDown
              ? "#EF4444"
              : anyAppUnavailable
              ? "#F59E0B"
              : "#D6D6DA";

            const isHovered = hoveredMarker === site.name;
            const markerScale = baseSize * markerScaleFactor * (isHovered ? 1.5 : 1);

            return (
              <Marker
                key={site.name}
                coordinates={[site.longitude, site.latitude]}
                onClick={() => onSiteClick(site)}
                onMouseEnter={() => {
                  setHoveredMarker(site.name);
                  bringToFront(index);
                }}
                onMouseLeave={() => setHoveredMarker(null)}
              >
                <g ref={(el) => (markerRefs.current[index] = el)} transform={`scale(${markerScale})`} className="marker-wrapper">
                  <g
                    fill={color}
                    stroke="#000000"
                    strokeWidth="0.5"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    transform="translate(-12, -15)"
                    className="marker-icon"
                  >
                    <path d="M12 2C8.13 2 5 5.13 5 9c0 5.25 7 13 7 13s7-7.75 7-13c0-3.87-3.13-7-7-7zm0 9.5c-1.38 0-2.5-1.12-2.5-2.5S10.62 6.5 12 6.5s2.5 1.12 2.5 2.5S13.38 11.5 12 11.5z" />
                  </g>
                  <text
                    className="marker-text"
                    textAnchor="middle"
                    x={0}
                    y={24}
                    fill={color}
                    stroke="#000000"
                    strokeWidth="0.5"  
                    fontWeight="900"  
                    fontSize="22px"
                  >
                    {site.name}
                  </text>
                </g>
              </Marker>
            );
          })}
        </ZoomableGroup>
      </ComposableMap>
    </div>
  );
};
