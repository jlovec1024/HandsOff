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

        if (isLoginPage) {
          // On login page, let the component handle the error
          message.error(data?.error || "Authentication failed");
          return Promise.reject(error);
        }

        // On other pages, clear auth state and redirect to login
        const { clearAuth } = useAuthStore.getState();
        clearAuth();
        message.error("Session expired, please login again");
        window.location.href = ROUTES.LOGIN;
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
