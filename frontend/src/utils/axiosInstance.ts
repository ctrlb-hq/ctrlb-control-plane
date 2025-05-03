import axios from "axios";

const apiUrl = (import.meta as ImportMetaWithEnv).env.VITE_API_URL;
const API_BASE_URL = `${apiUrl}/api/frontend/v2`;

const axiosInstance = axios.create({
	baseURL: API_BASE_URL,
	headers: {
		Authorization: `${localStorage.getItem("authToken")}`,
	},
});

// Flag to prevent multiple simultaneous refresh token requests
let isRefreshing = false;
let failedQueue: any[] = [];

// Function to process the failed requests queue
const processQueue = (error: any, token: string | null = null) => {
	failedQueue.forEach(prom => {
		if (token) {
			prom.resolve(token);
		} else {
			prom.reject(error);
		}
	});
	failedQueue = [];
};

// Request interceptor to add the access token to headers
axiosInstance.interceptors.request.use(
	config => {
		const token = localStorage.getItem("authToken");
		if (token) {
			config.headers["Authorization"] = `Bearer ${token}`;
		}
		return config;
	},
	error => Promise.reject(error),
);

// Response interceptor to handle 401 errors and refresh the token
axiosInstance.interceptors.response.use(
	response => response,
	async error => {
		const originalRequest = error.config;

		// If the error is a 401 and the request has not already been retried
		if (error.response?.status === 401 && !originalRequest._retry) {
			if (isRefreshing) {
				// If a refresh is already in progress, queue the request
				return new Promise((resolve, reject) => {
					failedQueue.push({ resolve, reject });
				})
					.then(token => {
						originalRequest.headers["Authorization"] = `Bearer ${token}`;
						return axiosInstance(originalRequest);
					})
					.catch(err => Promise.reject(err));
			}

			originalRequest._retry = true;
			isRefreshing = true;

			try {
				const refreshToken = localStorage.getItem("refreshToken");
				const response = await axios.post(`${apiUrl}/api/auth/v1/refresh`, {
					refresh_token: refreshToken,
				});

				const newauthToken = response.data.access_token;
				localStorage.setItem("authToken", newauthToken);

				// Process the queued requests with the new token
				processQueue(null, newauthToken);

				// Retry the original request with the new token
				originalRequest.headers["Authorization"] = `Bearer ${newauthToken}`;
				return axiosInstance(originalRequest);
			} catch (refreshError) {
				// If refreshing the token fails, reject all queued requests
				processQueue(refreshError, null);
				return Promise.reject(refreshError);
			} finally {
				isRefreshing = false;
			}
		}

		return Promise.reject(error);
	},
);

export default axiosInstance;

// Extend ImportMeta interface to include env property
interface ImportMetaWithEnv extends ImportMeta {
	env: {
		VITE_API_URL: string;
	};
}
