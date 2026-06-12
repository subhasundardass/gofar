function runUI(res) {
  for (const cmd of res.commands) {
    if (cmd.type === "open") {
      openWindow(cmd.data);
    }

    if (cmd.type === "toast") {
      alert(cmd.data.message);
    }

    if (cmd.type === "close") {
      closeWindow();
    }
  }
}

function openWindow(w) {
  document.getElementById("modal-title").innerText = w.title;
  document.getElementById("modal").style.display = "block";

  fetch(`/window/render/${w.key}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(w.payload || {}),
  })
    .then((r) => r.text())
    .then((html) => {
      document.getElementById("modal-body").innerHTML = html;
    });
}

document.addEventListener("datastar:response", (e) => {
  runUI(e.detail.response);
});
