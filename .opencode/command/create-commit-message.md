---
description: Generate a concise, human-friendly git commit message from the current changes.
tools: ['gitkraken/git-mcp/changes', 'git/cli/diff']
---

## Purpose

Write a single commit message (subject + optional body) describing the provided code changes. This command should only produce the commit message text; do not run any git commands or suggest commands to the user.

## Inputs

```text
$ARGUMENTS
```

If available, inspect changes with Gitkraken MCP; if that isn't available, fall back to using git to determine the diff. The implementation/agent running this command may read the working tree, staged changes, and recent commits as needed — your output must still be only the commit message.

## Commit message requirements

- Subject: one short line, imperative mood, ~50 characters max
- Optional body separated by a blank line; explain motivation, what changed, and important notes (tests, migrations, breaking changes)
- Use active voice and keep it human and story-like
- Prefer short paragraphs and a small bulleted list for multi-point changes
- Output only the message enclosed in a code block for easy copying

## Suggested structure (template)

- Subject (imperative, ~50 chars)

- Short summary: one sentence elaborating the subject.

- Motivation: 1–2 sentences explaining why this change was needed.

- What I changed:
  - Bullet list of high-level changes (one bullet per change)

- Notes: any important risks, tests added, or follow-ups

## Style tips

- Start with the user problem or integration need when helpful.
- Mention trade-offs or why alternatives were rejected in one sentence.
- Avoid low-level diffs; surface high-level intent and key touched areas.

## Examples

Small change:

```text
fix: ensure user avatar loads on profile page

Load user avatars asynchronously and add a placeholder while fetching to
prevent layout shift. This fixes a race where avatars were blank on slow
networks.

What I changed:
- Defer avatar rendering until image fetch resolves
- Add placeholder skeleton component

Notes: covered by new integration test, backward-compatible.
```

Larger change:

```text
feat: add ChatHookProvider to chatPromptFiles API

Wire hooks through the chatPromptFiles API so providers can enumerate
and surface individual hook entries to the core system.

Motivation: providers previously exposed hooks as file containers which
made it hard for the UI to list individual hooks.

What I changed:
- Add `ChatHookProvider` interface with `onDidChangeHooks`/`provideHooks`
- Add `chat.hooks` getter and `chat.onDidChangeHooks` event
- Add `chat.registerHookProvider()` registration API

Notes: Added unit tests for provider expansion. No breaking changes.
```

## Final rules

- Only output the commit message text (subject + optional body). Do not run or suggest any commands.
- Keep the subject within 50 characters when possible.
- Prefer human, concise, and intent-focused language.
