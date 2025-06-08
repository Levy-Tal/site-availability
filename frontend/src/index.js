import React from "react";
import ReactDOM from "react-dom/client"; // Change this import
import App from "./App";

const root = ReactDOM.createRoot(document.getElementById("root")); // Create root using createRoot from react-dom/client
root.render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
