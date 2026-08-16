import axios from "axios";

const isServer = typeof window === "undefined";
const baseURL = isServer
  ? process.env.NEXT_SERVER_API_URL || "http://127.0.0.1:8080/api/v1"
  : process.env.NEXT_PUBLIC_API_URL || "/api/v1";

const apiClient = axios.create({
  baseURL,
  timeout: 160000, // Increased to allow local AI models (Ollama) time to respond
  withCredentials: true, // Important for cookies (Session Auth)
});

// Response interceptor to handle errors uniformly
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    // Standardize error message
    if (typeof error.response?.data?.error === "string") {
      error.message = error.response.data.error;
    } else if (error.response?.data?.error?.message) {
      error.message = error.response.data.error.message;
    }

    return Promise.reject(error);
  }
);

export default apiClient;
