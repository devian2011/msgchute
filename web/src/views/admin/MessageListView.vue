<template>
  <BCard title="Messages History" class="shadow border-0 rounded-3 text-start">

    <div v-if="isLoading" class="text-center py-4">
      <BSpinner variant="primary"/>
    </div>

    <BAlert v-else-if="error" variant="danger" :model-value="true">
      {{ error }}
    </BAlert>

    <div v-else-if="resultData">

      <BCard class="border-0 shadow-sm rounded-3 mb-4 bg-white text-start">
        <div class="d-flex justify-content-between align-items-center mb-3 border-bottom pb-2">
          <h5 class="m-0 fw-bold text-dark fs-6">
            <i class="bi bi-funnel-fill text-primary me-1"></i> Search & Filter Query
          </h5>
          <BButton variant="link" size="sm" class="text-decoration-none p-0" @click="resetFilters">
            Reset All
          </BButton>
        </div>

        <BForm @submit.prevent="applyFilters">
          <BRow class="g-3">

            <BCol md="6">
              <BFormGroup label="Message UUIDs:" label-for="filter-ids" class="small fw-semibold text-muted">
                <BFormInput
                    id="filter-ids"
                    v-model="filterForm.idsRaw"
                    type="text"
                    placeholder="Comma separated UUIDs..."
                    class="custom-select2-input"
                />
              </BFormGroup>
            </BCol>

            <BCol md="6">
              <BFormGroup label="Recipients Targets (Async Search):" class="small fw-semibold text-muted">
                <Multiselect
                    v-model="filterForm.recipient"
                    mode="tags"
                    placeholder="Type email or phone to search..."
                    :searchable="true"
                    :close-on-select="false"
                    :clear-on-select="false"
                    :delay="300"
                    :resolve-on-load="false"
                    :object="false"
                    :options="async (query) => await searchRecipients(query)"
                    class="custom-select2"
                />
              </BFormGroup>
            </BCol>

            <BCol md="4">
              <BFormGroup label="Transports Providers:" class="small fw-semibold text-muted">
                <Multiselect
                    v-model="filterForm.transport"
                    mode="tags"
                    placeholder="Select transports..."
                    :options="filterDictionaries?.transports || []"
                    :searchable="true"
                    :close-on-select="false"
                    :clear-on-select="false"
                    class="custom-select2"
                />
              </BFormGroup>
            </BCol>

            <BCol md="4">
              <BFormGroup label="Sender Identities:" class="small fw-semibold text-muted">
                <Multiselect
                    v-model="filterForm.sender_ids"
                    mode="tags"
                    placeholder="Select senders..."
                    :options="filterDictionaries?.senderIds || []"
                    :searchable="true"
                    :close-on-select="false"
                    :clear-on-select="false"
                    class="custom-select2"
                />
              </BFormGroup>
            </BCol>

            <BCol md="4">
              <BFormGroup label="Template Codes:" class="small fw-semibold text-muted">
                <Multiselect
                    v-model="filterForm.code"
                    mode="tags"
                    placeholder="Select templates..."
                    :options="filterDictionaries?.templates || []"
                    :searchable="true"
                    :close-on-select="false"
                    :clear-on-select="false"
                    class="custom-select2"
                />
              </BFormGroup>
            </BCol>

            <BCol cols="12" class="border-top pt-2">
              <BFormGroup label="Payload Engine Statuses:" class="small fw-semibold text-muted">
                <div class="d-flex flex-wrap gap-2 mt-1">
                  <BFormCheckbox
                      v-for="status in ['running', 'succeeded', 'failed', 'declined']"
                      :key="status"
                      v-model="filterForm.status"
                      :value="status"
                      class="d-inline-block me-3 text-dark fw-medium text-capitalize"
                      style="font-size: 0.85rem;"
                  >
                    {{ status }}
                  </BFormCheckbox>
                </div>
              </BFormGroup>
            </BCol>

          </BRow>

          <div class="d-flex justify-content-end mt-3 border-top pt-2">
            <BButton type="submit" variant="primary" size="sm" class="px-4 shadow-sm" :disabled="isLoading">
              <BSpinner v-if="isLoading" small class="me-1"></BSpinner>
              Search
            </BButton>
          </div>
        </BForm>
      </BCard>

      <div class="table-responsive">
        <table class="table align-middle small">
          <thead class="table-dark">
          <tr>
            <th>Message Info</th>
            <th>Transport</th>
            <th>Payload Status</th>
            <th>Tasks Status</th>
            <th>Actions</th>
          </tr>
          </thead>
          <tbody>
          <!-- 1. Итерируем список всех оберток FullMessageInfoDto -->
          <template v-for="item in resultData.messages" :key="item.message.id">

            <!-- Основная строка с базовой информацией о сообщении -->
            <tr :class="{ 'table-active': expandedMessageId === item.message.id }">
              <td>
                <div class="fw-bold text-dark font-monospace text-truncate" style="max-width: 150px;"
                     :title="item.message.id">
                  {{ item.message.id }}
                </div>
                <div class="text-secondary fw-semibold">{{ item.message.subject }}</div>
                <small class="text-muted d-block">Sender: {{ item.message.senderId }}</small>
              </td>
              <td>
                <span class="badge bg-dark text-uppercase px-2">{{ item.message.transport }}</span>
              </td>
              <td>
                <span class="badge" :class="getStatusBadgeClass(item.message.status)">{{ item.message.status }}</span>
              </td>
              <td>
                <span class="badge bg-light text-dark border">Tasks: {{ item.tasks.length }}</span>
              </td>
              <td>
                <BButton variant="outline-primary" size="sm" @click="toggleExpand(item.message.id)">
                  {{ expandedMessageId === item.message.id ? 'Hide Audit' : 'Inspect' }}
                </BButton>
                <BButton
                    :to="`/admin/message/${item.message.id}`"
                    variant="dark"
                    size="sm"
                >
                  Detail
                </BButton>
              </td>
            </tr>

            <!-- 2. Развернутая панель аудита: показывается только для выбранного сообщения -->
            <tr v-if="expandedMessageId === item.message.id">
              <td colspan="5" class="bg-light p-4">
                <div class="border rounded bg-white p-4 shadow-sm">

                  <!-- РАЗДЕЛ А: ДАННЫЕ ЗАПРОСА И СТРАТЕГИЯ ПОВТОРОВ СООБЩЕНИЯ -->
                  <h6 class="fw-bold mb-3 text-dark border-bottom pb-2">1. Initial Request Payload & Message
                    Policies</h6>
                  <BRow class="mb-4 g-3">
                    <BCol md="4">
                      <div class="bg-light rounded p-2 h-100">
                        <span class="fw-bold text-secondary d-block mb-1 small text-uppercase">Template Params:</span>
                        <pre class="bg-white border rounded p-2 m-0 font-monospace json-block">{{
                            JSON.stringify(item.message.params, null, 2)
                          }}</pre>
                      </div>
                    </BCol>
                    <BCol md="4">
                      <div class="bg-light rounded p-2 h-100">
                        <span class="fw-bold text-secondary d-block mb-1 small text-uppercase">Meta / Headers:</span>
                        <pre class="bg-white border rounded p-2 m-0 font-monospace json-block">{{
                            JSON.stringify(item.message.meta, null, 2)
                          }}</pre>
                      </div>
                    </BCol>
                    <BCol md="4">
                      <div class="bg-light rounded p-2 h-100">
                        <span class="fw-bold text-secondary d-block mb-1 small text-uppercase">Global Retry Policy (*Retry):</span>
                        <pre class="bg-white border rounded p-2 m-0 font-monospace json-block"
                             v-if="item.message.retry">{{ JSON.stringify(item.message.retry, null, 2) }}</pre>
                        <div class="bg-white border rounded p-2 m-0 text-muted fst-italic small" v-else>No global retry
                          rules defined
                        </div>
                      </div>
                    </BCol>
                    <BCol md="6">
                      <div class="bg-light rounded p-2">
                        <span class="fw-bold text-secondary d-block mb-1 small text-uppercase">Recipients Data:</span>
                        <pre class="bg-white border rounded p-2 m-0 font-monospace json-block">{{
                            JSON.stringify(item.message.recipients, null, 2)
                          }}</pre>
                      </div>
                    </BCol>
                    <BCol md="6">
                      <div class="bg-light rounded p-2">
                        <span class="fw-bold text-secondary d-block mb-1 small text-uppercase">Deadlines & Schedule Timestamps:</span>
                        <div class="bg-white border rounded p-2 small font-monospace">
                          <div><strong>Template Code:</strong> {{ item.message.code || 'None (Direct Body)' }}</div>
                          <div><strong>Deadline:</strong>
                            {{ item.message.deadline ? item.message.deadline.toLocaleString() : 'N/A' }}
                          </div>
                          <div><strong>Schedule:</strong>
                            {{ item.message.schedule ? item.message.schedule.toLocaleString() : 'Immediate dispatch' }}
                          </div>
                        </div>
                      </div>
                    </BCol>
                  </BRow>

                  <!-- РАЗДЕЛ Б: СПИСОК СВЯЗАННЫХ ЗАДАЧ (Итерируем массив структур FullTask) -->
                  <h6 class="fw-bold mb-3 text-dark border-bottom pb-2">2. Async Retrier Processing Tasks & Backoff
                    Configuration</h6>

                  <div v-for="fullTask in item.tasks" :key="fullTask.task.id"
                       class="mb-4 p-3 border rounded bg-light-subtle">
                    <div class="d-flex justify-content-between bg-light p-2 rounded mb-3 flex-wrap gap-2">
                      <div>
                        <span class="badge bg-primary me-2">Worker: {{ fullTask.task.worker }}</span>
                        <span class="text-dark fw-semibold small font-monospace">Task ID: {{ fullTask.task.id }}</span>
                      </div>
                      <div class="d-flex align-items-center gap-2">
                  <span class="badge" :class="fullTask.task.isProcessed ? 'bg-success' : 'bg-secondary'">
                    {{ fullTask.task.isProcessed ? 'Processed' : 'Unprocessed' }}
                  </span>
                        <span class="badge" :class="getStatusBadgeClass(fullTask.task.status)">Task Status: {{
                            fullTask.task.status
                          }}</span>
                      </div>
                    </div>

                    <!-- Детальные параметры Backoff и планировщика задачи -->
                    <BRow class="mb-3 g-2 px-2">
                      <BCol md="3">
                        <span class="small text-secondary fw-semibold text-uppercase d-block">Attempts Progress:</span>
                        <div class="fw-bold font-monospace fs-5 text-dark">{{ fullTask.task.retries }} /
                          {{ fullTask.task.maxRetries }}
                        </div>
                      </BCol>
                      <BCol md="4">
                        <span class="small text-secondary fw-semibold text-uppercase d-block">Backoff Strategy:</span>
                        <div class="font-monospace small">
                          <strong>Code:</strong> {{ fullTask.task.backOffCode || 'N/A' }} <br/>
                          <strong>Params:</strong> {{ JSON.stringify(fullTask.task.backOffParams) }}
                        </div>
                      </BCol>
                      <BCol md="5">
                        <span
                            class="small text-secondary fw-semibold text-uppercase d-block">Task Schedule Windows:</span>
                        <div class="font-monospace text-muted" style="font-size: 0.75rem;">
                          <div><strong>Lock Until:</strong>
                            {{ fullTask.task.lockUntil ? fullTask.task.lockUntil.toLocaleString() : 'N/A' }}
                          </div>
                          <div><strong>Created At:</strong>
                            {{ fullTask.task.createdAt ? fullTask.task.createdAt.toLocaleString() : 'N/A' }}
                          </div>
                          <div><strong>Last Run:</strong>
                            {{ fullTask.task.lastRun ? fullTask.task.lastRun.toLocaleString() : 'Never' }}
                          </div>
                          <div><strong>Next Run:</strong>
                            {{ fullTask.task.nextRun ? fullTask.task.nextRun.toLocaleString() : 'No schedule' }}
                          </div>
                          <div><strong>Task Deadline:</strong>
                            {{ fullTask.task.deadline ? fullTask.task.deadline.toLocaleString() : 'None' }}
                          </div>
                        </div>
                      </BCol>
                    </BRow>

                    <!-- РАЗДЕЛ В: ЛОГИ ОШИБОК И ОТВЕТОВ БЭКЕНДА (Итерируем TaskExecutionResult) -->
                    <div class="mt-2 px-2">
                      <div class="fw-bold mb-2 small text-dark text-uppercase tracking-wider">Execution Attempts Audit
                        Log (Task Run History)
                      </div>
                      <div class="timeline-container ps-2" v-if="fullTask.results.length > 0">

                        <div v-for="(res, resIdx) in fullTask.results" :key="res.id"
                             class="border-start border-2 ps-3 pb-3 position-relative">
                          <div class="d-flex justify-content-between align-items-center mb-1 flex-wrap gap-1">
                            <div>
                        <span class="fw-bold me-2" :class="res.status === 'Success' ? 'text-success' : 'text-danger'">
                          Attempt #{{ resIdx + 1 }} [{{ res.status }}]
                        </span>
                              <span class="badge bg-danger py-0 small" v-if="res.isCritical">CRITICAL</span>
                              <small class="text-muted font-monospace ms-2">Result ID: {{ res.id }}</small>
                            </div>
                            <small class="text-muted font-monospace">
                              {{ res.runAt ? res.runAt.toLocaleString() : 'N/A' }}
                              <span class="text-dark fw-bold ms-1">({{ res.executionTimeMs.toFixed(1) }} ms)</span>
                            </small>
                          </div>

                          <!-- Вывод ответа воркера -->
                          <div class="mt-1">
                            <code
                                class="d-block text-dark bg-white border rounded p-2 font-monospace response-output shadow-inner">
                              {{
                                res.result || 'Empty result trace buffer string returned from back-end microservice.'
                              }}
                            </code>
                          </div>
                        </div>

                      </div>
                      <div class="text-muted fst-italic small ps-2 my-2" v-else>
                        No runtime executions have logged results for this task block yet.
                      </div>
                    </div>
                  </div>

                </div>
              </td>
            </tr>

          </template>
          </tbody>
        </table>

      </div>

      <div class="d-flex justify-content-between align-items-center mt-3 small">
        <span class="text-muted">
          Showing page {{ resultData.pagination.currentPage }} of {{ resultData.pagination.totalPages }}
          ({{ resultData.pagination.total }} total entries)
        </span>
        <div class="d-flex gap-1">
          <BButton
              variant="outline-secondary"
              size="sm"
              :disabled="resultData.pagination.currentPage <= 1 || isLoading"
              @click="changePage(resultData.pagination.currentPage - 1)"
          >
            Prev
          </BButton>
          <BButton
              variant="outline-secondary"
              size="sm"
              :disabled="resultData.pagination.currentPage >= resultData.pagination.totalPages || isLoading"
              @click="changePage(resultData.pagination.currentPage + 1)"
          >
            Next
          </BButton>
        </div>
      </div>

    </div>
  </BCard>
