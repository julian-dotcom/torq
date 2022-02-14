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
        <div class="menu-item">
            <div class="icon">
                <img src="./images/node-icon.svg" alt="Node page icon" />
            </div>
            <div class="text">
                <a href="/">Node</a>
            </div>
        </div>
        <div class="menu-item">
            <div class="icon">
                <img src="./images/channels-icon.svg" alt="Channels page icon" />
            </div>
            <div class="text">
                <a href="/channels">Channels</a>
            </div>
        </div>
    </div>
</nav>

<style lang="scss">

    .navigation {
        position: fixed;
        left: 0;
        top: 0;
        color: white;
        height: 100vh;
        width: 200px;
        width: auto;
        display: inline;
        margin-right: 20px;
        padding: 34px 30px;
    }
    .logo {
        display: flex;
    }
    .logo .logo-text {
        margin-top: 2px;
    }
    .logo .logo-symbol {
        margin-right: 30px;
    }
    .menu {
        display: grid;
        margin-top: 80px;
    }
    .menu-item {
        .icon {
            display: inline;
            vertical-align: middle;
        }
        .text {
            display: inline;
            vertical-align: middle;
            font-size: 20px;
            a {
                color: white;
                text-decoration: none;
                margin-left: 30px;
            }
        }
        margin-bottom: 20px;
    }
</style>
