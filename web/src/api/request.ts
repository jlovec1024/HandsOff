import axios, { AxiosError, type InternalAxiosRequestConfig } from "axios";
import { message } from "antd";
import { useAuthStore } from "../stores/auth";
import { ROUTES } from "../constants/routes";

// Create axios instance
const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || "http://localhost:8080/api",
  timeout: 30000,
});

// Request interceptor - add token to headers
request.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = localStorage.getItem("token");
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor - handle errors
request.interceptors.response.use(
  (response) => {
    return response;
  },
  (error: AxiosError<{ error: string }>) => {
    if (error.response) {
      const { status, data } = error.response;

      // Handle 401 Unauthorized
      if (status === 401) {
        const isLoginPage = window.location.pathname === ROUTES.LOGIN;
        const errorMsg = isLoginPage
          ? data?.error || "Authentication failed"
          : "Session expired, please login again";

        message.error(errorMsg);

        // Redirect to login only if not already on login page (prevents infinite loop)
        if (!isLoginPage) {
          const { clearAuth } = useAuthStore.getState();
          clearAuth();
          window.location.href = ROUTES.LOGIN;
        }

        return Promise.reject(error);
      }

      // Handle other errors
      const errorMsg = data?.error || "An error occurred";
      message.error(errorMsg);
    } else if (error.request) {
      message.error("Network error, please check your connection");
    } else {
      message.error("Request failed");
    }

    return Promise.reject(error);
  }
);

export default request;
