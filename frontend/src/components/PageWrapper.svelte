<script>
    // import { fly } from 'svelte/transition';
    // import { cubicInOut } from 'svelte/easing';
    export let refresh = '';
    // TODO: Fix slide in/out animations. Page is duplicated
    // const slideOut = () => {
    // 	return {
    // 		css: (t) => {
    // 			return `
    //     transform: translateX(${(1 - t) * 100}vw);
    //     `;
    // 		},
    // 		easing: cubicInOut,
    // 		duration: 1950
    // 	};
    // };

    // const slideIn = () => {
    // 	return {
    // 		css: (t) => {
    // 			return `
    //     transform: translateX(${100 - t * 100}vw);
    //     `;
    // 		},
    // 		easing: cubicInOut,
    // 		delay: 1950,
    // 		duration: 2250
    // 	};
    // };
</script>

<!-- 
in:slideIn
out:slideOut
 -->

<!--
      in:fly="{{ y: 200, duration: 2000, delay: 2000 }}"
    out:fly="{{ y: 200, duration: 2000 }}"
-->
{#key refresh}
    <div class="content">
        <slot />
    </div>
{/key}

<style>
    .content {
        /* Need to find a way to keep the width while there is a third element there
          Consider not using flexbox here.
       */
        /* float: left; */
        /* width: calc(100% - 220px); */
        /* display: none; */
        margin: 0;
        min-height: 100vh;
        background-color: #f3f4f5;
        margin: 0;
        padding: 0;
        /* border-radius: 3px 0 0 3px; */
        margin-left: 280px;
        padding-right: 0px;
    }
    /* .content:first-child {
      display: block;
  } */
</style>
