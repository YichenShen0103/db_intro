import axios from "axios";

const api = axios.create({
  baseURL: "/api",
});

// Add a request interceptor to inject the token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem("token");
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

export const authAPI = {
  login: (data) => api.post("/login", data),
  register: (data) => api.post("/register", data),
};

export const projectsAPI = {
  getAll: () => api.get("/projects"),
  getById: (id) => api.get(`/projects/${id}`),
  create: (data) => api.post("/projects", data),
  addMembers: (id, data) => api.post(`/projects/${id}/members`, data),
  dispatch: (id) => api.post(`/projects/${id}/dispatch`),
  getTracking: (id) => api.get(`/projects/${id}/tracking`),
  remind: (id, data) => api.post(`/projects/${id}/remind`, data),
  fetchEmails: (id) => api.post(`/projects/${id}/fetch-emails`),
  aggregate: (id) => api.post(`/projects/${id}/aggregate`),
  download: (id) =>
    api.get(`/projects/${id}/download`, { responseType: "blob" }),
};

export const teachersAPI = {
  getAll: (params) => api.get("/teachers", { params }),
  create: (data) => api.post("/teachers", data),
  update: (id, data) => api.put(`/teachers/${id}`, data),
  delete: (id) => api.delete(`/teachers/${id}`),
};

export const departmentsAPI = {
  getAll: () => api.get("/departments"),
};

export default api;
