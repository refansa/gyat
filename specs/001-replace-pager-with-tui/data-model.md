Entities

- ViewerSession
  - id: string
  - cursorLine: int
  - viewportHeight: int
  - viewportWidth: int
  - searchQuery: string
  - matches: []int
  - mode: string (e.g., "normal", "search")

Relationships

- Command output -> ViewerSession: A viewer session is created per command invocation when stdout is a TTY and paging is active.

Validation rules

- cursorLine >= 0 and <= max(0, len(lines)-viewportHeight)
- viewportHeight > 0
- match indexes within [0, len(lines)-1]

State transitions

- initial -> rendered (after Render is called)
- rendered -> interactive (after RunInteractiveSession starts)
- interactive -> closed (on quit)
