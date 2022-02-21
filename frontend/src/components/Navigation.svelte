<script>
  export let expanded = false

</script>


<nav class="navigation" class:expanded>

    <div class="logo">
        <div class="logo-symbol">
            <img src="./images/logo-symbol.svg" alt="LN.capital logo symbol" />
        </div>
        <div class="logo-text">
            <img src="./images/logo-text.svg" alt="LN.capital logo text" />
        </div>
    </div>

    <div class="menu">
        <a href="/" class="menu-item">
            <div class="icon">
                <img src="./images/node-icon.svg" alt="Node page icon" />
            </div>
            <div class="text">
                Node
            </div>
        </a>
        <a href="/channels" class="menu-item">
            <div class="icon">
                <img src="./images/channels-icon.svg" alt="Channels page icon" />
            </div>
            <div class="text">
                Channels
            </div>
        </a>
    </div>

</nav>

<style lang="scss">

    .navigation {
        position: fixed;
        top: 0;
        bottom: 0;
      z-index: 3;
        width: 110px;
        background-color: #3A463C;
        overflow: hidden;
        transition: width 200ms;
        transition-delay: 200ms;
        &.expanded {
          width: 280px;
          transition: width 250ms;
          transition-delay: 200ms;
        }
    }

    .logo {
        display: grid;
        grid-template-columns: 50px auto;
        grid-column-gap: 40px;
        margin-top: 40px;
        margin-left: 30px;
        margin-bottom: 40px;
        align-items: center;
      .icon {

      }
    }

    .menu {
      display: grid;
      grid-template-rows: repeat(1, auto);
      grid-row-gap: 20px;
      margin-left: 30px;
    }

    .menu-item {
        display: grid;
        grid-template-columns: 50px auto;
        grid-column-gap: 40px;

        align-items: center;
        text-decoration: none;
        .icon {
        }
        .text {
            font-size: 20px;
            color: white;
            text-decoration: none;
        }
    }

</style>
