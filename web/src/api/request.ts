import axios, { AxiosError, type InternalAxiosRequestConfig } from "axios";
import { message } from "antd";

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

      // Handle 401 Unauthorized - redirect to login
      if (status === 401) {
        localStorage.removeItem("token");
        localStorage.removeItem("user");
        window.location.href = "/login";
        message.error("Session expired, please login again");
      }
      // Handle other errors
      else {
        const errorMsg = data?.error || "An error occurred";
        message.error(errorMsg);
      }
    } else if (error.request) {
      message.error("Network error, please check your connection");
    } else {
      message.error("Request failed");
    }

    return Promise.reject(error);
  }
);

export default request;
