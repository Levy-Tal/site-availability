import React, {
  createContext,
  useContext,
  useReducer,
  useCallback,
  useEffect,
} from "react";
import { authAPI } from "../api/authAPI";
import Spinner from "../components/Spinner";

const AuthContext = createContext();

// Auth actions
const AUTH_ACTIONS = {
  INIT_START: "INIT_START",
  INIT_SUCCESS: "INIT_SUCCESS",
  INIT_ERROR: "INIT_ERROR",
  SET_AUTH_CONFIG: "SET_AUTH_CONFIG",
  LOGIN_START: "LOGIN_START",
  LOGIN_SUCCESS: "LOGIN_SUCCESS",
  LOGIN_ERROR: "LOGIN_ERROR",
  LOGOUT_START: "LOGOUT_START",
  LOGOUT_SUCCESS: "LOGOUT_SUCCESS",
  LOGOUT_ERROR: "LOGOUT_ERROR",
  SET_USER: "SET_USER",
  CLEAR_ERROR: "CLEAR_ERROR",
};

// Initial state
const initialState = {
  isInitialized: false,
  isLoading: false,
  isAuthenticated: false,
  needsAuthentication: false,
  user: null,
  session: null,
  authConfig: null,
  error: null,
  loginError: null,
};

// Auth reducer
function authReducer(state, action) {
  switch (action.type) {
    case AUTH_ACTIONS.INIT_START:
      return {
        ...state,
        isLoading: true,
        error: null,
      };

    case AUTH_ACTIONS.INIT_SUCCESS:
      return {
        ...state,
        isLoading: false,
        isInitialized: true,
        isAuthenticated: action.payload.isAuthenticated,
        needsAuthentication:
          action.payload.authConfig?.auth_enabled &&
          !action.payload.isAuthenticated,
        user: action.payload.user,
        session: action.payload.session,
        authConfig: action.payload.authConfig,
        error: null,
      };

    case AUTH_ACTIONS.INIT_ERROR:
      return {
        ...state,
        isLoading: false,
        isInitialized: true,
        error: action.payload,
      };

    case AUTH_ACTIONS.SET_AUTH_CONFIG:
      return {
        ...state,
        authConfig: action.payload,
      };

    case AUTH_ACTIONS.LOGIN_START:
      return {
        ...state,
        isLoading: true,
        loginError: null,
      };

    case AUTH_ACTIONS.LOGIN_SUCCESS:
      return {
        ...state,
        isLoading: false,
        isAuthenticated: true,
        needsAuthentication: false,
        user: action.payload.user,
        session: action.payload.session,
        loginError: null,
      };

    case AUTH_ACTIONS.LOGIN_ERROR:
      return {
        ...state,
        isLoading: false,
        isAuthenticated: false,
        user: null,
        session: null,
        loginError: action.payload.error,
      };

    case AUTH_ACTIONS.LOGOUT_START:
      return {
        ...state,
        isLoading: true,
      };

    case AUTH_ACTIONS.LOGOUT_SUCCESS:
      return {
        ...state,
        isLoading: false,
        isAuthenticated: false,
        needsAuthentication: state.authConfig?.auth_enabled || false,
        user: null,
        session: null,
        loginError: null,
      };

    case AUTH_ACTIONS.SET_USER:
      return {
        ...state,
        user: action.payload.user,
        session: action.payload.session,
        isAuthenticated: action.payload.user !== null,
      };

    case AUTH_ACTIONS.CLEAR_ERROR:
      return {
        ...state,
        error: null,
        loginError: null,
      };

    default:
      return state;
  }
}

