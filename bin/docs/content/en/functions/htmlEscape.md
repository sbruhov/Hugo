---
title: htmlEscape
linktitle:
description: Returns the given string with the reserved HTML codes escaped.
date: 2017-02-01
publishdate: 2017-02-01
lastmod: 2017-02-01
categories: [functions]
menu:
  docs:
    parent: "functions"
keywords: [strings, html]
signature: ["htmlEscape INPUT"]
workson: []
hugoversion:
relatedfuncs: [htmlUnescape]
deprecated: false
aliases: []
---

In the result `&` becomes `&amp;` and so on. It escapes only: `<`, `>`, `&`, `'` and `"`.

```
{{ htmlEscape "Hugo & Caddy > WordPress & Apache" }} → "Hugo &amp; Caddy &gt; WordPress &amp; Apache"
```