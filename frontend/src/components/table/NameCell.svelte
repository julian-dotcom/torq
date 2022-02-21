<script lang="ts">
  import Cell from './Cell.svelte'
  export let name;
  export let status;

</script>

<Cell>
  <div class="name-wrapper">
    <div class="name">{name}</div>
    <div class="status">{status}</div>
  </div>
</Cell>

<style lang="scss">
  .name-wrapper {
    .status {
      color: rgba(0,0,0,.4);
    }
  }
</style>