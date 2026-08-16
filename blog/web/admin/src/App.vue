<template>
  <n-config-provider :theme-overrides="themeOverrides">
    <n-message-provider>
      <router-view />
    </n-message-provider>
  </n-config-provider>
</template>

<script setup>
const themeOverrides = {
  common: { primaryColor: '#b8944c', primaryColorHover: '#d4b060', borderRadius: '4px' },
}
</script>
