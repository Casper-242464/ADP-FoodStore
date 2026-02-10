let all = [];
const CART_KEY = "cartItems";

async function loadProducts(){
  try {
    const res = await fetch("/products");
    if (!res.ok) {
      const text = await res.text().catch(() => "");
      throw new Error(text || `Failed to load products (${res.status})`);
    }
    const data = await res.json();
    all = Array.isArray(data) ? data : [];
    render();
  } catch (err) {
    all = [];
    render();
    console.error(err);
    const out = document.getElementById("createOut");
    if (out) {
      out.textContent = "Failed to load products. Please refresh.";
    }
  }
}

function capitalizeFirst(value){
  const s = String(value || "");
  if (!s) return s;
  return s.charAt(0).toUpperCase() + s.slice(1);
}

async function createProduct(){
  const role = localStorage.getItem("userRole") || "buyer";
  if (role !== "seller") {
    document.getElementById("createOut").textContent = "Only sellers can add products.";
    return;
  }
  const userId = localStorage.getItem("userId");
  if (!userId) {
    document.getElementById("createOut").textContent = "User ID is missing. Login again.";
    return;
  }
  const payload = {
    name: capitalizeFirst(document.getElementById("p_name").value.trim()),
    description: capitalizeFirst(document.getElementById("p_desc").value.trim()),
    price: Number(document.getElementById("p_price").value),
    stock: Number(document.getElementById("p_stock").value),
    category: capitalizeFirst(document.getElementById("p_cat").value.trim())
  };

  const res = await fetch("/products", {
    method: "POST",
    headers: { "Content-Type": "application/json", "X-User-Id": String(userId) },
    body: JSON.stringify(payload)
  });
  const text = await res.text();
  document.getElementById("createOut").textContent = text;
  loadProducts();
}

function render(){
  const q = (document.getElementById("q").value || "").toLowerCase().trim();
  const list = q ? all.filter(p => (p.category || "").toLowerCase().includes(q)) : all;
  const role = localStorage.getItem("userRole") || "buyer";
  const isSeller = role === "seller";

  const rows = document.getElementById("rows");
  rows.innerHTML = list.map(p => `
    <tr>
      <td class="col-id">${p.id ?? "-"}</td>
      <td class="prod-name">${isSeller ? `<input id="edit-name-${p.id ?? 0}" value="${escapeAttr(p.name ?? "")}" />` : escapeHtml(p.name ?? "Unnamed")}</td>
      <td>${isSeller ? `<input id="edit-desc-${p.id ?? 0}" value="${escapeAttr(p.description ?? "")}" />` : escapeHtml(p.description ?? "-")}</td>
      <td>${isSeller ? `<input id="edit-cat-${p.id ?? 0}" value="${escapeAttr(p.category ?? "")}" />` : escapeHtml(p.category ?? "-")}</td>
      <td>${isSeller ? `<input id="edit-stock-${p.id ?? 0}" type="number" min="0" value="${Number.isFinite(p.stock) ? p.stock : 0}" />` : (Number.isFinite(p.stock) ? p.stock : "-")}</td>
      <td class="col-price">${isSeller ? `<input id="edit-price-${p.id ?? 0}" type="number" min="0" step="0.01" value="${Number.isFinite(p.price) ? p.price : 0}" />` : (Number.isFinite(p.price) ? p.price : "-")}</td>
      <td>
        <div class="qty-control">
          <button class="qty-btn" type="button" onclick="stepQty(${p.id ?? 0}, -1)" ${p.stock === 0 ? "disabled" : ""}>-</button>
          <input id="qty-${p.id ?? 0}" type="number" min="0" ${Number.isFinite(p.stock) && p.stock > 0 ? `max="${p.stock}"` : ""} value="0" oninput="clampQty(${p.id ?? 0})" ${p.stock === 0 ? "disabled" : ""} />
          <button class="qty-btn" type="button" onclick="stepQty(${p.id ?? 0}, 1)" ${p.stock === 0 ? "disabled" : ""}>+</button>
        </div>
      </td>
      <td class="col-add"><button class="btn" onclick="addToCart(${p.id ?? 0})" ${p.stock === 0 ? "disabled" : ""}>Add</button></td>
      <td class="col-save">${isSeller ? `<button class="btn" type="button" onclick="saveProductRow(${p.id ?? 0})">Save</button>` : `<span class="hint">Seller only</span>`}</td>
    </tr>
  `).join("") || `<tr><td colspan="10" class="hint" style="padding:14px;">No products yet. Add seed data in DB or in-memory.</td></tr>`;
}

