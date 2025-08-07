// Authentication API client for handling all auth-related requests
class AuthAPI {
  constructor() {
    this.baseURL = "/auth";
  }

  async login(username, password) {
    const response = await fetch(`${this.baseURL}/login`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ username, password }),
      credentials: "include",
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || "Login failed");
    }

    return response.json();
  }

  async oidcLogin() {
    // Redirect to OIDC login endpoint
    // The backend will handle the redirect to the OIDC provider
    window.location.href = `${this.baseURL}/oidc/login`;
  }

  async logout() {
    const response = await fetch(`${this.baseURL}/logout`, {
      method: "POST",
      credentials: "include",
    });

    if (!response.ok) {
      throw new Error("Logout failed");
    }

    return response.json();
  }

  async getCurrentUser() {
    const response = await fetch(`${this.baseURL}/user`, {
      credentials: "include",
    });

    if (!response.ok) {
      throw new Error("Failed to get user info");
    }

    return response.json();
  }

  async getAuthConfig() {
    const response = await fetch(`${this.baseURL}/config`);

    if (!response.ok) {
      throw new Error("Failed to get auth config");
    }

    return response.json();
  }

  // Simple session validation - just check if user endpoint works
  async validateSession() {
    try {
      await this.getCurrentUser();
      return { isValid: true };
    } catch (error) {
      return { isValid: false };
    }
  }
}

// Custom error class for authentication-specific errors
class AuthError extends Error {
  constructor(message, code, additionalInfo = {}) {
    super(message);
    this.name = "AuthError";
    this.code = code;
    this.additionalInfo = additionalInfo;
  }

  // Check if this is a rate limiting error
  isRateLimited() {
    return this.code === "RATE_LIMITED";
  }

  // Check if this is an invalid credentials error
  isInvalidCredentials() {
    return this.code === "INVALID_CREDENTIALS";
  }

  // Check if this is a network error
  isNetworkError() {
    return this.code === "NETWORK_ERROR";
  }

  // Get rate limit information if available
  getRateLimitInfo() {
    if (this.isRateLimited()) {
      return this.additionalInfo;
    }
    return null;
  }
}

// Create and export a singleton instance
const authAPI = new AuthAPI();

export { authAPI, AuthError };

// Export individual methods for convenience - bound to maintain context
export const login = authAPI.login.bind(authAPI);
export const logout = authAPI.logout.bind(authAPI);
export const getCurrentUser = authAPI.getCurrentUser.bind(authAPI);
export const validateSession = authAPI.validateSession.bind(authAPI);
export const getAuthConfig = authAPI.getAuthConfig.bind(authAPI);
export const oidcLogin = authAPI.oidcLogin.bind(authAPI);
