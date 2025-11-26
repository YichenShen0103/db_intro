import axios from "axios";

const api = axios.create({
  baseURL: "/api",
  headers: {
    "Content-Type": "application/json",
  },
});

export const projectsAPI = {
  getAll: () => api.get("/projects"),
  getById: (id) => api.get(`/projects/${id}`),
  create: (data) => api.post("/projects", data),
  dispatch: (id, data) => api.post(`/projects/${id}/dispatch`, data),
  getTracking: (id) => api.get(`/projects/${id}/tracking`),
  remind: (id, data) => api.post(`/projects/${id}/remind`, data),
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