</template>

<script setup>
import {onMounted, reactive, ref} from 'vue';
import {messageService} from '@/api/message/message.service.js';
import Multiselect from '@vueform/multiselect';
import '@vueform/multiselect/themes/default.css';

const resultData = ref(null);
const isLoading = ref(false);
const error = ref(null);
const expandedMessageId = ref(null);

const filterForm = reactive({
  idsRaw: '',
  code: [],
  transport: [],
  sender_ids: [],
  recipient: [],
  status: []
});

function parseRawStringToArray(rawStr) {
  if (!rawStr || !rawStr.trim()) return [];
  return rawStr.split(',').map(item => item.trim()).filter(item => item !== '');
}

function buildFilterRequestPayload() {
  return {
    ids: parseRawStringToArray(filterForm.idsRaw),
    code: filterForm.code,
    transport: filterForm.transport,
    sender_ids: filterForm.sender_ids,
    recipient: filterForm.recipient,
    status: filterForm.status
  };
}

async function searchRecipients(query) {
  if (!query || query.length < 2) return [];
  try {
    return await messageService.getRecipientSuggestions(query);
  } catch (err) {
    console.error('Failed to stream recipient suggestions from remote cluster:', err);
    return [];
  }
}

async function loadHistory(page = 1) {
  isLoading.value = true;
  error.value = null;
  try {
    const filtersPayload = buildFilterRequestPayload();
    resultData.value = await messageService.getMessages(page, 10, filtersPayload);
  } catch (err) {
    error.value = err.message || 'Failed to download message data profiles.';
  } finally {
    isLoading.value = false;
  }
}