function loadCart(){
  try {
    return JSON.parse(localStorage.getItem(CART_KEY) || "[]");
  } catch {
    return [];
  }
}

function saveCart(items){
  localStorage.setItem(CART_KEY, JSON.stringify(items));
}

function clampQty(id){
  const input = document.getElementById(`qty-${id}`);
  if (!input) return;
  const max = Number(input.max);
  let val = Number(input.value);
  if (!Number.isFinite(val) || val < 0) val = 0;
  if (Number.isFinite(max) && max > 0 && val > max) val = max;
  input.value = val;
}

function stepQty(id, delta){
  const input = document.getElementById(`qty-${id}`);
  if (!input) return;
  input.value = Number(input.value || 1) + delta;
  clampQty(id);
}

function addToCart(id){
  const product = all.find(p => p.id === id);
  if (!product) return;

  const input = document.getElementById(`qty-${id}`);
  const rawQty = input ? Number(input.value) : 0;
  const qty = Number.isFinite(rawQty) ? Math.floor(rawQty) : 0;

  if (Number.isFinite(product.stock) && product.stock <= 0) {
    alert("Out of stock");
    return;
  }
  if (qty <= 0) {
    alert("Select quantity greater than 0");
    return;
  }
  if (Number.isFinite(product.stock) && qty > product.stock) {
    alert("Not enough stock");
    return;
  }

  const cart = loadCart();
  const existing = cart.find(item => item.id === id);
  if (existing) {
    existing.quantity += qty;
    if (Number.isFinite(product.stock) && existing.quantity > product.stock) {
      existing.quantity = product.stock;
    }
  } else {
    cart.push({
      id: product.id,
      name: product.name,
      description: product.description,
      category: product.category,
      price: product.price,
      stock: product.stock,
      quantity: qty
    });
  }
  saveCart(cart);
  alert("Added to cart");
}

async function saveProductRow(id){
  const role = localStorage.getItem("userRole") || "buyer";
  if (role !== "seller") {
    alert("Only sellers can edit products");
    return;
  }
  const userId = localStorage.getItem("userId");
  if (!userId) {
    alert("User ID is missing. Login again.");
    return;
  }

  const name = capitalizeFirst((document.getElementById(`edit-name-${id}`)?.value || "").trim());
  const description = capitalizeFirst((document.getElementById(`edit-desc-${id}`)?.value || "").trim());
  const category = capitalizeFirst((document.getElementById(`edit-cat-${id}`)?.value || "").trim());
  const price = Number(document.getElementById(`edit-price-${id}`)?.value);
  const stock = Number(document.getElementById(`edit-stock-${id}`)?.value);

  if (!name || !description || !category || !Number.isFinite(price) || !Number.isFinite(stock)) {
    alert("Invalid input");
    return;
  }

  const payload = { id, name, description, price, stock, category };
  const res = await fetch("/products", {
    method: "PUT",
    headers: { "Content-Type": "application/json", "X-User-Id": String(userId) },
    body: JSON.stringify(payload)
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    alert(data.error || "Failed to update product");
    return;
  }
  loadProducts();
}

function escapeHtml(s){
  return String(s).replaceAll("&","&amp;").replaceAll("<","&lt;").replaceAll(">","&gt;");
}
function escapeAttr(s){
  return String(s)
    .replaceAll("&","&amp;")
    .replaceAll("<","&lt;")
    .replaceAll(">","&gt;")
    .replaceAll("\"","&quot;")
    .replaceAll("'","&#39;");
}

function initProductsPage(){
  const searchInput = document.getElementById("q");
  if (searchInput) {
    searchInput.addEventListener("input", () => {
      const v = searchInput.value;
      searchInput.value = capitalizeFirst(v);
      render();
    });
  }

  const role = localStorage.getItem("userRole") || "buyer";
  const createCard = document.getElementById("createProductCard");
  if (role !== "seller" && createCard) {
    createCard.style.display = "none";
  }

  updateAuthButtons();
  loadProducts();
}

function updateAuthButtons() {
  const userToken = localStorage.getItem('userToken');
  const userName = localStorage.getItem('userName');
  
  const loginBtn = document.getElementById('loginBtn');
  const registerBtn = document.getElementById('registerBtn');
  const userNameSpan = document.getElementById('userName');
  const logoutBtn = document.getElementById('logoutBtn');
  const profileBtn = document.getElementById('profileBtn');
  if (!loginBtn || !registerBtn || !userNameSpan || !logoutBtn || !profileBtn) {
    return;
  }

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

if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", initProductsPage);
} else {
  initProductsPage();
}

window.addEventListener('storage', updateAuthButtons);
