---
name: feedback-en
description: Generate stories from user feedbacks, filtering already processed ones
---

# Skill: Feedback → Stories / Specification (EN)

This guide describes how to transform user feedbacks into actionable user stories, processing only the most recent feedbacks that have not yet been addressed by existing stories.

## What I Do

I read the `docs/specification/feedbacks.md` file, identify feedbacks that have not yet been processed. I **first group feedbacks by theme** — multiple feedbacks on the same topic become a single unit of work. For each group:
- **If the topic is simple** → I generate one or more stories directly (same format as story-en) with the iterative question/answer process.
- **If the topic is complex** (cross-cutting new feature, multiple sub-topics, architectural impact) → I create a sub-specification in `docs/specification/features/`, refine it via the specification-en process, then break it into stories via story-en.

**Rule:** multiple related feedbacks → one group → one story (if simple) or one sub-specification (if complex). No duplicates.

## When to Use Me

Use this skill when:
- You have collected user feedbacks in `docs/specification/feedbacks.md`
- You want to transform these feedbacks into structured, testable stories
- You want to avoid reprocessing already handled feedbacks

## Prerequisites

The `docs/specification/feedbacks.md` file must be structured as follows:

```markdown
## 2025-05-14

* Feedback 1
* Feedback 2
...

## [DATE]

* Feedback N
...
```

Each processed feedback gets a `[PROCESSED: story-XXX]` marker:

```markdown
* Feedback 1 [PROCESSED: story-018]
```

## Workflow

### Step 1 - Read the feedback file

- Read `docs/specification/feedbacks.md`
- Find feedbacks NOT marked `[PROCESSED: story-XXX]` or `[PROCESSED: feature-XXX.md]` (these are the ones to process)
- If ALL feedbacks are marked: announce nothing new to process → end of workflow

### Step 2 - Analyze existing stories

- Read stories in `docs/stories/done/` or `docs/stories/`
- Check if some unmarked feedbacks are already covered by existing stories
  - Example: "A screen with the list of zones" might already exist in Story #009
  - If so, add the `[PROCESSED: story-XXX]` marker (already covered) and do not recreate the story
- Remaining feedbacks (neither marked PROCESSED nor covered) are to be transformed into new stories

### Step 3 - Grouping and decision: direct story or sub-specification

- Group feedbacks by common theme (e.g., "editor", "review", "UI/UX", "archives")
- A single feedback can become a story if it is rich enough
- Multiple related feedbacks can be grouped into the same story
- **Decide the complexity level** of each group (or single feedback):
  - **Simple topic** → go to **Step 4a** (direct story generation)
  - **Complex topic** → go to **Step 4b** (create a sub-specification)
- Indicative complexity criteria (one is enough):
  - Multiple distinct roles or permissions
  - Multiple impacts on existing features
  - Numerous independent sub-topics
  - Architectural change or new database table
  - Requires several stories to cover the scope

---

### Path A — Simple topic: Direct story generation

#### Step 4a - Generate stories

- Read stories in `docs/stories/done/` to determine the next story number
- Create new stories in `docs/stories/` following the story-en format:

```markdown
# Story #XXX : [Title]

## Description
As a [role], I want to [action] so that [benefit].

## Origin
Feedback from [DATE]: "[feedback text]"

## Acceptance criteria
- [ ] Criterion 1
- [ ] Criterion 2

## Automated tests
- Unit test: [Description]
- Integration test: [Description]
- E2E test: [Description]

## Documentation
- [New documentation to create]

## User value
[Description of the compelling value for the end user]

## Dependencies
- [Story #YYY if applicable]
```

- Add an `## Origin` section to trace the source feedback
- Mark each processed feedback in `docs/specification/feedbacks.md` with `[PROCESSED: story-XXX]`
- Example: `* Do not show a home page [PROCESSED: story-018]`

#### Step 5a - Create questions file

- Create `docs/stories/questions.md` (or append to it if it already exists)
- Contains clarification questions for the newly generated stories
- Same format as story-en (context, question, proposals)

#### Step 6a - Analysis and clarification

- Analyze the new stories
- Check consistency with the specification `docs/specification/specification.md`
- Check consistency with existing stories in `docs/stories/done/`
- Add questions in `docs/stories/questions.md` if needed

#### Step 7a - Update

- The user answers the questions in `docs/stories/questions.md`
- The user types: `I've answered the questions, update the stories`
- Update the stories according to the answers
- Return to Step 6a

---

### Path B — Complex topic: Sub-specification then story breakdown

#### Step 4b - Create the sub-specification

- Create the `docs/specification/features/` directory if it doesn't exist
- Create `docs/specification/features/feedback-XXX.md` (XXX = short topic name, e.g., `feedback-match-archives.md`)
- Copy the feedback text into the file as a specification base
- Add an `## Origin` section with the date and source feedback text
- Structure the file with `[TO_CLARIFY: ...]` placeholders for unclear points
- Create `docs/specification/features/questions.md` with:
  - The context (feedback description)
  - Clarification questions (open-ended, with proposals)
  - Detected inconsistencies
- **Mark the feedback** in `docs/specification/feedbacks.md` with the `[PROCESSED: feature-XXX.md]` marker

#### Step 5b - Specification iteration

