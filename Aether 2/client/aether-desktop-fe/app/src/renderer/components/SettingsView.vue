<template>
  <div id="vertical-flex-container">
    <AetherHeader></AetherHeader>
    <div id="horizontal-flex-container">
      <AetherSidebar></AetherSidebar>
      <div id="main-content">
        This is settings
      </div>
    </div>
  </div>
</template>

<script>
  import AetherHeader from './Shared/AetherHeader'
  import AetherSidebar from './Shared/AetherSidebar'
  export default {
    components: {
      AetherHeader,
      AetherSidebar
    },
    name: 'settings-view'
  }
</script>

<style scoped lang="scss">
  #main-content
  {
    flex:1;
    background-color:#aaa;
  }
</style>
