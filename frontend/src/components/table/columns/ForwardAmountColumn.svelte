<script lang="ts">
  import Gauge from '../shared/Gauge.svelte';
  import FormatNumber from "../../helpers/FormatNumber.svelte";

  export let channels

</script>

<div class="column" style="grid-template-columns: auto auto">
  <div class="column-header">
    <div class="top">Amount forwarded</div>
    <div class="bottom">
      <div class="left">Outbound</div>
      <div class="right">Inbound</div>
    </div>
  </div>
  <div class="column-header total">
    <div class="top">&nbsp;</div>
    <div class="bottom">Total</div>
  </div>
  {#each channels.aggregatedForwards as fw}
    <div class="row-wrapper">
      <div class="cell-begin">
        <Gauge oValue={Number(fw.amountOut)} iValue={Number(fw.amountIn)}/>
      </div>
      <div class="cell-end">
        <FormatNumber value={Number(fw.amountOut) + Number(fw.amountIn)} decimals={0} notation="standard"/>
      </div>
    </div>
  {/each}
</div>

<style lang="scss">
  .column {
    display: grid;
    grid-auto-flow: row;
    grid-auto-columns: min-content;
    grid-row-gap: 10px;
  }
  .column-header {
    line-height: 200%;
    margin-bottom: 10px;
    display: grid;
    position: sticky;
    top: 105px;
    z-index: 2;
    background-color: #f3f4f5;
    border-bottom: 1px solid  #B3BCB5;
    .bottom {
      display: grid;
      grid-auto-flow: column;
      justify-content: space-between;
      color: rgba(0,0,0,0.2);
    }
    &.total {
      justify-items: center;
    }
  }
  .cell-begin {
    background-color: white;
    border-radius: 3px 0 0 3px;
    padding: 15px 15px;
    min-width: 300px;
    /*box-shadow: 10px 0px 10px 0px rgba(0, 0, 0, 0.05);*/
  }
  .cell-end {
    background-color: white;
    border-radius: 0 3px 3px 0;
    border-left: solid 1px;
    border-left-color: #F3F4F5;
    padding: 15px 15px 15px 25px;
    display: grid;
    align-items: center;
    text-align: right;
    min-width: 100px;
    /*box-shadow: 10px 0px 10px 0px rgba(0, 0, 0, 0.05);*/
  }
  .row-wrapper {
    display: contents;
  }

</style>