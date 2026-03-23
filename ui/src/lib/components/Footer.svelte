<script lang="ts">
  import { GITHUB_URL } from '$lib/config';

  let {
    temporalNamespace,
    temporalAddress,
  }: { temporalAddress: string; temporalNamespace: string } = $props();
</script>

<footer class="footer">
  <div class="content has-text-centered">
    <a
      href={GITHUB_URL}
      target="_blank"
      rel="noopener noreferrer"
      class="github-link has-text-grey"
      aria-label="GitHub repository"
    >
      <span class="icon is-large">
        <i class="mdi mdi-36px mdi-github"></i>
      </span>
    </a>

    <p>
      <strong>Temporal address</strong>:
      <code>{temporalAddress}</code>
    </p>

    <p>
      <strong>Temporal namespace</strong>:
      <code>{temporalNamespace}</code>
    </p>
  </div>
</footer>
