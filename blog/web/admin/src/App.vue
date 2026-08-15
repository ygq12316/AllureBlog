<template>
  <n-config-provider :theme-overrides="themeOverrides">
    <router-view />
  </n-config-provider>
</template>

<script setup>
const themeOverrides = {
  common: { primaryColor: '#b8944c', primaryColorHover: '#d4b060', borderRadius: '4px' },
}
</script>
