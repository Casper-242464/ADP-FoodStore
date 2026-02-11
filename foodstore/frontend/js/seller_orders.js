function formatPriceKZT(value) {
  const amount = Number(value);
  if (!Number.isFinite(amount)) return "-";
  return `${amount.toFixed(2)} ₸`;
}

function escapeHtml(s) {
  return String(s).replaceAll("&", "&amp;").replaceAll("<", "&lt;").replaceAll(">", "&gt;");
}

function setHint(text) {
  const hint = document.getElementById("sellerOrdersHint");
  if (hint) hint.textContent = text || "";
}

function renderOrders(orders) {
  const rows = document.getElementById("sellerOrderRows");
  if (!rows) return;

  const list = Array.isArray(orders) ? orders : [];
  if (!list.length) {
    rows.innerHTML = `<tr><td colspan="8" class="hint" style="padding:14px;">No incoming orders yet.</td></tr>`;
    return;
  }

  rows.innerHTML = list.map(order => {
    const buyer = `${escapeHtml(order.buyer_name || "Unknown")}<br><span class="hint">${escapeHtml(order.buyer_email || "-")}</span>`;
    const items = (order.items || []).map(item =>
      `${escapeHtml(item.name || item.product_name || "-")} x${item.quantity ?? "-"} (${formatPriceKZT(item.line_total)})`
    ).join("<br>");
    const created = order.created_at ? new Date(order.created_at).toLocaleString() : "-";
    return `
      <tr>
        <td>${order.id ?? "-"}</td>
        <td>${buyer}</td>
        <td>${items || "-"}</td>
        <td>${formatPriceKZT(order.seller_total)}</td>
        <td>${escapeHtml(order.delivery_address || "-")}</td>
        <td>${escapeHtml(order.phone_number || "-")}</td>
        <td>${escapeHtml(order.comment || "-")}</td>
        <td>${escapeHtml(created)}</td>
      </tr>
    `;
  }).join("");
}

function ensureSellerAccess() {
  const role = localStorage.getItem("userRole") || "buyer";
  const notice = document.getElementById("sellerOnlyNotice");
  const panel = document.getElementById("sellerOrdersPanel");

  if (role !== "seller") {
    if (notice) notice.style.display = "block";
    if (panel) panel.style.display = "none";
    return false;
  }

  if (notice) notice.style.display = "none";
  if (panel) panel.style.display = "block";
  return true;
}

async function loadSellerOrders() {
  const userId = localStorage.getItem("userId");
  if (!userId) {
    setHint("Login again: missing user id.");
    renderOrders([]);
    return;
  }

  try {
    const res = await fetch("/seller/orders", {
      headers: { "X-User-Id": String(userId) }
    });
    const data = await res.json().catch(() => ({}));
    if (!res.ok) {
      setHint(data.error || "Failed to load seller orders.");
      renderOrders([]);
      return;
    }

    const list = Array.isArray(data) ? data : [];
    setHint(list.length ? "" : "No incoming orders yet.");
    renderOrders(list);
  } catch (err) {
    console.error(err);
    setHint("Failed to load seller orders.");
    renderOrders([]);
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
  if (!loginBtn || !registerBtn || !userNameSpan || !logoutBtn || !profileBtn) {
    return;
  }

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
  window.location.href = "/";
}

function initSellerOrdersPage() {
  updateAuthButtons();
  if (!ensureSellerAccess()) {
    return;
  }
  loadSellerOrders();
}

if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", initSellerOrdersPage);
} else {
  initSellerOrdersPage();
}

window.addEventListener("storage", () => {
  updateAuthButtons();
  ensureSellerAccess();
});
