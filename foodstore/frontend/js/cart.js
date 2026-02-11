const CART_KEY = "cartItems";
const ORDER_KEY = "orderHistory";

function formatPriceKZT(value) {
  const amount = Number(value);
  if (!Number.isFinite(amount)) return "-";
  return `${amount.toFixed(2)} ₸`;
}

function loadCart() {
  try {
    return JSON.parse(localStorage.getItem(CART_KEY) || "[]");
  } catch {
    return [];
  }
}

function saveCart(items) {
  localStorage.setItem(CART_KEY, JSON.stringify(items));
}

function saveOrderToHistory(order) {
  const existing = JSON.parse(localStorage.getItem(ORDER_KEY) || "[]");
  existing.unshift(order);
  localStorage.setItem(ORDER_KEY, JSON.stringify(existing));
}

function renderCart() {
  const rows = document.getElementById("cartRows");
  const cart = loadCart();
  if (!cart.length) {
    rows.innerHTML = `<tr><td colspan="6" class="hint" style="padding:14px;">Cart is empty</td></tr>`;
    document.getElementById("cartTotal").textContent = "0.00 ₸";
    document.getElementById("cartHint").textContent = "";
    return;
  }

  let total = 0;
  rows.innerHTML = cart.map(item => {
    const line = Number(item.price || 0) * Number(item.quantity || 0);
    total += line;
    const id = Number(item.id || 0);
    const maxAttr = Number.isFinite(item.stock) && item.stock > 0 ? `max="${item.stock}"` : "";
    return `
      <tr>
        <td>${id || "-"}</td>
        <td>${escapeHtml(item.name ?? "Unnamed")}</td>
        <td>${formatPriceKZT(item.price)}</td>
        <td>
          <div class="qty-control">
            <button class="qty-btn" type="button" onclick="updateQty(${id}, -1)">-</button>
            <input id="cqty-${id}" type="number" min="1" ${maxAttr} value="${item.quantity ?? 1}" oninput="clampCartQty(${id})" />
            <button class="qty-btn" type="button" onclick="updateQty(${id}, 1)">+</button>
          </div>
        </td>
        <td>${formatPriceKZT(line)}</td>
        <td><button class="btn" type="button" onclick="removeItem(${id})">Remove</button></td>
      </tr>
    `;
  }).join("");
  document.getElementById("cartTotal").textContent = formatPriceKZT(total);
  document.getElementById("cartHint").textContent = "Items: " + cart.length;
}

function clampCartQty(id) {
  const input = document.getElementById(`cqty-${id}`);
  if (!input) return;
  const max = Number(input.max);
  let val = Number(input.value);
  if (!Number.isFinite(val) || val < 1) val = 1;
  if (Number.isFinite(max) && max > 0 && val > max) val = max;
  input.value = val;
  updateQty(id, 0);
}

function updateQty(id, delta) {
  const cart = loadCart();
  const item = cart.find(x => Number(x.id) === Number(id));
  if (!item) return;
  const input = document.getElementById(`cqty-${id}`);
  let val = Number(input ? input.value : item.quantity);
  if (!Number.isFinite(val) || val < 1) val = 1;
  val += delta;
  if (Number.isFinite(item.stock) && item.stock > 0 && val > item.stock) val = item.stock;
  if (val < 1) val = 1;
  item.quantity = val;
  saveCart(cart);
  renderCart();
}

function removeItem(id) {
  const cart = loadCart().filter(item => Number(item.id) !== Number(id));
  saveCart(cart);
  renderCart();
}

async function placeOrderFromCart() {
  const cart = loadCart();
  if (!cart.length) {
    alert("Cart is empty");
    return;
  }

  const user_id = Number(localStorage.getItem("userId") || 0);
  if (!Number.isFinite(user_id) || user_id <= 0) {
    alert("Login required to place order");
    window.location.href = "/ui/login";
    return;
  }

  const deliveryAddress = String(document.getElementById("orderAddress")?.value || "").trim();
  const phoneNumber = String(document.getElementById("orderPhone")?.value || "").trim();
  const comment = String(document.getElementById("orderComment")?.value || "").trim();
  if (!deliveryAddress || !phoneNumber) {
    alert("Enter delivery address and phone number");
    return;
  }

  const items = cart.map(item => ({
    product_id: item.id,
    quantity: item.quantity
  }));

  const payload = {
    user_id,
    delivery_address: deliveryAddress,
    phone_number: phoneNumber,
    comment,
    items
  };
  const res = await fetch("/orders", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload)
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    alert(data.error || "Failed to place order");
    return;
  }

  const total = cart.reduce((sum, item) => sum + Number(item.price || 0) * Number(item.quantity || 0), 0);
  saveOrderToHistory({
    order_id: data.order_id,
    user_id,
    items: cart,
    total_price: total,
    delivery_address: deliveryAddress,
    phone_number: phoneNumber,
    comment,
    created_at: new Date().toISOString()
  });
  saveCart([]);
  const addressInput = document.getElementById("orderAddress");
  const phoneInput = document.getElementById("orderPhone");
  const commentInput = document.getElementById("orderComment");
  if (addressInput) addressInput.value = "";
  if (phoneInput) phoneInput.value = "";
  if (commentInput) commentInput.value = "";
  window.location.href = "/ui/orders";
}

function escapeHtml(s) {
  return String(s).replaceAll("&", "&amp;").replaceAll("<", "&lt;").replaceAll(">", "&gt;");
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
  window.location.href = "/";
}

updateAuthButtons();
window.addEventListener("storage", updateAuthButtons);
renderCart();
