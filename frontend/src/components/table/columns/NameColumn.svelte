<script lang="ts">
  export let channels

  function getGroupName(fw) {
    let name = fw.groupName || fw.groupId.substring(0,20)
    if (fw.channels.length > 1) {
      return "("+ fw.channels.length +") " + name
    }
    return name
  }
</script>

<div class="column">
  <div class="column-header">
    <div class="top">Name</div>
    <div class="bottom">&nbsp;</div>
  </div>
  {#each channels.aggregatedForwards as fw}
    <div class="cell">
      <div class="name-wrapper">
        <div class="name">{getGroupName(fw)}</div>
        <div class="status">{"active"}</div>
      </div>
    </div>
  {/each}
</div>

<style lang="scss">
  .column {
    display: grid;
    align-items: start;
    grid-auto-flow: row;
    grid-auto-columns: min-content;
    grid-row-gap: 10px;
    position: relative;
  }
  .column-header {
    line-height: 200%;
    margin-bottom: 10px;
    position: sticky;
    top: 105px;
    z-index: 2;
    background-color: #f3f4f5;
    border-bottom: 1px solid  #B3BCB5;
    //box-shadow: 0px 0px 15px 0px rgba(0, 0, 0, 0.05);
    //clip-path: polygon(0% 0%, 100% 0%, 100% 120%, 0% 120%);
    .bottom {
      display: grid;
      grid-auto-flow: column;
      justify-content: space-between;
      color: rgba(0,0,0,0.2);
    }
  }
  .cell {
    background-color: white;
    border-radius: 3px;
    padding: 15px 15px;
  }
  .name-wrapper {
    .status {
      color: rgba(0,0,0,.4);
    }
  }


</style>