// Auth Provider component
export function AuthProvider({ children }) {
  const [state, dispatch] = useReducer(authReducer, initialState);

  // Use the authAPI singleton instance

  // Initialize authentication state
  const initialize = useCallback(async () => {
    dispatch({ type: AUTH_ACTIONS.INIT_START });
    try {
      // First, get authentication configuration
      const authConfig = await authAPI.getAuthConfig();

      if (!authConfig.auth_enabled) {
        // Authentication is disabled - allow access without login
        dispatch({
          type: AUTH_ACTIONS.INIT_SUCCESS,
          payload: {
            isAuthenticated: false,
            user: null,
            session: null,
            authConfig: authConfig,
          },
        });
        return;
      }

      // Authentication is enabled - check if there's an existing session
      try {
        const sessionValidation = await authAPI.validateSession();
        if (sessionValidation.isValid) {
          // Session exists and is valid - get user info
          const userInfo = await authAPI.getCurrentUser();

          dispatch({
            type: AUTH_ACTIONS.INIT_SUCCESS,
            payload: {
              isAuthenticated: true,
              user: userInfo.user,
              session: userInfo.session,
              authConfig: authConfig,
            },
          });
        } else {
          // No valid session - user needs to login
          dispatch({
            type: AUTH_ACTIONS.INIT_SUCCESS,
            payload: {
              isAuthenticated: false,
              user: null,
              session: null,
              authConfig: authConfig,
            },
          });
        }
      } catch (sessionError) {
        // Session validation failed silently - just mark as not authenticated
        dispatch({
          type: AUTH_ACTIONS.INIT_SUCCESS,
          payload: {
            isAuthenticated: false,
            user: null,
            session: null,
            authConfig: authConfig,
          },
        });
      }
    } catch (error) {
      console.error("Authentication initialization error:", error);
      dispatch({
        type: AUTH_ACTIONS.INIT_ERROR,
        payload: error.message || "Failed to initialize authentication",
      });
    }
  }, []);

  // Login function
  const login = useCallback(async (username, password) => {
    dispatch({ type: AUTH_ACTIONS.LOGIN_START });

    try {
      const loginResponse = await authAPI.login(username, password);

      if (loginResponse.success) {
        // Get user information
        const userInfo = await authAPI.getCurrentUser();

        dispatch({
          type: AUTH_ACTIONS.LOGIN_SUCCESS,
          payload: {
            user: userInfo.user,
            session: userInfo.session,
          },
        });

        return { success: true };
      } else {
        throw new Error(loginResponse.message || "Login failed");
      }
    } catch (error) {
      dispatch({
        type: AUTH_ACTIONS.LOGIN_ERROR,
        payload: { error: error.message || "Login failed" },
      });
      return { success: false, error: error.message || "Login failed" };
    }
  }, []);

  // Logout function
  const logout = useCallback(async () => {
    dispatch({ type: AUTH_ACTIONS.LOGOUT_START });

    try {
      await authAPI.logout();
      dispatch({ type: AUTH_ACTIONS.LOGOUT_SUCCESS });
      return { success: true };
    } catch (error) {
      dispatch({
        type: AUTH_ACTIONS.LOGOUT_ERROR,
        payload: error.message || "Logout failed",
      });
      return { success: false, error: error.message || "Logout failed" };
    }
  }, []);

  // Clear error function
  const clearError = useCallback(() => {
    dispatch({ type: AUTH_ACTIONS.CLEAR_ERROR });
  }, []);

  // Initialize on mount
  useEffect(() => {
    initialize();
  }, [initialize]);

  // Context value
  const value = {
    // State
    ...state,

    // Actions
    login,
    logout,
    clearError,

    // Helpers
    isAdmin: state.user?.is_admin || false,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

// Custom hook to use auth context
export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}

// Export the AuthContext for direct access if needed
export { AuthContext };

// HOC for components that require authentication
export function withAuth(WrappedComponent) {
  return function AuthenticatedComponent(props) {
    const auth = useAuth();

    if (!auth.isInitialized) {
      return (
        <div
          style={{
            position: "fixed",
            top: 0,
            left: 0,
            width: "100vw",
            height: "100vh",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            zIndex: 2000,
          }}
        >
          <Spinner size="lg" label="Initializing" />
        </div>
      );
    }

    if (auth.needsAuthentication) {
      // This will be handled by the main App component
      return null;
    }

    return <WrappedComponent {...props} auth={auth} />;
  };
}

export default AuthContext;
