<script>
  export let expanded = false

  // import { fly } from 'svelte/transition';
  // import { cubicInOut } from 'svelte/easing';

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

<div class:expanded class="content">
  <slot />
</div>


<style lang="scss">
  .content {
    min-height: 100vh;
    background-color: #f3f4f5;
    margin-left: 110px;
    transition: margin-left 200ms;
    transition-delay: 200ms;
  }
  .content.expanded {
    margin-left: 280px;
    transition: margin-left 250ms;
    transition-delay: 200ms;
  }
</style>
