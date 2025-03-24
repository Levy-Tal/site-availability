import React, { useState, useEffect } from "react";
import { ComposableMap, Geographies, Geography, Marker, ZoomableGroup } from "react-simple-maps";

// Configuration constants
const MIN_ZOOM = 2;
const MAX_ZOOM = 10;

export const MapComponent = ({ locations, onSiteClick }) => {
  const geoUrl = "/data/countries-50m.json";
  const [zoom, setZoom] = useState(1);
  const [center, setCenter] = useState([0, 0]);

  // Calculate bounding box
  const calculateBounds = (locations) => {
    if (!locations.length) return null;
    return locations.reduce(
      (acc, loc) => {
        acc[0][0] = Math.min(acc[0][0], loc.Longitude);
        acc[0][1] = Math.min(acc[0][1], loc.Latitude);
        acc[1][0] = Math.max(acc[1][0], loc.Longitude);
        acc[1][1] = Math.max(acc[1][1], loc.Latitude);
        return acc;
      },
      [[Infinity, Infinity], [-Infinity, -Infinity]]
    );
  };

  useEffect(() => {
    if (locations.length > 0) {
      const bounds = calculateBounds(locations);
      if (bounds) {
        const center = [
          (bounds[0][0] + bounds[1][0]) / 2,
          (bounds[0][1] + bounds[1][1]) / 2,
        ];
        setCenter(center);
        setZoom(Math.max(MIN_ZOOM, Math.min(MAX_ZOOM, 4))); // Adjusted zoom calculation
      }
    }
  }, [locations]);

  if (!locations.length) return <div>Loading map...</div>;

  return (
    <div
      id="map-container"
      style={{ width: "100vw", height: "100vh", position: "relative", overflow: "hidden" }}
    >
      <ComposableMap>
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
              <circle r={3 + zoom / 10} fill="blue" />
            </Marker>
          ))}
        </ZoomableGroup>
      </ComposableMap>
    </div>
  );
};
