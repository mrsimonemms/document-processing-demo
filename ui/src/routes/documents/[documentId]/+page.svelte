<script lang="ts">
  import { untrack } from 'svelte';

  import Breadcrumbs from '$lib/components/Breadcrumbs.svelte';
  import type { PageData } from './$types';

  let { data }: { data: PageData } = $props();

  let phase = $state(untrack(() => data.phase));
  let summary = $state(untrack(() => data.summary));
  let provider = $state(untrack(() => data.provider));
  let fallbackOccurred = $state(untrack(() => data.fallbackOccurred ?? false));
  let qa: { question: string; answer: string }[] = $state(
    untrack(() => [...data.qa]),
  );
  let error = $state('');
  let submitting = $state(false);

  // Re-seed local state whenever the route navigates to a different session.
  // Reading data.documentId outside untrack() makes it the sole trigger;
  // the remaining reads inside untrack() pick up current values without
  // creating additional dependencies.
  $effect(() => {
    // eslint-disable-next-line @typescript-eslint/no-unused-expressions
    data.documentId;
    untrack(() => {
      phase = data.phase;
      summary = data.summary;
      provider = data.provider;
      fallbackOccurred = data.fallbackOccurred ?? false;
      qa = [...data.qa];
      error = '';
      submitting = false;
    });
  });

  async function refreshState() {
    try {
      const res = await fetch(`/api/documents/${data.documentId}`);
      const state = await res.json();
      // Do not regress an already-ended session; the workflow may not have
      // completed yet when this is called immediately after ending.
      if (phase !== 'ended') {
        phase = state.phase;
      }
      summary = state.summary ?? summary;
      provider = state.provider ?? provider;
      fallbackOccurred = state.fallbackOccurred ?? false;
      qa = state.qa ?? qa;
    } catch {
      // Network error — keep current state.
    }
  }

  // Auto-poll while the workflow is processing.
  $effect(() => {
    if (phase !== 'processing') return;
    const id = setInterval(refreshState, 3000);
    return () => clearInterval(id);
  });

  async function upload(event: SubmitEvent) {
    event.preventDefault();
    submitting = true;
    error = '';
    const form = event.currentTarget as HTMLFormElement;
    const fd = new FormData(form);
    try {
      const res = await fetch(`/api/documents/${data.documentId}/upload`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          content: fd.get('content'),
          scenario: fd.get('scenario') ?? 'happy_path',
        }),
      });
      const body = await res.json();
      if (body.error) {
        error = body.error;
      } else {
        await refreshState();
      }
    } catch (err) {
      console.error('Upload error:', err);
      error = 'Upload failed. Please try again.';
    }
    submitting = false;
  }

  async function ask(event: SubmitEvent) {
    event.preventDefault();
    submitting = true;
    error = '';
    const form = event.currentTarget as HTMLFormElement;
    const fd = new FormData(form);
    try {
      const res = await fetch(`/api/documents/${data.documentId}/questions`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          question: fd.get('question'),
          scenario: fd.get('scenario') ?? 'happy_path',
        }),
      });
      const body = await res.json();
      if (body.error) {
        error = body.error;
      } else {
        form.reset();
        await refreshState();
      }
    } catch (err) {
      console.error('Ask error:', err);
      error = 'Request failed. Please try again.';
    }
    submitting = false;
  }

  async function end() {
    submitting = true;
    error = '';
    try {
      await fetch(`/api/documents/${data.documentId}/end`, { method: 'POST' });
      // Set ended optimistically so the UI transitions immediately. The
      // workflow may take a moment to complete, so refreshState() is called
      // with phase already 'ended' to avoid a transient regression to
      // 'summarised' while the signal is being processed.
      phase = 'ended';
      await refreshState();
    } catch (err) {
      console.error('End session error:', err);
      error = 'Failed to end session. Please try again.';
    }
    submitting = false;
  }
</script>

<Breadcrumbs
  items={[
    { label: 'Documents', href: '/documents' },
    { label: data.documentId, href: `/documents/${data.documentId}` },
  ]}
/>

