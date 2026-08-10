import apiClient from './../client';
import { WorkerBoardDto } from './worker.dto';

export const workerService = {
    async fetchWorkerStatuses() {
        const dataPayload = await apiClient.get('/api/admin/v1/workers/status');

        return new WorkerBoardDto(dataPayload);
    }
};
