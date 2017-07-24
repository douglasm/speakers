function toggle_password(target1, target2){
    var d = document;
    var tag0 = d.getElementById("showhide");
    var tag1 = d.getElementById(target1);
    var hasTwo = false;
    if(target2 != ""){
        var tag2 = d.getElementById(target2);
        hasTwo = true;
    }

    attr = tag1.getAttribute('type');
    if (attr == 'password'){
        tag1.setAttribute('type', 'text');
        if(hasTwo){
            tag2.setAttribute('type', 'text');
        }
        tag0.innerHTML = 'Hide password';

    } else {
        tag1.setAttribute('type', 'password');
        if(hasTwo){
            tag2.setAttribute('type', 'password');
        }
        tag0.innerHTML = 'Show password';
    }
}
