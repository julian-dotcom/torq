
<script lang="ts">
    import ChannelsTable from '../components/ChannelsTable.svelte';
    import { grpc } from '@improbable-eng/grpc-web';
    import { onMount, beforeUpdate} from 'svelte';
    import {torqrpcClientImpl, GrpcWebImpl} from '../torqrpc/torq'
    let channels;

    const rpc = new GrpcWebImpl('https://localhost:50051', {
        debug: false,
        metadata: new grpc.Metadata({}),
    });

    const client = new torqrpcClientImpl(rpc);
    const day = (60*60*24*1000)
    let getForwards = async function() {
        return client.GetAggrigatedForwards({
            peerIds: {pubKeys: []},
            fromTs: 0, //(Date.now() - (7*day))/1000
            toTs: Date.now()/1000,
        })
    }
	
</script>

<svelte:head>
    <title>Torq | Channels</title>
</svelte:head>

<div class="forwards">
    {#await getForwards()}
        <div>Loading forwarding activity</div>
    {:then channels}
        <ChannelsTable channels={channels} />
    {/await}
    
</div>

<style>


</style>