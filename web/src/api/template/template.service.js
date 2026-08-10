import apiClient from './../client';
import qs from 'qs';

export const templateService = {
    async getTemplates(page = 1, perPage = 10, filters = {}) {
        const cleanedFilters = {};
        Object.keys(filters).forEach(key => {
            if (Array.isArray(filters[key]) && filters[key].length > 0) {
                cleanedFilters[key] = filters[key];
            } else if (filters[key] && !Array.isArray(filters[key])) {
                cleanedFilters[key] = filters[key];
            }
        });

        const responseData = await apiClient.get('/api/admin/v1/template', {
            params: {
                page,
                per_page: perPage,
                ...cleanedFilters
            },
            paramsSerializer: params => {
                return qs.stringify(params, { arrayFormat: 'repeat' });
            }
        });

        return responseData;
    },

    async updateTemplate(code, templatePayload) {
        return await apiClient.put(`/api/admin/v1/template/${encodeURIComponent(code)}`, templatePayload);
    },

    async createTemplate(code, templatePayload) {
        return await apiClient.post(`/api/admin/v1/template`, templatePayload);
    }
};
