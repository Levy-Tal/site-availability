import React, { useState } from "react";
import { ComposableMap, Geographies, Geography, Marker, ZoomableGroup } from "react-simple-maps";

export const MapComponent = ({ locations, onSiteClick }) => {
  const geoUrl = "/data/countries-50m.json"; // Your GeoJSON file path

  const [zoom, setZoom] = useState(1); // Initial zoom level (1 is the default)

  // Handle mouse wheel zoom
  const handleWheel = (event) => {
    const zoomFactor = event.deltaY > 0 ? 0.9 : 1.1; // Zoom out if scroll down, zoom in if scroll up
    setZoom((prevZoom) => Math.max(0.5, Math.min(prevZoom * zoomFactor, 10))); // Limit zoom between 0.5x and 10x
    event.preventDefault(); // Prevent the default scrolling behavior
  };

  // Ensure the map is displayed when locations are available
  if (!locations || locations.length === 0) {
    return <div>Loading map...</div>;
  }

  return (
    <div
      id="map-container"
      style={{
        width: "100vw",
        height: "100vh",
        position: "relative",
        overflow: "hidden",
      }}
      onWheel={handleWheel} // Add the wheel event listener here
    >
      <ComposableMap>
        <ZoomableGroup zoom={zoom}>
          <Geographies geography={geoUrl}>
            {({ geographies }) =>
              geographies.map((geo) => (
                <Geography
                  key={geo.rsmKey}
                  geography={geo}
                  style={{
                    default: {
                      fill: "#D6D6DA", // Default fill color
                      outline: "none", // Disable the outline
                    },
                    hover: {
                      fill: "#D6D6DA", // Hover effect
                      outline: "none", // Disable hover outline
                    },
                    pressed: {
                      fill: "#D6D6DA", // Clicked effect
                      outline: "none", // Disable clicked outline
                    },
                  }}
                />
              ))
            }
          </Geographies>
          {locations.map((site) => (
            <Marker
              key={site.name}
              coordinates={[site.Longitude, site.Latitude]}
              onClick={() => onSiteClick(site)}
            >
              <circle r={5} fill="red" />
            </Marker>
          ))}
        </ZoomableGroup>
      </ComposableMap>
    </div>
  );
};
