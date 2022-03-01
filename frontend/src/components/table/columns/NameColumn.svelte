<script lang="ts">
  export let channels

  function getGroupName(fw) {
    let name = fw.groupName || fw.groupId.substring(0,20)
    if (fw.channels.length > 1) {
      return "("+ fw.channels.length +") " + name
    }
    return name
  }

  function getChannelStatus(chanInfo) {
    let open = 0
    let total = 0
    for (let chan of chanInfo) {
      if (!chan.closed) {
        open++
      }
      total = chanInfo.length
    }
    if (total == 1) {
      return open ? "Open" : "Closed"
    }
    return open+"/"+total+" Open"
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
        <div class="status">{getChannelStatus(fw.channels)}</div>
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
    top: 0px;
    z-index: 2;
    background-color: white;
    //border-bottom: 1px solid  #B3BCB5;
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
    white-space: nowrap;
    .status {
      color: rgba(0,0,0,.4);
    }
  }


</style>