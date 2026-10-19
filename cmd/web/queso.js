document.addEventListener("goready", () => {

    const inputs = document.getElementsByClassName("queso_input")

    for (var i = 0; i < inputs.length; i++) {
        inputs[i].addEventListener("input", () => {
            goApp.Decode();
        });
    }

    archInput = document.getElementById("arch_version")

    archInput.addEventListener("change", () => {
        goApp.Decode();
    });
});

var isFileInput = false;
var fileInputSrc = document.getElementById("file_input_src")

function toggleFileInput(enabled) {

    var left = document.getElementById("left")
    var gutter = document.getElementById("gutter")

    if (enabled) {
        left.style.display = "none"
        gutter.style.display = "none"
    } else {
        left.style.display = "flex"
        gutter.style.display = "block"
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

const gutter = document.getElementById('gutter');

gutter.addEventListener('pointerdown', (e) => {

    e.preventDefault();
    gutter.setPointerCapture(e.pointerId);

    const onMove = (e) => {
        const split = document.getElementById('split');
        const left = document.getElementById('left');
        const rect = split.getBoundingClientRect();
        const x = e.clientX - rect.left - (gutter.offsetWidth);
        const min = 100;
        const max = rect.width - 100 - gutter.offsetWidth;
        left.style.width = Math.min(Math.max(x, min), max) + 'px';
    };

    const onUp = () => {
        gutter.removeEventListener('pointermove', onMove);
        gutter.removeEventListener('pointerup', onUp);
    };

    gutter.addEventListener('pointermove', onMove);
    gutter.addEventListener('pointerup', onUp);
});
