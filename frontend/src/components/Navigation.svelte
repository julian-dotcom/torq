<nav class="navigation">

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

    .logo {
        display: grid;
        grid-template-columns: 50px auto;
        grid-column-gap: 40px;
        margin-top: 40px;
        margin-left: 40px;
        margin-bottom: 40px;
        align-items: center;
      .icon {

      }
    }

    .menu {
      display: grid;
      grid-template-rows: repeat(1, auto);
      grid-row-gap: 20px;
      margin-left: 40px;
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
