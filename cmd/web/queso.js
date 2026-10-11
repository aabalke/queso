document.addEventListener("goready", () => {

    const inputs = document.getElementsByClassName("queso input")

    for (var i = 0; i < inputs.length; i++) {
        inputs[i].addEventListener("input", () => {
            goApp.Decode(inputs[i]);
        });
    }
});
