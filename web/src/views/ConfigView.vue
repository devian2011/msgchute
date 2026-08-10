<template>
  <BCard
      title="Settings"
      class="shadow border-0 rounded-3 overflow-hidden text-start"
  >
    <BForm @submit.prevent="saveAndRedirect">

      <BFormGroup
          label="Host:"
          label-for="host-input"
          class="mb-3 fw-semibold text-muted small"
      >
        <BFormInput
            id="host-input"
            v-model="host"
            type="text"
            placeholder="https://example.com"
            class="mt-1"
            required
        />
      </BFormGroup>

      <BFormGroup
          label="Header name:"
          label-for="header-input"
          class="mb-3 fw-semibold text-muted small"
      >
        <BFormInput
            id="header-input"
            v-model="apiHeader"
            type="text"
            placeholder="Type header name"
            class="mt-1"
            required
        />
      </BFormGroup>

      <BFormGroup
          label="Key (API Key):"
          label-for="api-key-input"
          class="mb-4 fw-semibold text-muted small"
      >
        <BFormInput
            id="api-key-input"
            v-model="apiKey"
            type="password"
            placeholder="Type api key"
            class="mt-1"
            required
        />
      </BFormGroup>

      <div class="d-flex gap-2">
        <BButton
            type="submit"
            variant="primary"
            class="flex-grow-1 py-2 fw-medium shadow-sm"
        >
          Save Configuration
        </BButton>

        <BButton
            type="button"
            variant="outline-danger"
            class="py-2 px-3 fw-medium"
            @click="store.clearConfig"
        >
          Clear
        </BButton>
      </div>

    </BForm>
  </BCard>
</template>

<script setup>
import { useConfigStore } from '@/stores/config'
import { storeToRefs } from 'pinia'
import { useRouter } from 'vue-router'
import {BButton, BCard, BForm, BFormGroup, BFormInput} from "bootstrap-vue-next";

const store = useConfigStore()
const router = useRouter()

const { host, apiKey, apiHeader } = storeToRefs(store)

function saveAndRedirect() {
  if (host.value && apiKey.value && apiHeader.value) {
    router.push({ name: 'Home' })
  } else {
    alert('Please fill in all fields!')
  }
}
</script>
