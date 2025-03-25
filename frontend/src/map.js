import React, { useState, useEffect, useRef } from "react";
import { ComposableMap, Geographies, Geography, Marker, ZoomableGroup } from "react-simple-maps";

const INITIAL_ZOOM = 5;
const DEFAULT_SCALE = 1000;

export const MapComponent = ({ locations, onSiteClick }) => {
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
        minLon: Math.min(acc.minLon, loc.Longitude),
        minLat: Math.min(acc.minLat, loc.Latitude),
        maxLon: Math.max(acc.maxLon, loc.Longitude),
        maxLat: Math.max(acc.maxLat, loc.Latitude),
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
          {locations.map((site) => (
            <Marker key={site.name} coordinates={[site.Longitude, site.Latitude]} onClick={() => onSiteClick(site)}>
              <text x={0} y={-2} fill="#fff" fontSize={5} textAnchor="middle" fontWeight="bold" style={{ textShadow: "1px 1px 2px rgba(0, 0, 0, 0.5)" }}>{site.name}</text>
              <circle r={1} fill="blue" />
            </Marker>
          ))}
        </ZoomableGroup>
      </ComposableMap>
    </div>
  );
};
