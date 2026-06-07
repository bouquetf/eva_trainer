---
name: commit-en
description: Proposes a conventional commit message in English based on a story and local changes
---

## What I do

I take a story reference (e.g. `#031` or `story-031`), inspect the locally modified files (`git status`, `git diff`), analyze them in the context of that story, and propose a well-formatted commit message in English following the project's conventional commit style.

## When to use me

After implementing the tasks of a story, when you are ready to commit your changes and want a descriptive, consistent commit message that references the story.

## Workflow

### Step 1 — Read the story

- Take the story reference provided as argument (e.g. `/ commit-en 031`)
- Read the file `docs/stories/story-XXX.md` to understand the story purpose, acceptance criteria, and context
- If the story file does not exist, note it and proceed with only the git context

### Step 2 — Inspect local changes

- Run `git status` to see staged/unstaged/untracked files
- Run `git diff --cached` and `git diff` to examine the actual changes
- Run `git log --oneline -5` to see the recent commit style on this branch
- Categorize the changes: new feature, bug fix, refactoring, tests, docs, chore, etc.

### Step 3 — Analyze in story context

For each modified file, determine:
- Is this implementing a new story feature?
- Is this fixing a bug discovered during the story?
- Is this test code for the story?
- Is this refactoring or docs related to the story?

Identify the primary type and scope of the change.

### Step 4 — Propose a commit message

Generate exactly **one** commit message proposal in the project's conventional commit format:

```
<type>(<scope>): <short description>

Story #XXX

- <bullet point summarizing a change>
- <bullet point summarizing another change>
```

**Type** (pick one):
- `feat` — new feature
- `fix` — bug fix
- `refactor` — code restructuring
- `test` — tests only
- `doc` — documentation
- `chore` — maintenance, deps, tooling

**Scope** (optional, inferred from the story or changed files — e.g. `editor`, `teams`, `auth`, `api`, `db`)

**Short description**: imperative mood, no period, ≤ 72 chars.

**Body**: bullet points summarizing key changes, referencing files or areas when helpful.

Present the proposal to the user with a brief explanation of why this type/scope was chosen.

### Step 5 — Revise (optional)

If the user requests adjustments (different type, more detail, squash vs split), revise the proposal accordingly. Iterate until the user is satisfied.

## Usage examples

```
/ commit-en 031
Reads docs/stories/story-031.md, inspects git changes, proposes:

feat(teams): reorganize team settings into collapsible sections

Story #031

- Add CollapsibleSection reusable accordion component with 9 unit tests
- Reorder TeamSettings page (Infos → Cartes → Membres → Invitations → Quitter → Danger)
- Fix 12 existing E2E tests that relied on default-open sections
- Update selectors in roles, invitations, deletion, and transfer specs
```
