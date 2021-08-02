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
     data-ham-layout='../layouts/default.html'
     data-ham-layout-css='[
     "../assets/css/test.css",
     "../assets/css/test2.css"
     ]'
     data-ham-layout-js='[
     "../assets/js/test.js",
     "../assets/js/test.js"
     ]'
>
  <embed type="ham/partial" src="../partials/header.html"/>
</div>
```
### Partials
Partials are reusable html modules that can be included on a page or layout
```html
<h1>Welcome to HAM</h1>
```
### INSTALLING HAM
`go install github.com/fobilow/ham@latest`

For specific version, replace @latest with version number

### USING HAM
* ham new [sitename]
* ham build -w [working dir] -out [output directory]
* ham serve -w [working dir]
* ham version
* ham help
