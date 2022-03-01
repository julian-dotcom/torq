<script lang="ts">
  import FormatNumber from "../../helpers/FormatNumber.svelte";

  export let channels

</script>

<div class="columna" style="grid-template-columns: auto auto">
  <div class="column-header" >
    Turnover
  </div>
  <div class="column-header" >
    Capacity
  </div>
  {#each channels.aggregatedForwards as fw}
    <div class="cell-begin">
      <FormatNumber value={Number(fw.turnover)} decimals={2} notation="standard"/>
    </div>
    <div class="cell-end">
      <FormatNumber value={Number(fw.capacity)} decimals={0} notation="standard"/>
    </div>
  {/each}
</div>

<style lang="scss">
  .columna {
    display: grid;
    grid-auto-flow: row;
    grid-auto-columns: auto;
    grid-row-gap: 10px;
    align-items: start;
  }
  .column-header {
    height: 64px;
    line-height: 200%;
    margin-bottom: 11px;
    justify-items: center;
    display: grid;
    position: sticky;
    top: 0;
    z-index: 2;
    background-color: white;
    .bottom {
      color: rgba(0,0,0,0.2);
    }
  }
  .cell-begin {
    height: 68.4px;
    background-color: white;
    border-radius: 3px 0 0 3px;
    padding: 15px 15px 15px 35px;
    display: grid;
    text-align: right;
    align-items: center;
    /*box-shadow: 10px 0px 10px 0px rgba(0, 0, 0, 0.05);*/
  }
  .cell-end {
    height: 68.4px;
    background-color: white;
    border-radius: 0 3px 3px 0;
    border-left: solid 1px;
    border-left-color: #F3F4F5;
    padding: 15px 15px 15px 25px;
    display: grid;
    align-items: center;
    text-align: right;
    /*box-shadow: 10px 0px 10px 0px rgba(0, 0, 0, 0.05);*/
  }


</style>