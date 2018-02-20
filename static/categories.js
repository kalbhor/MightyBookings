var counter = 1;
var limit = 5;
function addInput(divName){
     if (counter == limit)  {
          alert("You have reached the limit of adding " + counter + " inputs");
     }
     else {
          var newdiv = document.createElement('div');
          newdiv.innerHTML = "Category Name " + (counter) + " <br><input type='text' name='categoryname" + counter  + "'>";
          document.getElementById(divName).appendChild(newdiv);
          var newdiv = document.createElement('div');
          newdiv.innerHTML = "Category Seats " + (counter) + " <br><input type='number' name='seats" + counter + "'>";
          document.getElementById(divName).appendChild(newdiv);
          var newdiv = document.createElement('div');	
          newdiv.innerHTML = "Category Price " + (counter) + " <br><input type='number' name='price" + counter + "'>";
          document.getElementById(divName).appendChild(newdiv);
          counter++;
     }
}
