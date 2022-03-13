
<script lang="ts">
    import ChannelsTable from '../components/ChannelsTable.svelte';
    import TablePage from '../components/table/Table.svelte';
    import { grpc } from '@improbable-eng/grpc-web';
    // import { onMount, beforeUpdate} from 'svelte';
    import {torqrpcClientImpl, GrpcWebImpl} from '../torqrpc/torq'


</script>

<svelte:head>
    <title>Torq | Channels</title>
</svelte:head>

<div class="forwards">
<!--  <ChannelsTable />-->
  <Table />
</div>

<style>


</style>
