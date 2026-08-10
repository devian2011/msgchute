<template>
  <div class="dashboard-board w-100">

    <section class="mb-5">
      <div class="d-flex align-items-center mb-3">
        <h2 class="fs-4 fw-bold text-dark m-0">Admin Workspace</h2>
        <span class="badge bg-danger-subtle text-danger ms-2 px-2 py-1 rounded-pill small fw-semibold">Restricted</span>
      </div>

      <BRow class="g-3">
        <BCol md="6" lg="4">
          <BLink to="/admin/messages" class="dashboard-card-link">
            <BCard class="h-100 shadow-sm border-0 dashboard-card text-start">
              <div class="fw-bold text-dark mb-1">Messages</div>
              <div class="text-muted small">Manage sent history, logs, and system dispatches.</div>
            </BCard>
          </BLink>
        </BCol>

        <BCol md="6" lg="4">
          <BLink to="/admin/templates" class="dashboard-card-link">
            <BCard class="h-100 shadow-sm border-0 dashboard-card text-start">
              <div class="fw-bold text-dark mb-1">Templates</div>
              <div class="text-muted small">Edit global layouts, presets, and dynamic variables.</div>
            </BCard>
          </BLink>
        </BCol>

        <BCol md="6" lg="4">
          <BLink to="/admin/worker/statuses" class="dashboard-card-link">
            <BCard class="h-100 shadow-sm border-0 dashboard-card text-start">
              <div class="fw-bold text-dark mb-1">Worker Statuses</div>
              <div class="text-muted small">Monitor microservice performance and queues.</div>
            </BCard>
          </BLink>
        </BCol>
      </BRow>
    </section>

    <section>
      <div class="d-flex align-items-center mb-3">
        <h2 class="fs-4 fw-bold text-dark m-0">Public Actions</h2>
        <span class="badge bg-success-subtle text-success ms-2 px-2 py-1 rounded-pill small fw-semibold">Open</span>
      </div>

      <BRow class="g-3">
        <BCol md="6" lg="4">
          <BLink to="/send" class="dashboard-card-link">
            <BCard class="h-100 shadow-sm border-0 dashboard-card text-start action-card text-start">
              <div class="fw-bold text-primary mb-1">Send Message</div>
              <div class="text-muted small">Dispatch a live notification payload immediately.</div>
            </BCard>
          </BLink>
        </BCol>
      </BRow>
    </section>

  </div>
</template>

<script setup>

</script>

<style scoped>
.dashboard-card-link {
  text-decoration: none !important;
  display: block;
}

.dashboard-card {
  transition: transform 0.2s ease, box-shadow 0.2s ease, border-color 0.2s ease;
  border-left: 4px solid #495057 !important;
}

.dashboard-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 0.5rem 1rem rgba(0, 0, 0, 0.08) !important;
  background-color: #f8f9fa;
}

.action-card {
  border-left: 4px solid #0d6efd !important;
}

.action-card:hover {
  background-color: #f0f5ff;
}

.bg-danger-subtle {
  background-color: #f8d7da;
}

.bg-success-subtle {
  background-color: #d1e7dd;
}
</style>
