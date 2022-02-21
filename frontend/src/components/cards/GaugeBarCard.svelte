<script lang="ts">
    import CardWrapper, { CardWrapperProps } from './CardWrapper.svelte';
    import Gauge from './rows/Gauge.svelte';
    import Row from './rows/Row.svelte';

    type GaugeBarCardProps = CardWrapperProps & {
        iValue: number;
        oValue: number;
        totalRow?: boolean;
    };

    export let props: GaugeBarCardProps;
    
    $: props.iValue = Number(props.iValue)
    $: props.oValue = Number(props.oValue)
    $: totalValue = props.iValue + props.oValue

</script>

<CardWrapper props={{ type: props.type, heading: props.heading }}>
    <Gauge oValue={props.oValue} iValue={props.iValue} percent={true} />

    {#if props.totalRow}
        <Row label="Total" value={totalValue} notation="standard" />
    {/if}

</CardWrapper>
