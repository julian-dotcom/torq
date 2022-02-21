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
    <div class="bar-percent">
        <div class="outbound">

        </div>
        <div class="inbound">

        </div>
    </div>
    <div class="bar">
      <div class="bar-value" style="width: {oValuePercent}%" />
      <div class="bar-percent">
        <span><FormatNumber value={oValuePercent} decimals={0}/>%</span>
        <span><FormatNumber value={iValuePercent} decimals={0}/>%</span>
      </div>
    </div>
    <div class="bar-values">
        <div class="outbound">
            {#if values}
              <FormatNumber value={oValue} decimals={0} notation="standard" />
              {postfix}
            {/if}

        </div>
        <div class="inbound">
            {#if values}
                <FormatNumber value={iValue} decimals={0} notation="standard" />
                {postfix}
            {/if}
        </div>
    </div>
</div>

<style>
    .bar-values {
      display: flex;
      justify-content: space-between;
      padding: 5px 3px 1px;
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
      padding: 12px 15px 5.5px;
      font-size: 16px;
      line-height: 200%;
    }
    .bar {
      position: relative;
      background-color: #B3BCB5;
      width: 100%;
      height: 2em;
      margin: 4px 0;
      font-size: 12px;
      line-height: 2em;
    }
    .bar-value {
      position: absolute;
      top:0;
      left: 0;
      background-color: #94A097;
      height: 2em;
      width: 0%;
      max-width: calc(100% - 4px);
      min-width: 4px;
    }

</style>