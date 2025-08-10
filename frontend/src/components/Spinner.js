import React from "react";

export default function Spinner({ size = "md", label = "Loading" }) {
  const sizeClass =
    size === "lg" ? "sa-spinner--lg" : size === "sm" ? "sa-spinner--sm" : "";

  return (
    <div role="status" aria-label={label} className="sa-spinner-container">
      <div className={`sa-spinner ${sizeClass}`} />
    </div>
  );
}
