<script lang="ts">
    import FormatNumber from '../../helpers/FormatNumber.svelte';
    export let postfix = '';
    export let iValue: number;
    export let oValue: number;
    export let values = true;
    export let percent = false;
    $: iValue = Number(iValue);
    $: oValue = Number(oValue);
    $: iValuePercent = (Number(iValue) / (Number(oValue) + Number(iValue))) * 100;
    $: oValuePercent = (Number(oValue) /  (Number(oValue) + Number(iValue))) * 100;
</script>

<div class="bar-row">
    <div class="bar-values">

        <div class="inbound">
            {#if values}
                <FormatNumber value={iValue} decimals={0} notation="standard" />
                {postfix}
            {/if}
        </div>
        <div class="outbound">
            {#if values}
              <FormatNumber value={oValue} decimals={0} notation="standard" />
              {postfix}
            {/if}

        </div>

    </div>
    <div class="bar">
      <div class="bar-value" style="width: {iValuePercent}%" />
      <div class="bar-percent">
<!--        <span><FormatNumber value={oValuePercent} decimals={0}/>%</span>-->
<!--        <span><FormatNumber value={iValuePercent} decimals={0}/>%</span>-->
      </div>
    </div>
</div>

<style>
    .bar-values {
      display: flex;
      justify-content: space-between;
      column-gap: 100px;

      white-space: nowrap;
      margin-bottom: 5px;
    }
    .bar-percent {
      position: absolute;
      top:0;
      left: 0;
      width: 100%;
      color: white;
      display: flex;
      justify-content: space-between;
      padding: 0 5px;
    }
    .bar-row {
      min-width: 220px;
    }
    .bar {
      position: relative;
      background-color: #66786A;
      width: 100%;
      height: 10px;
      margin-bottom: 5px;
    }
    .bar-value {
      position: absolute;
      top:0;
      left: 0;
      background-color: #B3BCB5;
      height: 10px;
      width: 0%;
      max-width: calc(100% - 4px);
      min-width: 4px;
    }

</style>


