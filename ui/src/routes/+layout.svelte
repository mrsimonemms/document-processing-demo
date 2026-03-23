<script lang="ts">
  import 'bulma/css/bulma.css';
  import type { Snippet } from 'svelte';

  import type { LayoutData } from './$types';
  import Footer from '$lib/components/Footer.svelte';
  import Navbar from '$lib/components/Navbar.svelte';

  let { data, children }: { data: LayoutData; children: Snippet } = $props();
</script>

<Navbar />

<section class="section is-flex-grow-1">
  <div class="container">
    {@render children()}
  </div>
</section>

<Footer
  temporalAddress={data.temporalAddress}
  temporalNamespace={data.temporalNamespace}
/>