- The user answers the questions in `docs/specification/features/questions.md`
- The user types: `I've answered the questions, continue`
- Read the answers, update `feedback-XXX.md` by replacing placeholders with answers
- Resolve identified inconsistencies
- Check if any `[TO_CLARIFY: ...]` or `[INCONSISTENCY: ...]` remain:
  - **If any remain** → generate a new `questions.md` with remaining questions → return to the start of Step 5b
  - **If none** → delete `questions.md` → proceed to Step 6b

#### Step 6b - Break down into stories (story-en)

- The sub-specification is complete (no more placeholders)
- Run the story-en workflow on `docs/specification/features/feedback-XXX.md`:
  - Read the specification
  - Break into user stories in `docs/stories/` (each with an `## Origin` section pointing to `feature-XXX.md`)
  - Create `docs/stories/questions.md` if points remain unclear
  - Apply the iterative question/answer cycle from story-en
- Once stories are finalized (no more questions):
  - Update `feedbacks.md` if sub-feedbacks need individual marking
  - Announce the created stories

---

### Step 8 - End of workflow

- When nothing remains to clarify (for all paths)
- Delete the `questions.md` file if it exists
- Move new stories to `docs/stories/done/` (or leave them in `docs/stories/` per convention)
- Update `AGENTS.md` and `workflow.md` if necessary
- Announce the end of the workflow

## Feedback file format

```markdown
## [DATE]

* [Feedback text] [PROCESSED: story-XXX]
* [Feedback text] [PROCESSED: feature-XXX.md]
* [Feedback text]

## [OTHER DATE]

* [Feedback text] [PROCESSED: story-XXX]
```

Feedbacks marked `[PROCESSED: story-XXX]` or `[PROCESSED: feature-XXX.md]` are ignored.
Feedbacks without a marker are to be processed.

## Example

### Initial feedbacks.md file
```markdown
## 2025-05-14

* Do not show a home page: go directly to "My teams"
* Show an error message when trying to open the editor and someone else has the lock [PROCESSED: story-010]
* Editor help popup: display the list of commands and associated keyboard shortcuts
```

### Generated stories
- `docs/stories/story-018.md`: "Remove the home page" (based on feedback #1)
- `docs/stories/story-019.md`: "Editor help popup" (based on feedback #3)

### feedbacks.md after processing
```markdown
## 2025-05-14

* Do not show a home page: go directly to "My teams" [PROCESSED: story-018]
* Show an error message when trying to open the editor and someone else has the lock [PROCESSED: story-010]
* Editor help popup: display the list of commands and associated keyboard shortcuts [PROCESSED: story-019]
```

## How to move from one step to another

### Path A — Simple topic

| Transition | User Action | Agent Action |
|------------|-------------|--------------|
| Step 1 → 2 | Provide the feedback file path as argument | Reads feedbacks, analyzes existing stories |
| Step 2 → 3 | Automatic | Groups unprocessed feedbacks by theme and decides simple/complex |
| Step 3 → 4a | Automatic (simple topic) | Generates new stories in docs/stories/ |
| Step 4a → 5a | Automatic after generation | Creates questions.md with points to clarify |
| Step 5a → 6a | Automatic | Analyzes stories and adds questions |
| Step 6a → 7a | Type "I've answered the questions, update the stories" | Updates stories according to answers |
| Step 7a → 6a | Automatic after update | Re-checks consistency |
| Step 7a → 8 | Automatic (no more questions) | Deletes questions.md, marks processed feedbacks, announces completion |

### Path B — Complex topic

| Transition | User Action | Agent Action |
|------------|-------------|--------------|
| Step 1 → 2 | Provide the feedback file path as argument | Reads feedbacks, analyzes existing stories |
| Step 2 → 3 | Automatic | Groups unprocessed feedbacks by theme and decides simple/complex |
| Step 3 → 4b | Automatic (complex topic) | Creates docs/specification/features/feedback-XXX.md + questions.md |
| Step 4b → 5b | Automatic | Waits for user answers |
| Step 5b → 5b | Type "I've answered the questions, continue" | Updates spec, checks remaining placeholders |
| Step 5b → 6b | Automatic (no more placeholders) | Deletes questions.md, runs story-en on the sub-specification |
| Step 6b → 8 | Automatic (stories finalized) | Announces created stories |

**Important:** The agent does not monitor file changes in real time. The user must always give an explicit instruction for the agent to move to the next step.

## Usage examples

### Simple case — Direct story generation

```
/ feedback feedbacks.md
```

The agent reads `docs/specification/feedbacks.md`, identifies unmarked feedbacks, checks existing stories, and generates new stories with the iterative clarification process.

### Complex case — Sub-specification then stories

```
/ feedback feedbacks.md
```

The agent reads `docs/specification/feedbacks.md`, detects that a feedback (e.g., "Match archives") is too complex for a single story, then:

1. Creates `docs/specification/features/feedback-match-archives.md`
2. Generates `docs/specification/features/questions.md` with points to clarify (roles, data relationships, etc.)
3. Iterates with the user until the specification is complete
4. Deletes `questions.md`
5. Runs story-en on the specification to break it into multiple stories
6. Marks the original feedback as `[PROCESSED: feature-match-archives.md]`

### Handling duplicates with existing stories

If an unmarked feedback is already covered by an existing story:

1. The agent identifies it during Step 2
2. It adds the `[PROCESSED: story-XXX]` marker in the feedbacks.md file
3. It does not generate a new story
4. It can notify the user of the duplicate in questions.md
