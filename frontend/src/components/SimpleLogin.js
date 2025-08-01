import React, { useState } from "react";
import { useAuth } from "../contexts/AuthContext";
import { oidcLogin } from "../api/authAPI";

export default function SimpleLogin() {
  const { login, loginError, isLoading, authConfig } = useAuth();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");

  const handleSubmit = async (e) => {
    e.preventDefault();
    await login(username, password);
  };

  const handleOIDCLogin = () => {
    oidcLogin();
  };

  // Check what authentication methods are available
  const hasLocal = authConfig?.auth_methods?.includes("local");
  const hasOIDC = authConfig?.auth_methods?.includes("oidc");
  const oidcProviderName = authConfig?.oidc_provider_name || "SSO";

  return (
    <div className="auth-login-container">
      <div className="auth-login-form">
        <h2 className="auth-login-title">Site Availability Login</h2>

        {loginError && <div className="auth-error-message">{loginError}</div>}

        {/* OIDC Login Option */}
        {hasOIDC && (
          <div className="auth-oidc-section">
            <button
              type="button"
              onClick={handleOIDCLogin}
              className="auth-oidc-btn"
              disabled={isLoading}
            >
              Login with {oidcProviderName}
            </button>
          </div>
        )}

        {/* Divider if both methods are available */}
        {hasLocal && hasOIDC && (
          <div className="auth-divider">
            <span>OR</span>
          </div>
        )}

        {/* Local Login Form */}
        {hasLocal && (
          <div className="auth-local-section">
            <form onSubmit={handleSubmit} className="auth-local-form">
              <input
                type="text"
                placeholder="Username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                className="auth-form-field"
                required
              />

              <input
                type="password"
                placeholder="Password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="auth-form-field"
                required
              />

              <button
                type="submit"
                disabled={isLoading}
                className="auth-submit-btn"
              >
                {isLoading ? "Logging in..." : "Login"}
              </button>
            </form>
          </div>
        )}
      </div>
    </div>
  );
}
