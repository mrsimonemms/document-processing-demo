<script lang="ts">
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import type { Snippet } from 'svelte';

  let { children }: { children: Snippet } = $props();

  function newSession(): void {
    goto(`/documents/${crypto.randomUUID()}`);
  }
</script>

<header
  class="is-flex is-justify-content-space-between is-align-items-flex-start mb-5 pb-4"
>
  <div class="is-flex is-flex-direction-column">
    <h1 class="title is-4 mb-0">
      {$page.data.title ?? 'Document processing'}
    </h1>
    {#if $page.data.subtitle}
      <p class="subtitle is-6 is-family-monospace mb-0">
        {$page.data.subtitle}
      </p>
    {/if}
  </div>
  <button type="button" class="button" onclick={newSession}>New session</button>
</header>

{@render children()}

<style>
  /* Border only — flex layout expressed via Bulma helpers above */
  header {
    border-bottom: 1px solid #e0e0e0;
  }
</style>
