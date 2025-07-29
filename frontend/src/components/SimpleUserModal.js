import React from "react";
import { useAuth } from "../contexts/AuthContext";

export default function SimpleUserModal({ isOpen, onClose }) {
  const { user, logout } = useAuth();

  if (!isOpen) return null;

  const handleLogout = async () => {
    await logout();
    onClose();
  };

  return (
    <div className="user-modal-overlay" onClick={onClose}>
      <div className="user-modal-content" onClick={(e) => e.stopPropagation()}>
        <div className="user-modal-header">
          <h3 className="user-modal-title">User Information</h3>
          <button onClick={onClose} className="user-modal-close">
            ×
          </button>
        </div>

        <div className="user-info-item">
          <span className="user-info-label">Username:</span>{" "}
          {user?.username || "Unknown"}
        </div>
        <div className="user-info-item">
          <span className="user-info-label">Auth Method:</span>{" "}
          {user?.auth_method || "Unknown"}
        </div>
        <div className="user-info-item">
          <span className="user-info-label">Groups:</span>{" "}
          {user?.groups && user.groups.length > 0
            ? user.groups.join(", ")
            : "No groups"}
        </div>
        <div className="user-info-item">
          <span className="user-info-label">Roles:</span>{" "}
          {user?.roles && user.roles.length > 0
            ? user.roles.join(", ")
            : "No roles"}
        </div>

        <div className="user-modal-actions">
          <button
            onClick={onClose}
            className="user-modal-btn user-modal-btn--secondary"
          >
            Close
          </button>
          <button
            onClick={handleLogout}
            className="user-modal-btn user-modal-btn--danger"
          >
            Logout
          </button>
        </div>
      </div>
    </div>
  );
}
