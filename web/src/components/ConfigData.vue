<template>
  <BRow class="align-items-center bg-white p-3 rounded shadow-sm">
    <BCol md="4" class="fw-medium text-center">Host: <span class="text-muted">{{ host }}</span></BCol>
    <BCol md="4" class="fw-medium text-center">Header: <span class="text-muted">{{ apiHeader }}</span></BCol>
    <BCol md="4" class="fw-medium text-center">
      Key:
      <span
          class="text-muted cursor-pointer user-select-none transition-blur"
          :class="{ 'blurred-text': !isKeyVisible }"
          @click="toggleKeyVisibility"
      >
        {{ apiKey || '—' }}
      </span>
    </BCol>
  </BRow>
</template>

<script setup>
import { ref } from 'vue'
import { useConfigStore } from '@/stores/config'
import { storeToRefs } from 'pinia'

const store = useConfigStore()
const { host, apiKey, apiHeader } = storeToRefs(store)

const isKeyVisible = ref(false)

function toggleKeyVisibility() {
  isKeyVisible.value = !isKeyVisible.value
}

</script>

<style scoped>
.blurred-text {
  filter: blur(5px);
}

.cursor-pointer {
  cursor: pointer;
}

.transition-blur {
  transition: filter 0.2s ease-in-out;
}
</style>
