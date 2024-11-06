---
title: Markdown cheatsheet
date: 2023-05-01
description: Shows common markdown syntax and allow the user to see how it's rendered.
tags: [Markdown]
author: grepsuzette
---
The markdown is formatted by gnoweb, but styling may vary depending on the theme you use. An optional header in Front Matter is possible, if so the `title` will be used for page titles.

* Use `*α*` to *emphasize* something. 
* Use `**α**` to put something in **bold**.
* Use `~~α~~` for ~~strikethrough~~
* Use HTML `<sub>2</sub>` for subscript, e.g. H<sub>2</sub>O
* Use HTML `<sup>n</sup>` for superscript, e.g. x<sup>n</sup>
* Use a syntax like `[Markdown Guide](https://www.markdownguide.org)` to create links like [Markdown Guide](https://www.markdownguide.org)
* Use a target beginning by `/r/` to **render realms from gno.land**, e.g. `[gnoface](/r/demo/art/gnoface:42)` creating links to gno.land smart contracts like [gnoface](/r/demo/art/gnoface:42) (see also [on gno.land directly](https://gno.land/r/demo/art/gnoface:42))



## Lists

Begin with a star `*` or a dash `-` followed by a space for regular lists:

- First item
- Second item
- Third item

Begin with a number followed by a dot for ordered lists. If you use the same number it will be automatically incremented.

1. First item
2. Second item
3. Third item

Nest lists, creating mixed-style lists, by prepending two spaces:

- First item
- Second item
- Third item
  - a
  - a
  - a
    * abc
    * def
      1. ghi
      1. jkl
      1. mno
    * pqr
    * stu
  * vwx

## Blockquote

```
> a first quote,
> it may be on several lines. - Somebody
>> a response
>>> another response and so on.
```


> a first quote,
> it may be on several lines. - Somebody
>> a response
>>> another response and so on.

## Four spaces

    this text 
    starts with four 
    leading spaces.
    it will be shown
    verbatim

this paragraph
doesn't start with spaces.
therefore it will be shown 
normally.

## Code

These may still need some work (in CSS) with some themes.

### Inline 

This is code is inline (surrounded by backquotes): `let j ← 4`

### Multi-line, fenced with triple backquotes

We mean delimited by "```":

```
Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut 
labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco 
laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in 
voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat 
non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.
```

### Same on one very long line

```
Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.
```

### Fenced code block (triple backquotes followed by word "go"):
```go
import "fmt"

func main() {
    fmt.Println("hello world")
}
```
## Horizontal Rule

Something like `---` on its own line:

---

Note that if such a line appear on the first line, it will introduce a Front Matter header.

## Image

`![alt text](https://www.markdownguide.org/assets/images/tux.png)`:

![alt text](https://www.markdownguide.org/assets/images/tux.png)

You may also use the HTML `<img>` tag.

## Table

```
| Syntax      | Description |
| ----------- | ----------- |
| Header      | Title       |
| Paragraph   | Text        |
```

gives


| Syntax | Description |
| ----------- | ----------- |
| Header | Title |
| Paragraph | Text |


## Footnote

Here's a sentence with a footnote. [^1]

[^1]: This is the footnote.

(Note: footnotes are not supported yet)

## Task List

```
- [x] Write the press release
- [ ] Update the website
- [ ] Contact the media
```

- [x] Write the press release
- [ ] Update the website
- [ ] Contact the media


# Front Matter example

Front Matter is a simple header format like this:

```
---
title: Markdown cheatsheet
date: 2023-05-01
description: Shows common markdown syntax and allow the user to see how it's rendered.
tags: [Markdown]
author: bob
---
```
It is optional.

There are 2 ways to define a title for a page.

1. automatically: by default it will be read from the URL's path (removing the .md extension)
2. by using a Front Matter header, e.g.:

```
---
title: My awesome page title
---
```

For the **home page**, the Asteroid name is used by default, unless overwritten as above.

---

# Title level 1
## Title level 2
### Title level 3
#### Title level 4
##### Title level 5
###### Title level 6
