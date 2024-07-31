The formatting is the same as gnoweb. The markup is certainly the same, 
but the css style may have a couple improvements, such as for tables and blockquotes.

# H1
## H2
### H3

Titles produced by:
```
# H1
## H2
### H3
```

`**bold**` -> **bold**.

`*italic*` buterin -> *italic* buterin.

### Blockquote

> Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. - Nero
>> Alea jacta est - Iulius Caesar
>>> Really? - Nero
>>>> Roll d100 now - The DM

How:

```
> a first quote,
> it may be on several lines. - Somebody
>> a response
>>> another response and so on.
```

### Ordered List

1. First item
2. Second item
3. Third item

Use a number followed by a dot. If you use the same number it will be automatically incremented.

### Unordered List

- First item
- Second item
- Third item

Use a star `*` or a dash `-` followed by a space.

### Mixed lists

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

You can nest lists by prepending two spaces.

# 4 spaces

    What happens when you start a line or paragraph with 4 leading spaces.

### Code

This is an inline `code` (backquoted).

Multi-line (fenced with triple backquotes):

```
Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut 
labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco 
laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in 
voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat 
non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.
```
Here is a longer sample (long line):

```
Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.
```
Fenced code block (triple backquotes followed by word "go"):
```go
import "fmt"

func main() {
    fmt.Println("hello world")
}
```
### Horizontal Rule

Something like `---` on its own line:

---

### Link

`[Markdown Guide](https://www.markdownguide.org)`:
[Markdown Guide](https://www.markdownguide.org)

You may also employ HTML `<a>` links.

### Image

`![alt text](https://www.markdownguide.org/assets/images/tux.png)`:

![alt text](https://www.markdownguide.org/assets/images/tux.png)

You may also use the HTML `<img>` tag.

### Table

| Syntax | Description |
| ----------- | ----------- |
| Header | Title |
| Paragraph | Text |


### Footnote

Here's a sentence with a footnote. [^1]

[^1]: This is the footnote.

(Note: footnotes are not supported yet)

### Strikethrough

Use `~~foo~~`: ~~foo~~

### Task List

- [x] Write the press release
- [ ] Update the website
- [ ] Contact the media

### Subscript

Use HTML `<sub>2</sub>`: H<sub>2</sub>O

### Superscript

Use HTML `<sup>n</sup>`: x<sup>n</sup>

