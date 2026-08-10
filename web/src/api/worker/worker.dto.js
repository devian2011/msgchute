export class WorkerStateDto {
    constructor(name, backendData) {
        this.name = name;
        this.status = backendData.status;
        this.activeTasks = backendData.active_tasks ?? 0;
        this.activeWorkers = backendData.active_workers ?? 0;
        this.cbState = backendData.cb_state;
    }

    get isCcClosed() {
        return this.cbState === 'Closed' || this.cbState === 0;
    }
}

export class WorkerBoardDto {
    constructor(unwrappedData) {
        // unwrappedData is now directly the raw map[string] object passed from the interceptor
        const rawMap = unwrappedData || {};

        this.workers = Object.entries(rawMap).map(
            ([workerName, workerData]) => new WorkerStateDto(workerName, workerData)
        );
    }

    get totalActiveTasks() {
        return this.workers.reduce((sum, w) => sum + w.activeTasks, 0);
    }
}
