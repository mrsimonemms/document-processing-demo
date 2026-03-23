<script lang="ts">
  interface Item {
    label: string;
    href?: string;
  }

  let { items }: { items: Item[] } = $props();
</script>

<nav class="breadcrumb is-small has-bullet-separator" aria-label="breadcrumbs">
  <ul>
    {#each items as item, i}
      {#if i === items.length - 1}
        <li class="is-active">
          <a href={item.href} aria-current="page">{item.label}</a>
        </li>
      {:else}
        <li>
          <a href={item.href}>{item.label}</a>
        </li>
      {/if}
    {/each}
  </ul>
</nav>
