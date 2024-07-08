HAM: Html As Modules
====

HAM makes modular HTML development possible. HAM provides a framework and compilier
for developing static HTML web pages

HAM CONCEPTS
====

### Layouts
A Layout defines the structure of a web page
```html
<html lang="en">
<head>
     <meta charset="UTF-8">
     <title>HAM</title>
     <link type="ham/layout-css"/>
</head>
<body>

<div id="app-info"></div>
<div class="container">
     <div class="row">
          <embed type="ham/page"/>
     </div>
</div>
<embed type="ham/layout-js"/>
</body>
</html>
```

### Pages
A Page must have a layout. A page gives a layout content
```html
<div class="page"
     data-ham-page-config='{
     "layout": "../layouts/default.html",
     "css": ["../assets/app/css/index.css"],
     "js": ["../assets/app/js/index.js"]
     }'
>
  <embed type="ham/partial" src="../partials/header.html"/>
</div>
```
### Partials
Partials are reusable html modules that can be included on a page or layout
```html
<h1>Welcome to HAM</h1>
```

### Final Result
```html
<html lang="en">
<head>
  <meta charset="UTF-8"/>
  <title>HAM</title>
  <link rel="stylesheet" href="/assets/app/css/index.css?v=202102082042"/>
</head>
<body>
<div id="app-info"></div>
<div class="container">
  <div class="row">
    <div class="page">
      <h1>Welcome to HAM</h1>
    </div>
  </div>
</div>
<script src="../assets/app/js/index.js?v=202102082042"></script>
</body>
</html>
```

### INSTALLING HAM
`go install github.com/fobilow/ham/cmd/ham@latest`

For specific version, replace @latest with version number

### USING HAM
* ham init [sitename]
* ham build -w [working dir] -o [output directory]
* ham serve -w [working dir]
* ham version
* ham help
