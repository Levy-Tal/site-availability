import React, { useState, useEffect } from "react";
import { ComposableMap, Geographies, Geography, Marker, ZoomableGroup } from "react-simple-maps";

// Configuration constants
const ZOOM_ADJUSTMENT_FACTOR = 1; // Adjust zoom factor
const MIN_ZOOM = 2; // Minimum zoom level
const MAX_ZOOM = 10; // Maximum zoom level
const HORIZONTAL_MIN_PERCENTAGE = 0.25; // 25% of the total longitude
const HORIZONTAL_MAX_PERCENTAGE = 0.75; // 75% of the total longitude

// Marker size configuration (fixed size)
const MARKER_SIZE = MIN_ZOOM; // Fixed marker size in pixels

export const MapComponent = ({ locations, onSiteClick }) => {
  const geoUrl = "/data/countries-50m.json"; // Your GeoJSON file path

  const [zoom, setZoom] = useState(1); // Initial zoom level
  const [scaleFactor, setScaleFactor] = useState(1); // Track scale factor
  const [center, setCenter] = useState([0, 0]); // Initial center
  const [scale, setScale] = useState(1500); // Initial scale

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

    // Restrict the bounding box horizontally between 25% and 75% of the screen width
    const width = window.innerWidth;
    const height = window.innerHeight;

    const minLongitude = (bounds[0][0] + bounds[1][0]) * HORIZONTAL_MIN_PERCENTAGE;
    const maxLongitude = (bounds[0][0] + bounds[1][0]) * HORIZONTAL_MAX_PERCENTAGE;
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

    return { restrictedCenter, zoomFactor, scale };
  };

  useEffect(() => {
    if (locations.length > 0) {
      const { restrictedBounds, width, height } = calculateBounds(locations);
      const { restrictedCenter, zoomFactor, scale } = calculateZoomAndCenter(restrictedBounds, width, height);

      setCenter(restrictedCenter); // Set initial center, but don't reset it on zoom
      setZoom(zoomFactor); // Set initial zoom
      setScale(scale * 100); // Multiply scale by 10 to make it look normal
    }
  }, [locations]);

  // Ensure the map is displayed when locations are available
  if (!locations || locations.length === 0) {
    return <div>Loading map...</div>;
  }

  // Handle zoom and move events
  const handleMove = ({ x, y, k, dragging }) => {
    console.log("Moving:", x, y, k, dragging);
    setScaleFactor(k); // Update scale factor based on zoom
  };

  return (
    <div
      id="map-container"
      style={{
        width: "100vw",
        height: "100vh",
        position: "relative",
        overflow: "hidden",
      }}
      // Add event listener to prevent default with passive: false
      ref={(div) => {
        if (div) {
          div.addEventListener(
            "wheel",
            (event) => {
              const zoomFactor = event.deltaY > 0 ? 0.9 : 1.1;
              setZoom((prevZoom) => Math.max(0.5, Math.min(prevZoom * zoomFactor, 10)));
              event.preventDefault(); // Prevent default behavior
            },
            { passive: false } // Important to set passive: false
          );
        }
      }}
    >
      <ComposableMap
        projectionConfig={{
          scale: scale, // Use the dynamically calculated scale (with -10 multiplication)
        }}
      >
        <ZoomableGroup
          center={center} // Keep the center as is, don't adjust it on zoom
          zoom={zoom} // Use zoom value from state
          onMove={handleMove} // Use single handler for move event
          onMoveStart={({ coordinates, zoom }) => {
            console.log("Movement started:", coordinates, zoom);
          }}
          onMoveEnd={({ coordinates, zoom }) => {
            setScaleFactor(zoom);
            console.log("Movement ended:", coordinates, zoom);
          }}
        >
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
              {/* Fixed size circle based on the scale factor */}
              <circle r={3/scaleFactor} fill="blue" />
            </Marker>
          ))}
        </ZoomableGroup>
      </ComposableMap>
    </div>
  );
};
