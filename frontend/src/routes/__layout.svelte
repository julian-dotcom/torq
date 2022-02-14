<script>
    import { page } from '$app/stores'; // svelte@next
    import Navigation from '../components/Navigation.svelte';
    import PageWrapper from '../components/PageWrapper.svelte';
</script>

<svelte:head>
    <link rel="preconnect" href="https://fonts.googleapis.com" />
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin />
    <link
        href="https://fonts.googleapis.com/css2?family=Inter:wght@400;600&display=swap"
        rel="stylesheet"
    />
    <link rel="stylesheet" href="./styles/normalize.css" />
    <link rel="stylesheet" href="./styles/typography.scss" />
    <link rel="stylesheet" href="./styles/global.scss" />
</svelte:head>

<div class="page">
    <Navigation />
    <div class="content-wrapper">
        <PageWrapper refresh={$page.path}>
            <slot />
        </PageWrapper>
    </div>
</div>

<style lang="scss">

    .page {
        background-color: #3a463c;
        width: 100%;
        overflow-x: hidden;
    }
</style>
