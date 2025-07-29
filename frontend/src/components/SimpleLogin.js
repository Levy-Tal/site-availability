import React, { useState } from "react";
import { useAuth } from "../contexts/AuthContext";

export default function SimpleLogin() {
  const { login, loginError, isLoading } = useAuth();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");

  const handleSubmit = async (e) => {
    e.preventDefault();
    await login(username, password);
  };

  return (
    <div className="auth-login-container">
      <form onSubmit={handleSubmit} className="auth-login-form">
        <h2 className="auth-login-title">Site Availability Login</h2>

        {loginError && <div className="auth-error-message">{loginError}</div>}

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

        <button type="submit" disabled={isLoading} className="auth-submit-btn">
          {isLoading ? "Logging in..." : "Login"}
        </button>
      </form>
    </div>
  );
}
