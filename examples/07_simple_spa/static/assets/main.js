(function() {
    let value = 0;
    const btn = document.getElementById("pusher")
    const val = document.getElementById("value")

    function updateVal() {
        val.innerHTML = value;
    }

    btn.onclick = function() {
        value++;
        updateVal();
    }

    updateVal();
}())