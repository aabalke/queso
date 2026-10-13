document.addEventListener("goready", () => {

    const inputs = document.getElementsByClassName("queso input")

    for (var i = 0; i < inputs.length; i++) {
        inputs[i].addEventListener("input", () => {
            goApp.Decode();
        });
    }
});

var isFileInput = false;
var fileInputSrc = document.getElementById("file_input_src")

function toggleFileInput(enabled) {

    var inputContainer = document.getElementById("hex_input_container")

    if (enabled) {
        inputContainer.style.display = "none"
    } else {
        inputContainer.style.display = "flex"
    }

    fileInputSrc.disabled = !enabled

    isFileInput = enabled

    goApp.Decode();
}


fileInputSrc.addEventListener("change", async () => {
    const file = fileInputSrc.files[0];
    if (!file) return;

    if (file.size > KB * 256) {
        document.getElementById("disassembly").textContent = "file size " + getSizeString(file.size) + " > max of 256kb"
        return
    }

    const bytes = new Uint8Array(await file.arrayBuffer());
    goApp.ProcessFileUpload(file.name, bytes);
    goApp.Decode();
});

const KB = 1024
const MB = KB * 1024
const GB = MB * 1024

function getSizeString(size) {

    if (size / GB > 1) {
        return size / GB + "gb"
    }
    if (size / MB > 1) {
        return size / MB + "mb"
    }
    if (size / KB > 1) {
        return size / KB + "kb"
    }

    return size + " bytes"
}

var showOptions = true;
var arrow = document.getElementById("options_arrow")
arrow.innerHTML = "&#9660;"

function toggleOptions() {
    var arrow = document.getElementById("options_arrow")
    var container = document.getElementById("options_container")

    showOptions = !showOptions

    if (showOptions) {
        arrow.innerHTML = "&#9660;"
        container.style.display = "flex"
    } else {
        arrow.innerHTML = "&#9654;"
        container.style.display = "none"
    }
}


