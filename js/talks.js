window.onload = function()
{
  // Menu initialization statements.
    // post('/talkslist/', {name: 'Johnny Bravo'});
}

function post(path, params, method) {
    method = method || "post"; // Set method to post by default if not specified.

    // The rest of this code assumes you are not using a library.
    // It can be made less wordy if you use one.
    var form = document.createElement("form");
    form.setAttribute("method", method);
    form.setAttribute("action", path);

    for(var key in params) {
        if(params.hasOwnProperty(key)) {
            var hiddenField = document.createElement("input");
            hiddenField.setAttribute("type", "hidden");
            hiddenField.setAttribute("name", key);
            hiddenField.setAttribute("value", params[key]);

            form.appendChild(hiddenField);
         }
    }

    document.body.appendChild(form);
    form.submit();
}

function placeclick(cb){
    var xmlHttp = new XMLHttpRequest();
    var url = "/setplace";

    xmlHttp.open("POST", url, true);
    xmlHttp.setRequestHeader("Content-type", "application/x-www-form-urlencoded");
    xmlHttp.send("place=" + cb.value + "&sel=" + cb.checked);
}

function httpGetAsync(numUsers) {
  var xmlHttp = new XMLHttpRequest();
  var url = "/getserial";
  var cb = document.getElementById('pc');

  if(numUsers==3){
    url += "opt"
  }
  url += "/"
  url += cb.checked;
  xmlHttp.onreadystatechange = function() { 
    if (xmlHttp.readyState == 4 && xmlHttp.status == 200)
      myFunction(xmlHttp.responseText);
  }
  xmlHttp.open("GET", url, true); // true for asynchronous 
  xmlHttp.send(null);
}

function myFunction(arr) {
}
