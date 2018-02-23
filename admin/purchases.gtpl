<!DOCTYPE HTML>
<html lang="en">
  <head>
    <meta charset="utf-8"/>
    <title>Purchases</title>
  </head>
  <body>
<table border="1">
<tr>
<td>Name</td>
<td>Phone</td>
<td>Email</td>
<td>Quantity</td>
<td>Total Price</td>
<td>Hall</td>
<td>Category</td>
<td>Movie</td>
</tr>
{{ if . }}
       {{ range . }}
<tr>
<td>{{ .Name }}</td>
<td>{{ .Phone}}</td>
<td>{{ .Email}}</td>
<td>{{ .Quantity}}</td>
<td>{{ .Quantity * .Show.Categories[0].Price }}</td>
<td> {{ .Show.HallName }} </td>
<td>{{ .Show.Categories[0].Name }} </td>
<td> {{ .Show.Movie.Name }} </td>
</tr>
{{ end }}
     {{ end }}</table>
</body>
</html>
