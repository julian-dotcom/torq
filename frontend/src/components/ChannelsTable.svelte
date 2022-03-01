<script>
  import { onMount } from 'svelte';
  import ChannelColumn from '../components/ChannelColumn.svelte';
  import ChannelFilter from '../components/ChannelFilter.svelte';
  import HorizontalBarsCard from './cards/HorizontalBarsCard.svelte';
  import { grpc } from '@improbable-eng/grpc-web';
  import {torqrpcClientImpl, GrpcWebImpl} from '../torqrpc/torq'
  import GaugeBarCard from "./cards/GaugeBarCard.svelte";


  onMount(() => {
    document.onkeydown = function(evt) {
        evt = evt || window.event;
        let isBody = document.activeElement == document.body
        let isEscape = false;
        let isF = false;
        if ("key" in evt) {
            isEscape = (evt.key === "Escape" || evt.key === "Esc");
            isF = (evt.key === "f" || evt.key === "f");
        }
        if (isBody && (isEscape) && (open = true)) {
            open = false
        }
        if (isBody && isF) {
            open = !open
        }
    };
  });


    let open = false;

    let openTab = () => {
        open = true
    }

    let closeTab = () => {
        open = false
    }

    function getGroupName(fw) {
        let name = fw.groupName || fw.groupId.substring(0,20)
        if (fw.channels.length > 1) {
            return "("+ fw.channels.length +") " + name
        }
        return name
    }


    let today = new Date()
    let seven_days_ago =  new Date(new Date().setDate(today.getDate()-7))

    let lastFromDate = ''
    let lastToDate = ''
    let fromDate = new Date(seven_days_ago.setHours(1, 0, 0)).toISOString().substring(0,19)
    let toDate = new Date(today.setHours(24, 59, 0)).toISOString().substring(0,19)

    const rpc = new GrpcWebImpl('https://localhost:50051', {
        debug: false,
        metadata: new grpc.Metadata({}),
    });

    const client = new torqrpcClientImpl(rpc);

    let p
    let getForwards = function() {
        lastFromDate = fromDate
        lastToDate = toDate
        return client.GetAggrigatedForwards({
          peerIds: {pubKeys: []},
          fromTs: Date.parse(fromDate).valueOf() / 1000, //(Date.now() - (7*day))/1000
          toTs: Date.parse(toDate).valueOf() / 1000,
        })
    }

    p = getForwards()

    let updateForwards = function(fd, td) {
      if ((lastFromDate !== fd )|| (lastToDate !== td)) {
        p = getForwards()
      }
    }

    $: updateForwards(fromDate, toDate)

    // TODO: Clean up this mess
    let month = {
      0: "Jan",
      1: "Feb",
      2: "Mar",
      3: "Apr",
      4: "May",
      5: "Jun",
      6: "Jul",
      7: "Aug",
      8: "Sep",
      9: "Okt",
      10: "Nov",
      11: "Des",
    }
    let pf = {
      1: "st",
      2: "nd",
      3: "rd"
    }
    let postfix = function(day) {
      let res = "th"

      if (day <= 3 ) {
        return pf[day]
      } else if ((day >= 10) && (day <= 19)) {
        return res
      } else {
        day = ""+day
        let lastChar
        lastChar = day.substring(day.length,1)
        if (pf[lastChar]) {
          return pf[lastChar]
        }
        return "th"
      }
    }

    let formatDate = function(date) {
      date = new Date(date)
      let d = month[date.getMonth()] + " " + date.getDate() + postfix(date.getDate())

      return d
    }


</script>

<div class="tab-wrapper">

    <ChannelFilter {open} {closeTab} bind:fromDate={fromDate} bind:toDate={toDate}/>

    <div class="channels-table-wrapper" class:open >
        <div class="table-controls" class:open>
            <div class="filter-button table-settings" on:click={openTab}>Filter</div>
            <div class="filter-button table-settings" on:click={openTab}>Sort</div>
            <div class="filter-button time-filter" on:click={openTab} >
              {formatDate(fromDate)}&emsp;-&emsp;{formatDate(toDate)}
            </div>
        </div>

    {#await p }
        <div>Loading forwarding activity</div>
    {:then channels}
        <div class="channels-table">
          <div class="row column-header">
            <ChannelColumn alias={"Forwarded amount"}></ChannelColumn>
            <ChannelColumn alias={"Fee revenue"}></ChannelColumn>
            <ChannelColumn alias={"Forward count"}></ChannelColumn>
          </div>

          {#each channels.aggregatedForwards as fw}
            <div class="row">

              <!-- <div class="name">{getGroupName(fw)}</div>-->
              <GaugeBarCard
                props={{
                  type: fw.groupId,
                  heading: getGroupName(fw),
                  oValue: fw.amountOut,
                  iValue: fw.amountIn,
                  totalRow: false
                }}
              />

              <GaugeBarCard
                props={{
                  type: fw.groupId,
                  heading: "-",
                  oValue: fw.revenueOut,
                  iValue: fw.revenueIn,
                  totalRow: false
                }}
              />


            </div>
          {/each}
      </div>
    {/await}
      </div>
</div>

<style lang="scss">

    .tab-wrapper {
      position: relative;
    }
    .channels-table-wrapper {
      padding-left: 40px;
      margin-left: 0;
      transition: margin-left 200ms; // The close speed (reverse logic)
      transition-timing-function: ease;
      &.open {
        margin-left: 375px;
        transition: margin-left 250ms; // The open speed (reverse logic)
        transition-timing-function: ease;
      }
    }
    .channels-table {
      .column-header {
        position: sticky;
        top:0;
      }
        display: grid;
        grid-template-columns: max-content;
    }
    .row {
      display: grid;
      grid-auto-columns: 330px;
      column-gap: 40px;
      grid-auto-flow: column;
    }
    .card
    .name {
      position: sticky;
      margin-left: -40px;
      padding-left: 40px;
      left: 0;
    }
    .table-controls {
      padding-bottom: 30px;
      padding-top: 50px;
      background-color: #f3f4f5;
      z-index: 2;
      display: block;
      width: 100%;
      position: sticky;
      top: 0;

      .filter-button {
        height: 33px;
        font-size: 16px;
        background-color: #66786A;
        color: white;
        display: inline-block;
        line-height: 33px;
        padding: 0 15px;
        border-radius: 3px;
        margin-right: 17px;
      }
      .time-filter {
        background-color: #C7D1C9;
        color: #3A463C;
      }
    }
</style>