{#if phase === 'ended'}
  <!-- ── Ended notice ───────────────────────────────────────── -->
  <section class="mb-5">
    <p class="has-text-grey">Session ended.</p>
  </section>

  {#if summary}
    <!-- ── Summary (read-only) ──────────────────────────────── -->
    <section class="mb-5">
      <h2 class="subtitle is-5">Summary</h2>
      <p>{summary}</p>
      <p class="is-size-7 has-text-grey">
        Provider: {provider}{fallbackOccurred ? ' (fallback used)' : ''}
      </p>
    </section>
  {/if}

  {#if qa.length > 0}
    <!-- ── Conversation history ──────────────────────────────── -->
    <section class="mb-5">
      <h2 class="subtitle is-5">Conversation</h2>
      <div class="mt-5 is-flex is-flex-direction-column">
        {#each [...qa].reverse() as item}
          <div class="card">
            <div class="card-content">
              <div class="content">
                <p class="mb-1"><strong>Q:</strong> {item.question}</p>
                <p class="mb-0"><strong>A:</strong> {item.answer}</p>
              </div>
            </div>
          </div>
        {/each}
      </div>
    </section>
  {/if}
{:else if phase === 'pending'}
  <!-- ── Upload ─────────────────────────────────────────────── -->
  <section class="mb-5">
    <h2 class="subtitle is-5">Process a document</h2>
    <form onsubmit={upload}>
      <div class="field">
        <label class="label" for="content">Document content</label>
        <div class="control">
          <textarea
            class="textarea"
            id="content"
            name="content"
            rows="10"
            required
          ></textarea>
        </div>
      </div>

      <div class="field">
        <label class="label" for="scenario">Scenario</label>
        <div class="control">
          <div class="select">
            <select id="scenario" name="scenario">
              <option value="happy_path">Happy path</option>
              <option value="fail_once_summarise">Fail once on summarise</option
              >
              <option value="primary_provider_failure"
                >Primary provider failure</option
              >
            </select>
          </div>
        </div>
      </div>

      {#if error}
        <p class="help is-danger">{error}</p>
      {/if}

      <div class="field">
        <div class="control">
          <button
            type="submit"
            class="button is-primary"
            class:is-loading={submitting}
            disabled={submitting}
          >
            Process document
          </button>
        </div>
      </div>
    </form>
  </section>
{:else if phase === 'processing'}
  <!-- ── Processing ─────────────────────────────────────────── -->
  <section class="mb-5">
    <p class="has-text-grey">
      Processing&hellip;
      <button
        type="button"
        class="button is-small is-ghost"
        onclick={refreshState}>Refresh</button
      >
    </p>
  </section>
{:else if phase === 'summarised'}
  <!-- ── Summary ────────────────────────────────────────────── -->
  <section class="mb-5">
    <h2 class="subtitle is-5">Summary</h2>
    <p>{summary}</p>
    <p class="is-size-7 has-text-grey">
      Provider: {provider}{fallbackOccurred ? ' (fallback used)' : ''}
    </p>
  </section>

  <!-- ── Q&A ────────────────────────────────────────────────── -->
  <section class="mb-5">
    <h2 class="subtitle is-5">Ask a question</h2>

    <form onsubmit={ask}>
      <div class="field">
        <label class="label" for="question">Question</label>
        <div class="control">
          <input
            class="input"
            id="question"
            name="question"
            type="text"
            required
          />
        </div>
        {#if submitting}
          <p class="help is-info">Thinking&hellip;</p>
        {/if}
      </div>

      <div class="field">
        <label class="label" for="qa-scenario">Scenario</label>
        <div class="control">
          <div class="select">
            <select id="qa-scenario" name="scenario">
              <option value="happy_path">Happy path</option>
              <option value="fail_once_summarise">Fail once on summarise</option
              >
              <option value="primary_provider_failure"
                >Primary provider failure</option
              >
            </select>
          </div>
        </div>
      </div>

      {#if error}
        <p class="help is-danger">{error}</p>
      {/if}

      <div class="field is-grouped">
        <div class="control">
          <button
            type="submit"
            class="button is-primary"
            class:is-loading={submitting}
            disabled={submitting}
          >
            Ask
          </button>
        </div>
        <div class="control">
          <button
            type="button"
            class="button"
            disabled={submitting}
            onclick={end}>End session</button
          >
        </div>
      </div>
    </form>

    {#if qa.length > 0}
      <div class="mt-5 is-flex is-flex-direction-column">
        {#each [...qa].reverse() as item}
          <div class="card">
            <div class="card-content">
              <div class="content">
                <p class="mb-1"><strong>Q:</strong> {item.question}</p>
                <p class="mb-0"><strong>A:</strong> {item.answer}</p>
              </div>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </section>
{:else if phase === 'failed'}
  <!-- ── Failed ─────────────────────────────────────────────── -->
  <section class="mb-5">
    <p class="has-text-danger">
      Processing failed. Check the Temporal UI for details.
    </p>
  </section>
{/if}
