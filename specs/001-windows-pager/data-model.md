# Data Model: windows-pager

This document extracts entities and state relevant to the pager feature.

## Entities

- Pager Session
  - id: ephemeral unique identifier for a session
  - cursor_position: integer line offset
  - viewport_height: integer
  - viewport_width: integer
  - search_query: optional string
  - matches: list of match positions (ints)
  - mode: enum {viewing, searching}

- Output Stream
  - raw_bytes: the raw bytes produced by a command
  - is_text: boolean
  - is_tty: boolean

## Validation Rules

- When is_tty is false, pager MUST NOT be invoked.
- For is_text false (binary data), pager MUST be bypassed to avoid corruption.
