const ORDER_KEY = "orderHistory";

function formatPriceKZT(value) {
  const amount = Number(value);
  if (!Number.isFinite(amount)) return "-";
  return `${amount.toFixed(2)} ₸`;
}

function setOrdersHint(text) {
  const el = document.getElementById("ordersHint");
  el.textContent = text || "";
}

function setUserIdInfo(userId) {
  const el = document.getElementById("userIdInfo");
  if (userId && Number.isFinite(userId)) {
    el.textContent = `User ID: ${userId}`;
  } else {
    el.textContent = "User ID is not set. Place an order from Cart first.";
  }
}

function renderOrders(orders) {
  const rows = document.getElementById("orderRows");
  const list = Array.isArray(orders) ? orders : JSON.parse(localStorage.getItem(ORDER_KEY) || "[]");
  if (!list.length) {
    rows.innerHTML = `<tr><td colspan="7" class="hint" style="padding:14px;">No orders yet</td></tr>`;
    return;
  }
  rows.innerHTML = list.map(order => {
    const items = (order.items || []).map(item => {
      const name = item.name ?? item.product_name ?? item.id ?? item.product_id ?? "-";
      return `${name} x${item.quantity ?? "-"}`;
    }).join(", ");
    const created = order.created_at ? new Date(order.created_at).toLocaleString() : "-";
    return `
      <tr>
        <td>${order.order_id ?? order.id ?? "-"}</td>
        <td>${escapeHtml(items || "-")}</td>
        <td>${formatPriceKZT(order.total_price)}</td>
        <td>${escapeHtml(order.delivery_address || "-")}</td>
        <td>${escapeHtml(order.phone_number || "-")}</td>
        <td>${escapeHtml(order.comment || "-")}</td>
        <td>${escapeHtml(created)}</td>
      </tr>
    `;
  }).join("");
}

async function loadOrders() {
  const storedId = localStorage.getItem("userId");
  const userId = Number(storedId || 0);
  setUserIdInfo(Number.isFinite(userId) && userId > 0 ? userId : 0);
  if (!Number.isFinite(userId) || userId <= 0) {
    setOrdersHint("Add items to cart and press Order to create history.");
    renderOrders();
    return;
  }

  try {
    const res = await fetch(`/orders?user_id=${userId}`);
    if (!res.ok) {
      setOrdersHint("Could not load orders from server. Showing local history.");
      renderOrders();
      return;
    }
    const data = await res.json();
    const list = Array.isArray(data) ? data : [];
    setOrdersHint(list.length ? "" : "No orders yet for this user.");
    renderOrders(list);
  } catch {
    setOrdersHint("Could not load orders from server. Showing local history.");
    renderOrders();
  }
}

function escapeHtml(s){
  return String(s).replaceAll("&","&amp;").replaceAll("<","&lt;").replaceAll(">","&gt;");
}

function updateAuthButtons() {
  const userToken = localStorage.getItem('userToken');
  const userName = localStorage.getItem('userName');
  
  const loginBtn = document.getElementById('loginBtn');
  const registerBtn = document.getElementById('registerBtn');
  const userNameSpan = document.getElementById('userName');
  const logoutBtn = document.getElementById('logoutBtn');
  const profileBtn = document.getElementById('profileBtn');

  if (userToken) {
    loginBtn.style.display = 'none';
    registerBtn.style.display = 'none';
    userNameSpan.style.display = 'inline';
    logoutBtn.style.display = 'inline-block';
    profileBtn.style.display = 'inline-block';
    userNameSpan.textContent = userName || 'User';
  } else {
    loginBtn.style.display = 'inline-block';
    registerBtn.style.display = 'inline-block';
    userNameSpan.style.display = 'none';
    logoutBtn.style.display = 'none';
    profileBtn.style.display = 'none';
  }
}

function logout() {
  localStorage.removeItem('userToken');
  localStorage.removeItem('userEmail');
  localStorage.removeItem('userName');
  localStorage.removeItem('userRole');
  localStorage.removeItem('userDate');
  localStorage.removeItem('userId');
  updateAuthButtons();
  window.location.href = '/';
}

updateAuthButtons();
window.addEventListener('storage', updateAuthButtons);
loadOrders();
