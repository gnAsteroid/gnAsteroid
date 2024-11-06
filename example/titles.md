There are 2 ways to define a title for a page.

1. automatically: by default it will be read from the URL's path (removing the .md extension)
2. by using a Front Matter header, e.g.:

```
---
title: My awesome page title
---
```

For the **home page**, the Asteroid name is used by default, unless overwritten as above.

We recommend not setting a title in the markdown itself (as you would with `# My title`). First because it is difficult to enforce whether you will use `#` or `##`. Secondly because our policy is to *not* modify the provided markdown, meaning it could be problematic with themes. 

*We think in general users will find more comfortable not to have to set a title (letting it being derived from the path in the URL), a major selling point of using asteroids being to create content rapidly.*
