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
          <span className="user-info-label">Username:</span>
          <div className="user-info-content">
            <span className="user-info-value">
              {user?.username || "Unknown"}
            </span>
          </div>
        </div>
        <div className="user-info-item">
          <span className="user-info-label">Auth Method:</span>
          <div className="user-info-content">
            <span className="user-info-value">
              {user?.auth_method || "Unknown"}
            </span>
          </div>
        </div>
        <div className="user-info-item">
          <span className="user-info-label">Groups:</span>
          <div className="user-info-content user-info-list-container">
            {user?.groups && user.groups.length > 0 ? (
              <div className="user-info-list">
                {user.groups.map((group, index) => (
                  <div key={index} className="user-info-list-item">
                    • {group}
                  </div>
                ))}
              </div>
            ) : (
              <span className="user-info-value">No groups</span>
            )}
          </div>
        </div>
        <div className="user-info-item">
          <span className="user-info-label">Roles:</span>
          <div className="user-info-content user-info-list-container">
            {user?.roles && user.roles.length > 0 ? (
              <div className="user-info-list">
                {user.roles.map((role, index) => (
                  <div key={index} className="user-info-list-item">
                    • {role}
                  </div>
                ))}
              </div>
            ) : (
              <span className="user-info-value">No roles</span>
            )}
          </div>
        </div>

        <div className="user-modal-actions">
          <button onClick={handleLogout} className="auth-submit-btn">
            Logout
          </button>
        </div>
      </div>
    </div>
  );
}
