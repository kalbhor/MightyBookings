<html>
	<head>
		<form id = "Movie" action="/adddata" method="post" align="center">
			Hall Name:<br>
			<input type="text" name="hallname"><br>
			
			Movie Name:<br>
			<input type="text" name="moviename"><br>

			Movie Description:<br>
			<input type="text" name="description"><br>

			Movie Trailer:<br>
			<input type="text" name="trailer"><br>

			Movie Poster:<br>
			<input type="text" name="poster"><br>

			Movie Time:<br>
			<input type="time" name="time"><br>

			Movie Date:<br>
			<input type="date" name="date"><br>

			<!--Categories-->
			<script src="static/categories.js" language="Javascript" type="text/javascript"></script>
    		<div id="dynamicInput">
     		</div> 
     <input type="button" value="Add Category" onClick="addInput('dynamicInput');">

     <button type="submit" form="Movie" value="Submit">Submit</button>

		</form>
	</head>
</html>
