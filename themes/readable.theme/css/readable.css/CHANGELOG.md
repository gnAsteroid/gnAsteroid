# Changelog

All notable changes to the project will be noted in this file.

## v1.1.0 - 2023-04-14

### Added
- Added navbar styles
	- `none`, `default`, `classy`, `blockout`, `boxes`, `roundesque`
- Add `<button>` styling
- Added context-aware vertical margin to `<details>`
- Added CHANGELOG
- Added readable.min.css

### Fixed
- Fixed sibling selectors, which were using `~` instead of `+`

### Changed
- Moved to 0BSD from the Unlicense
- Darkened background in dark theme
- Removed vertical margin from nested lists
- Applied navbar styling to *all* <nav> elements, not just the first direct child of a body or header (each can have its own style)
	- Made the default navbar theme non-animated (to revert, use `data-style="classy"` on the `<nav>` element)

## v1.0.1 - 2023-02-15

### Added
- Browsers sending prefers-contrast: high will now receive high-contrast versions of the dark and light themes with white on black and black on white respectively. Sites can explicitly set this using data-high-contrast="on" as an attribute on the main `<html>` tag of a page. Note: this does not affect custom themes.
- `<video>` elements are now treated the same way as `<img>` elements.

### Fixed
- `<nav>` tags which are direct children of `<header>` elements can now be considered main navigation by the code, applying styling properly.

## v1.0.0

This is the initial public release of readable.css.
