<script lang="ts">

  export let open
  export let closeTab

  export let fromDate
  export let toDate

  let fdate = fromDate
  let tdate = toDate

  let handleFrom = (e) => {
    fromDate = fdate
  }

  let handleTo = () => {
    toDate = tdate

  }

  // let last_month_from = new Date(new Date(new Date().setMonth(today.getMonth() -1, 1)).setHours(1,0,0))
  //     .toISOString().substring(0,19)
  // let last_month_to = new Date(new Date(new Date().setMonth(today.getMonth(), 1)).setHours(0,59,0))
  //     .toISOString().substring(0,19)

</script>


<div class:open class="filter-tab-wrapper">
  <div class="filter-tab">
    <div class="filter-tab-header">
      <div class="heading">
        Table menu
      </div>
      <div class="close" on:click={closeTab}>
        &#215;
      </div>
    </div>

    <div class="filter-section time">
      <div class="heading">
        Time interval
      </div>
      <div class="fields">
        <div class="field-wrapper">
          <label for="from-time">From</label>
          <input type="datetime-local" id="from-time"  name="from-time"
                 bind:value={fdate}
                 on:blur={handleFrom}
          />
        </div>
        <div class="field-wrapper" >
          <label for="to-time">To</label>
          <input type="datetime-local" id="to-time"  name="to-time"
                 bind:value={tdate}
                 on:blur={handleTo}/>
        </div>
      </div>
    </div>
  </div>

</div>

<style lang="scss">
    .filter-tab-wrapper {
      position: fixed;
      top: 0;
      bottom: 0;
      background-color: #66786A;
      width: 0;
      transition: width 200ms; // The close speed (reverse logic)
      transition-timing-function: ease;
      overflow-x: hidden;
      color: white;
      z-index: 4;
    }
    .filter-tab-wrapper.open {
      width: 375px;
      transition: width 250ms; // The open speed (reverse logic)
      transition-timing-function: ease;
    }
    .filter-tab {
      width: 375px;
    }
    .filter-tab-header {
      display: grid;
      grid-template-columns: auto min-content;
      align-items: center;
      margin-top: 20px;
      font-size: 25px;
      .heading {
        margin-left: 30px;
        white-space: nowrap;
      }
      .close {
        color: white;
        padding: 20px 30px;
      }
    }
    .filter-section {
      margin: 30px;
      .heading {
        font-size: 20px;
      }
    }
    .fields {
      margin-top: 20px;
      display: grid;
      .field-wrapper {
        display: grid;
        grid-template-columns: 100px auto;
        line-height: 28px;
        label {
            color: rgba(255,255,255,0.5);
        }
        ::-webkit-calendar-picker-indicator {
          filter: invert(1);
          margin-left: 10px;
          margin-top: 2px;
        }
        input[type=datetime-local] {
          color: white;
          background-color: transparent;
          border: none;
          text-align: right;

        }
      }
    }


</style>
