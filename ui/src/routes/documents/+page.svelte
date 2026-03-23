<script lang="ts">
  import type { PageData } from './$types';

  let { data }: { data: PageData } = $props();

  function formatDate(iso: string): string {
    return new Date(iso).toLocaleString(undefined, {
      dateStyle: 'medium',
      timeStyle: 'short',
    });
  }

  function statusLabel(status: string): string {
    switch (status) {
      case 'RUNNING':
        return 'Running';
      case 'COMPLETED':
        return 'Completed';
      case 'FAILED':
        return 'Failed';
      case 'TIMED_OUT':
        return 'Timed out';
      case 'CANCELED':
        return 'Cancelled';
      case 'TERMINATED':
        return 'Terminated';
      default:
        return status;
    }
  }

  function statusClass(status: string): string {
    switch (status) {
      case 'RUNNING':
        return 'is-info';
      case 'COMPLETED':
        return 'is-success';
      case 'FAILED':
      case 'TIMED_OUT':
      case 'TERMINATED':
        return 'is-danger';
      default:
        return 'is-light';
    }
  }
</script>

{#if data.documents.length === 0}
  <p class="has-text-grey">
    No sessions yet. Use the button above to start one.
  </p>
{:else}
  <table class="table is-fullwidth is-hoverable">
    <thead>
      <tr>
        <th>Session ID</th>
        <th>Status</th>
        <th>Started</th>
        <th>Closed</th>
      </tr>
    </thead>
    <tbody>
      {#each data.documents as doc}
        <tr>
          <td>
            <a
              href="/documents/{doc.id}"
              class="id-link is-family-monospace is-size-7"
              title={doc.id}
            >
              {doc.id.slice(0, 8)}&hellip;
            </a>
          </td>
          <td>
            <span class="tag {statusClass(doc.status)}"
              >{statusLabel(doc.status)}</span
            >
          </td>
          <td class="time has-text-grey">{formatDate(doc.startTime)}</td>
          <td class="time has-text-grey"
            >{doc.closeTime ? formatDate(doc.closeTime) : '—'}</td
          >
        </tr>
      {/each}
    </tbody>
  </table>
{/if}

<style>
  /* color: inherit prevents Bulma's link colour overriding the table text */
  .id-link {
    color: inherit;
  }

  /* white-space: nowrap has no Bulma utility equivalent */
  .time {
    white-space: nowrap;
  }
</style>
