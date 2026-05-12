import axios from 'axios';

const API_BASE_URL = 'http://localhost:8080/api';

const api = axios.create({
  baseURL: API_BASE_URL,
});

export const registerPatient = (data) => api.post('/register', data);
export const getRegistrations = () => api.get('/registrations');
export const callPatient = (data) => api.post('/call', data);
export const finishVisit = (id) => api.post(`/registrations/${id}/finish`);

export const createPrescription = (data) => api.post('/prescriptions', data);
export const updatePrescription = (id, data) => api.put(`/prescriptions/${id}`, data);
export const confirmPrescription = (id) => api.post(`/prescriptions/${id}/confirm`);
export const voidPrescription = (id) => api.post(`/prescriptions/${id}/void`);
export const dispensePrescription = (id, data) => api.post(`/prescriptions/${id}/dispense`, data);
export const getPrescriptions = () => api.get('/prescriptions');
export const getPrescriptionsByRegistration = (regId) => api.get(`/registrations/${regId}/prescriptions`);

export const getHerbs = () => api.get('/herbs');
export const updateHerb = (data) => api.put('/herbs', data);

export default api;
