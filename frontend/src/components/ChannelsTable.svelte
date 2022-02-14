<script>
    import ChannelColumn from '../components/ChannelColumn.svelte';
    import HorizontalBarsCard from './cards/HorizontalBarsCard.svelte';

    export let channels;

    let open = false;

    let openTab = () => {
        open = true
    }

    let closeTab = () => {
        open = false
    }

    document.onkeydown = function(evt) {
        evt = evt || window.event;
        var isEscape = false;
        if ("key" in evt) {
            isEscape = (evt.key === "Escape" || evt.key === "Esc");
        } else {
            isEscape = (evt.keyCode === 27);
        }
        if ((isEscape) && (open = true)) {
            open = false
        }
    };

    function getGroupName(fw) {
        let name = fw.groupName || fw.groupId.substring(0,20)
        if (fw.channels.length > 1) {
            return "("+ fw.channels.length +") " + name
        }
        return name
    }


</script>

<div class="tab-wrapper">
    <div class:open class="filter-tab">
        <div class="filter-tab-header">
            <div class="heading">
                Filter
            </div>
            <div class="close" on:click={closeTab}>x</div>
        </div>
    </div>
    <div class="channels-table-wrapper">
        <div class="table-controls">
            <div class="filter-button table-settings" on:click={openTab}>Filter</div>
            <div class="filter-button table-settings" on:click={openTab}>Sort</div>
            <div class="filter-button time-filter" on:click={openTab} >January 5th&emsp;-&emsp;February 5th</div>
        </div>

        <div class="channels-table">
            <ChannelColumn alias={"Forwarded Amount"}>
                {#each channels.aggregatedForwards as fw}
                    <HorizontalBarsCard
                        props={{
                            type: fw.groupId,
                            heading: getGroupName(fw),
                            oValue: fw.amountOut,
                            iValue: fw.amountIn,
                            totalRow: true
                        }}
                    />
                {/each}
            </ChannelColumn>

            <ChannelColumn alias={"Fees earned"}>
                {#each channels.aggregatedForwards as fw}
                    <HorizontalBarsCard
                        props={{
                        type: fw.groupId,
                        heading: getGroupName(fw),
                        oValue: fw.feeOut,
                        iValue: fw.feeIn,
                        totalRow: true
                        }}
                    />
                {/each}
            </ChannelColumn>

            <ChannelColumn alias={"Forwarded Count"}>
                {#each channels.aggregatedForwards as fw}
                    <HorizontalBarsCard
                            props={{
                            type: fw.groupId,
                            heading: getGroupName(fw),
                            oValue: fw.countOut,
                            iValue: fw.countIn,
                            totalRow: true
                        }}
                    />
                {/each}
            </ChannelColumn>





        </div>
    </div>
</div>

<style lang="scss">

    .tab-wrapper {
      display: flex;
      flex-direction: row;
      flex-basis: 100%;
      align-items: stretch;
      height: 100vh;
      column-gap: 40px;
      overflow: hidden;
    }
    .filter-tab {
      flex-direction: row;
      flex-flow: column nowrap;
      background-color: #66786A;
      width: 0px;
      transition: width 200ms; // The close speed (reverse logic)
      transition-timing-function: ease;
      overflow-x: hidden;
    }
    .filter-tab.open {
      width: 375px;
      transition: width 250ms; // The open speed (reverse logic)
      transition-timing-function: ease;
    }
    .channels-table-wrapper {
        display: flex;
        flex-direction: column;
        margin-top: 40px;
        overflow-x: auto;
      flex-grow: 1;
    }
    .channels-table {
        margin-right: 20px;
        display: flex;
        flex-flow: column nowrap;
        flex-direction: row;
        align-items: stretch;
        column-gap: 30px;
        overflow: auto;
        -ms-overflow-style: none; /* Internet Explorer 10+ */
        scrollbar-width: none; /* Firefox */
    }
    .table-controls {
      margin-bottom: 30px;
        .filter-button {
          height: 33px;
          font-size: 16px;
          background-color: #66786A;
          color: white;
          display: inline-block;
          line-height: 33px;
          padding: 0px 15px;
          border-radius: 3px;
          margin-right: 17px;
        }
        .time-filter {
          background-color: #C7D1C9;
          color: #3A463C;
        }
    }
    /* .channels-table-wrapper {
        position: relative;
    }
    .channels-table-wrapper::before {
        content: " ";
        width: 10px;
        height: 100%;
        position: absolute;
        top: 0;
        left: 0;
        background: linear-gradient(90deg, #F3F4F5 0%, rgba(243, 244, 245, 0) 100%);
    } */
    .channels-table::-webkit-scrollbar {
        width: 0; /* Remove scrollbar space */
        background: transparent; /* Optional: just make scrollbar invisible */
    }
</style>
