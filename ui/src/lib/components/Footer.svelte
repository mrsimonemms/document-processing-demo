<script lang="ts">
  import { GITHUB_URL } from '$lib/config';

  let {
    temporalNamespace,
    temporalAddress,
  }: { temporalAddress: string; temporalNamespace: string } = $props();
</script>

<footer class="footer">
  <div class="columns is-centered">
    <div class="column is-4">
      <ul>
        <li>
          <a
            href={GITHUB_URL}
            target="_blank"
            rel="noopener noreferrer"
            aria-label="Source code"
          >
            Source code
          </a>
        </li>
        <li>
          <a
            href="https://temporal.io"
            target="_blank"
            rel="noopener noreferrer"
            aria-label="Temporal homepage"
          >
            Temporal
          </a>
        </li>
      </ul>
    </div>
    <div class="column is-4">
      <ul>
        <li>
          <strong>Temporal address</strong>:
          <code>{temporalAddress}</code>
        </li>
        <li>
          <strong>Temporal namespace</strong>:
          <code>{temporalNamespace}</code>
        </li>
      </ul>
    </div>
  </div>
</footer>
