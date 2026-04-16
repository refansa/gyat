# data-model.md

## Entities

- Repository Entry
  - id: string (unique identifier; e.g., path or manifest key)
  - display_name: string
  - path: string
  - current_branch: string | null
  - summary_state: string (one-line status summary)
  - metadata: object (optional key/value map for extra info)

- Repository Status View
  - repo_id: string (links to Repository Entry)
  - tabs: list of Status Tab
  - active_tab: string

- Status Tab
  - id: string
  - title: string
  - content: string | structured object

## Validation Rules

- Repository Entry must include id and display_name.
- Tabs list must include at least one tab (default: "Overview").
- When producing JSON output, keys must be stable and typed as specified in contracts/*.json

## State Transitions

- Selection
  - previous: repo_id | null
  - next: repo_id | null

- Tab Switch
  - previous_tab: string
  - active_tab: string
