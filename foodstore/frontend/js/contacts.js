function escapeHtml(s) {
  return String(s).replaceAll("&", "&amp;").replaceAll("<", "&lt;").replaceAll(">", "&gt;");
}

function setContactOut(text, isError) {
  const out = document.getElementById("contactOut");
  if (!out) return;
  out.textContent = text || "";
  out.style.color = isError ? "#b91c1c" : "";
}

function setAdminHint(text) {
  const hint = document.getElementById("adminMessagesHint");
  if (!hint) return;
  hint.textContent = text || "";
}

function renderAdminMessages(messages) {
  const rows = document.getElementById("adminMessageRows");
  if (!rows) return;
  const list = Array.isArray(messages) ? messages : [];
  if (!list.length) {
    rows.innerHTML = `<tr><td colspan="6" class="hint" style="padding:14px;">No messages yet</td></tr>`;
    return;
  }
  rows.innerHTML = list.map(msg => {
    const created = msg.created_at ? new Date(msg.created_at).toLocaleString() : "-";
    return `
      <tr>
        <td>${msg.id ?? "-"}</td>
        <td>${escapeHtml(msg.name || "-")}</td>
        <td>${escapeHtml(msg.email || "-")}</td>
        <td>${escapeHtml(msg.message || "-")}</td>
        <td>${escapeHtml(msg.status || "-")}</td>
        <td>${escapeHtml(created)}</td>
      </tr>
    `;
  }).join("");
}

async function loadAdminMessages() {
  const userId = localStorage.getItem("userId");
  if (!userId) {
    setAdminHint("Missing user id. Login again.");
    renderAdminMessages([]);
    return;
  }
  try {
    const res = await fetch("/contact/messages", {
      headers: { "X-User-Id": String(userId) }
    });
    const data = await res.json().catch(() => ({}));
    if (!res.ok) {
      setAdminHint(data.error || "Failed to load messages.");
      renderAdminMessages([]);
      return;
    }
    const list = Array.isArray(data) ? data : [];
    setAdminHint(list.length ? "" : "No messages from users yet.");
    renderAdminMessages(list);
  } catch (err) {
    console.error(err);
    setAdminHint("Failed to load messages.");
    renderAdminMessages([]);
  }
}

function setupContactForm() {
  const form = document.getElementById("contactForm");
  if (!form) return;

  const currentName = localStorage.getItem("userName");
  const currentEmail = localStorage.getItem("userEmail");
  if (currentName) {
    const nameInput = document.getElementById("contactName");
    if (nameInput && !nameInput.value) nameInput.value = currentName;
  }
  if (currentEmail) {
    const emailInput = document.getElementById("contactEmail");
    if (emailInput && !emailInput.value) emailInput.value = currentEmail;
  }

  form.addEventListener("submit", async (event) => {
    event.preventDefault();
    setContactOut("");

    const name = String(document.getElementById("contactName")?.value || "").trim();
    const email = String(document.getElementById("contactEmail")?.value || "").trim();
    const message = String(document.getElementById("contactMessage")?.value || "").trim();
    if (!name || !email || !message) {
      setContactOut("All fields are required.", true);
      return;
    }

    try {
      const headers = { "Content-Type": "application/json" };
      const userId = localStorage.getItem("userId");
      if (userId) headers["X-User-Id"] = String(userId);

      const res = await fetch("/contact", {
        method: "POST",
        headers,
        body: JSON.stringify({ name, email, message })
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) {
        setContactOut(data.error || "Failed to send message.", true);
        return;
      }
      setContactOut("Message sent successfully.");
      const messageInput = document.getElementById("contactMessage");
      if (messageInput) messageInput.value = "";
    } catch (err) {
      console.error(err);
      setContactOut("Failed to send message.", true);
    }
  });
}

function applyContactPageRoleMode() {
  const role = localStorage.getItem("userRole") || "buyer";
  const formCard = document.getElementById("contactFormCard");
  const adminCard = document.getElementById("adminMessagesCard");
  if (!formCard || !adminCard) return;

  if (role === "administrator") {
    formCard.style.display = "none";
    adminCard.style.display = "block";
    loadAdminMessages();
  } else {
    formCard.style.display = "block";
    adminCard.style.display = "none";
  }
}

function updateAuthButtons() {
  const userToken = localStorage.getItem("userToken");
  const userName = localStorage.getItem("userName");

  const loginBtn = document.getElementById("loginBtn");
  const registerBtn = document.getElementById("registerBtn");
  const userNameSpan = document.getElementById("userName");
  const logoutBtn = document.getElementById("logoutBtn");
  const profileBtn = document.getElementById("profileBtn");

  if (userToken) {
    loginBtn.style.display = "none";
    registerBtn.style.display = "none";
    userNameSpan.style.display = "inline";
    logoutBtn.style.display = "inline-block";
    profileBtn.style.display = "inline-block";
    userNameSpan.textContent = userName || "User";
  } else {
    loginBtn.style.display = "inline-block";
    registerBtn.style.display = "inline-block";
    userNameSpan.style.display = "none";
    logoutBtn.style.display = "none";
    profileBtn.style.display = "none";
  }
}

function logout() {
  localStorage.removeItem("userToken");
  localStorage.removeItem("userEmail");
  localStorage.removeItem("userName");
  localStorage.removeItem("userRole");
  localStorage.removeItem("userDate");
  localStorage.removeItem("userId");
  updateAuthButtons();
  applyContactPageRoleMode();
  window.location.href = "/";
}

function initContactsPage() {
  updateAuthButtons();
  setupContactForm();
  applyContactPageRoleMode();
}

if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", initContactsPage);
} else {
  initContactsPage();
}

window.addEventListener("storage", () => {
  updateAuthButtons();
  applyContactPageRoleMode();
});
