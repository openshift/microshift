const address = document.getElementById("address");
const output = document.getElementById("output");
const result = document.getElementById("result");
const button = document.getElementById("enable_console");
const link = document.getElementById("divCheckbox");
function cmd_run() {
    /* global cockpit */
    cockpit.spawn(["bash","-c","/usr/share/cockpit/cockpit-console/enable_console.sh"])
            .then(cmd_success)
            .catch(cmd_fail);

    result.textContent = "";
}

function cmd_success(console_url) {
    window.parent.open(console_url);
}

function cmd_fail(cmd_output) {
    result.style.color = "red";
    result.textContent = "cmd failed:"+cmd_output;
}


// Connect the button to starting the "ping" process
button.addEventListener("click", cmd_run);

// Send a 'init' message.  This tells integration tests that we are ready to go
cockpit.transport.wait(function() { });