const filterDictionaries = ref(null);
const isDictionariesLoading = ref(false);

async function fetchFiltersDictionaries() {
  isDictionariesLoading.value = true;
  try {
    // Сервис вернет инстанс MessageDictionariesDto
    filterDictionaries.value = await messageService.getMessageDictionaries();
  } catch (err) {
    console.error('Failed to parse dictionaries DTO layers:', err.message);
  } finally {
    isDictionariesLoading.value = false;
  }
}

function applyFilters() {
  expandedMessageId.value = null;
  loadHistory(1); // При поиске всегда сбрасываем на 1-ю страницу
}

function resetFilters() {
  filterForm.idsRaw = '';
  filterForm.codeRaw = [];
  filterForm.transportRaw = [];
  filterForm.senderIdsRaw = [];
  filterForm.recipientRaw = '';
  filterForm.status = [];

  expandedMessageId.value = null;
  loadHistory(1);
}

function toggleExpand(id) {
  expandedMessageId.value = expandedMessageId.value === id ? null : id;
}

function changePage(newPage) {
  expandedMessageId.value = null;
  loadHistory(newPage);
}

function getStatusBadgeClass(status) {
  if (status === 'succeeded') return 'bg-success';
  if (status === 'pending') return 'bg-warning text-dark';
  return 'bg-danger';
}

onMounted(() => {
  loadHistory(1);
  fetchFiltersDictionaries();
});
</script>

<style scoped>
.table-active {
  box-shadow: inset 4px 0 0 #0d6efd;
}

code {
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
