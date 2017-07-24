function getCheckedBoxes(url) {
  var xmlHttp = new XMLHttpRequest();
  var placeBoxes = document.getElementsByName("places");
  var placesChecked = [];
  // loop over them all
  for (var i=0; i<placeBoxes.length; i++) {
     // And stick the checked ones onto an array...
     if (placeBoxes[i].checked) {
        placesChecked.push(placeBoxes[i].id);
     }
  }
  // Return the array if it is non-empty, or null
  xmlHttp.onreadystatechange = function() { 
    if (xmlHttp.readyState == 4 && xmlHttp.status == 200)
      keyFunction(xmlHttp.responseText);
  }

  xmlHttp.open("POST", url, true);
  xmlHttp.setRequestHeader("Content-type", "application/x-www-form-urlencoded");
  xmlHttp.send("places=" + placesChecked);
}

function keyFunction(arr) {
  document.getElementById("tab").innerHTML = arr;
}
