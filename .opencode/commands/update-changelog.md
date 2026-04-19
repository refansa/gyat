---
description: Add unreleased changes OR release a new version
argument-hint: "[version]"
---

Update the CHANGELOG.md file for a new release.

If $ARGUMENTS is empty:
  - This is an UNRELEASED change. Add it to the `[Unreleased]` section under the appropriate category.
  - Ask the user what they changed and in which category it belongs (Added, Changed, Deprecated, Removed, Fixed, Security).
  - Write a description in imperative mood.

If $ARGUMENTS has a version (e.g., "2.8.0"):
  - This is a RELEASE. Move the `[Unreleased]` section content to a new version section with today's date (YYYY-MM-DD format).
  - Create a new empty `[Unreleased]` section at the top.
  - Update all version links at the bottom.
  - Use Keep a Changelog format.

Steps for release:
1. Read current CHANGELOG.md
2. Extract content from `[Unreleased]` section
3. Add new version section with today's date
4. Add new empty `[Unreleased]` section
5. Update version links at the bottom
6. Write the updated CHANGELOG.md