import React, { useState, useEffect } from "react";
import { ComposableMap, Geographies, Geography, Marker, ZoomableGroup } from "react-simple-maps";

// Configuration constants
const ZOOM_ADJUSTMENT_FACTOR = 1; // 15% more zoom
const MIN_ZOOM = 2; // Minimum zoom level
const MAX_ZOOM = 4; // Maximum zoom level
const HORIZONTAL_MIN_PERCENTAGE = 0.25; // 10% of the total longitude
const HORIZONTAL_MAX_PERCENTAGE = 0.75; // 90% of the total longitude

export const MapComponent = ({ locations, onSiteClick }) => {
  const geoUrl = "/data/countries-50m.json"; // Your GeoJSON file path

  const [zoom, setZoom] = useState(MIN_ZOOM); // Initial zoom level
  const [center, setCenter] = useState([0, 0]); // Initial center

  // Function to calculate the bounding box and apply restrictions
  const calculateBounds = (locations) => {
    const bounds = locations.reduce(
      (acc, loc) => {
        acc[0][0] = Math.min(acc[0][0], loc.Longitude);
        acc[0][1] = Math.min(acc[0][1], loc.Latitude);
        acc[1][0] = Math.max(acc[1][0], loc.Longitude);
        acc[1][1] = Math.max(acc[1][1], loc.Latitude);
        return acc;
      },
      [[Infinity, Infinity], [-Infinity, -Infinity]]
    );

    // Restrict the bounding box horizontally between 10% and 90% of the screen width
    const width = window.innerWidth;
    const height = window.innerHeight;

    const minLongitude = (bounds[0][0] + bounds[1][0]) * HORIZONTAL_MIN_PERCENTAGE;  // 10% of the total longitude width
    const maxLongitude = (bounds[0][0] + bounds[1][0]) * HORIZONTAL_MAX_PERCENTAGE;  // 90% of the total longitude width
    const minLatitude = bounds[0][1];
    const maxLatitude = bounds[1][1];

    const restrictedBounds = [
      [minLongitude, minLatitude],
      [maxLongitude, maxLatitude]
    ];

    return { restrictedBounds, width, height };
  };

  // Function to calculate the zoom and center based on the bounding box
  const calculateZoomAndCenter = (restrictedBounds, width, height) => {
    // Calculate the center of the restricted bounding box
    const restrictedCenter = [
      (restrictedBounds[0][0] + restrictedBounds[1][0]) / 2,
      (restrictedBounds[0][1] + restrictedBounds[1][1]) / 2,
    ];

    // Calculate the zoom level to fit the restricted bounding box (zoom as much as possible)
    const scale = Math.min(
      width / (restrictedBounds[1][0] - restrictedBounds[0][0]),
      height / (restrictedBounds[1][1] - restrictedBounds[0][1])
    );

    // Adjust zoom level
    const zoomFactor = Math.max(MIN_ZOOM, Math.min(scale * ZOOM_ADJUSTMENT_FACTOR, MAX_ZOOM));

    return { restrictedCenter, zoomFactor };
  };

  useEffect(() => {
    if (locations.length > 0) {
      const { restrictedBounds, width, height } = calculateBounds(locations);
      const { restrictedCenter, zoomFactor } = calculateZoomAndCenter(restrictedBounds, width, height);

      setCenter(restrictedCenter);
      setZoom(zoomFactor);
    }
  }, [locations]);

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
      onWheel={(event) => {
        const zoomFactor = event.deltaY > 0 ? 0.9 : 1.1;
        setZoom((prevZoom) => Math.max(MIN_ZOOM, Math.min(prevZoom * zoomFactor, MAX_ZOOM)));
        event.preventDefault();
      }}
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
