<script lang="ts">
  import { goto } from '$app/navigation';

  function newSession(): void {
    goto(`/documents/${crypto.randomUUID()}`);
  }
</script>

<nav class="navbar has-shadow is-spaced">
  <div class="navbar-brand">
    <a class="navbar-item" href="/"> Document Processing Demo </a>
  </div>

  <div class="navbar-end">
    <div class="navbar-item">
      <div class="field is-grouped">
        <p class="control">
          <button type="button" class="button is-primary" onclick={newSession}>
            New Session
          </button>
        </p>
      </div>
    </div>
  </div>
</nav>
