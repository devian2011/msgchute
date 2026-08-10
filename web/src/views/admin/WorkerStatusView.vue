<template>
  <BCard title="Worker Performance Monitor" class="shadow-sm border-0 text-start">

    <div v-if="isLoading" class="text-center py-4">
      <BSpinner variant="primary" />
    </div>

    <BAlert v-else-if="error" variant="danger" :model-value="true">
      {{ error }}
    </BAlert>

    <div v-else-if="board" class="mt-3">
      <!-- Суммарная аналитика сверху -->
      <div class="alert alert-secondary py-2 small d-flex justify-content-between mb-3">
        <span>Total Monitored Workers: <strong>{{ board.workers.length }}</strong></span>
        <span>Total Active Tasks: <strong>{{ board.totalActiveTasks }}</strong></span>
      </div>

      <div class="table-responsive">
        <table class="table table-hover align-middle mb-0 small">
          <thead class="table-light">
          <tr>
            <th>Worker Name</th>
            <th>Status</th>
            <th class="text-center">Active Workers</th>
            <th class="text-center">Active Tasks</th>
            <th>Circuit Breaker</th>
          </tr>
          </thead>
          <tbody>
          <tr v-for="worker in board.workers" :key="worker.name">
            <td class="fw-bold text-dark">{{ worker.name }}</td>
            <td>
                <span
                    class="badge text-uppercase"
                    :class="worker.status === 'running' || worker.status === 1 ? 'bg-success' : 'bg-warning'"
                >
                  {{ worker.status }}
                </span>
            </td>
            <td class="text-center">{{ worker.activeWorkers }}</td>
            <td class="text-center font-monospace">{{ worker.activeTasks }}</td>
            <td>
                <span
                    class="badge"
                    :class="worker.isCcClosed ? 'bg-danger' : 'bg-outline-success text-success border border-success'"
                >
                  {{ worker.cbState }}
                </span>
            </td>
          </tr>
          </tbody>
        </table>
      </div>
    </div>

    <BButton
        variant="outline-primary"
        class="mt-3 w-100"
        :disabled="isLoading"
        @click="loadWorkerData"
    >
      Refresh Statuses
    </BButton>
  </BCard>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { workerService } from '@/api/worker/worker.service';

const board = ref(null);
const isLoading = ref(false);
const error = ref(null);

async function loadWorkerData() {
  isLoading.value = true;
  error.value = null;
  try {
    board.value = await workerService.fetchWorkerStatuses();
  } catch (err) {
    error.value = err.message || 'Failed to fetch worker configurations';
  } finally {
    isLoading.value = false;
  }
}

onMounted(() => {
  loadWorkerData();
});
</script>
