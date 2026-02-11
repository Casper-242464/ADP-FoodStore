let all = [];
const CART_KEY = "cartItems";

async function loadProducts() {
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
  }
}

function placeholderImage(name) {
  const label = (String(name || "No Image").trim() || "No Image").slice(0, 24);
  const svg = `<svg xmlns='http://www.w3.org/2000/svg' width='640' height='420' viewBox='0 0 640 420'><defs><linearGradient id='g' x1='0' y1='0' x2='1' y2='1'><stop offset='0%' stop-color='#dbeafe'/><stop offset='100%' stop-color='#c7d2fe'/></linearGradient></defs><rect width='640' height='420' fill='url(#g)'/><circle cx='525' cy='92' r='64' fill='rgba(255,255,255,0.42)'/><circle cx='132' cy='346' r='92' fill='rgba(255,255,255,0.30)'/><text x='50%' y='52%' dominant-baseline='middle' text-anchor='middle' fill='#1e3a8a' font-family='Segoe UI,Arial,sans-serif' font-size='32' font-weight='700'>${escapeSvg(label)}</text></svg>`;
  return `data:image/svg+xml;charset=UTF-8,${encodeURIComponent(svg)}`;
}

function escapeSvg(s) {
  return String(s)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll("\"", "&quot;")
    .replaceAll("'", "&#39;");
}

function productImageSrc(product) {
  const img = String(product?.image_url || "").trim();
  return img || placeholderImage(product?.name || "Product");
}

function formatPrice(value) {
  const n = Number(value);
  return Number.isFinite(n) ? n.toFixed(2) : "0.00";
}

function formatPriceWithUnit(value, unit) {
  return `${formatPrice(value)} TG/${formatUnit(unit)}`;
}

function formatUnit(value) {
  const v = String(value || "").trim().toLowerCase();
  if (v === "kg") return "kg";
  if (v === "pack") return "pack";
  return "piece";
}

function render() {
  const q = (document.getElementById("q")?.value || "").toLowerCase().trim();
  const list = q
    ? all.filter(p => `${p.name || ""} ${p.category || ""}`.toLowerCase().includes(q))
    : all;

  const rows = document.getElementById("rows");
  rows.innerHTML = list.map(renderCard).join("") || `<div class="card hint" style="margin-top:12px;">No products yet.</div>`;
}

function renderCard(product) {
  const id = Number(product.id) || 0;
  const stock = Number.isFinite(Number(product.stock)) ? Number(product.stock) : 0;
  const userID = Number(localStorage.getItem("userId") || 0);
  const isOwnProduct = userID > 0 && Number(product.seller_id) === userID;
  const outOfStock = stock <= 0;
  const buyingBlocked = outOfStock || isOwnProduct;
  const imageSrc = productImageSrc(product);
  const unit = formatUnit(product.unit);

  return `
    <article class="product-card">
      <img class="product-image" src="${escapeAttr(imageSrc)}" alt="${escapeAttr(product.name || "Product image")}" loading="lazy" />
      <div class="product-content">
        <div class="product-head">
          <h3 class="product-title">${escapeHtml(product.name || "Unnamed")}</h3>
          <span class="product-category">${escapeHtml(product.category || "-")}</span>
        </div>

        <p class="product-desc">${escapeHtml(product.description || "-")}</p>

        <div class="product-meta">
          <span class="product-price">${formatPriceWithUnit(product.price, unit)}</span>
          <span class="product-stock ${outOfStock ? "danger" : ""}">Stock: ${stock} ${unit}</span>
          <span class="product-id">ID: ${id || "-"}</span>
        </div>
        ${isOwnProduct ? `<div class="hint danger" style="margin-top:6px;">You cannot buy your own product.</div>` : ""}

        <div class="product-actions">
          <div class="qty-control">
            <button class="qty-btn" type="button" onclick="stepQty(${id}, -1)" ${buyingBlocked ? "disabled" : ""}>-</button>
            <input id="qty-${id}" type="number" min="0" ${stock > 0 ? `max="${stock}"` : ""} value="0" oninput="clampQty(${id})" ${buyingBlocked ? "disabled" : ""} />
            <button class="qty-btn" type="button" onclick="stepQty(${id}, 1)" ${buyingBlocked ? "disabled" : ""}>+</button>
          </div>
          <button class="btn" type="button" onclick="addToCart(${id})" ${buyingBlocked ? "disabled" : ""}>Add to Cart</button>
        </div>
      </div>
    </article>
  `;
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

function clampQty(id) {
  const input = document.getElementById(`qty-${id}`);
  if (!input) return;
  const max = Number(input.max);
  let val = Number(input.value);
  if (!Number.isFinite(val) || val < 0) val = 0;
  if (Number.isFinite(max) && max > 0 && val > max) val = max;
  input.value = val;
}

function stepQty(id, delta) {
  const input = document.getElementById(`qty-${id}`);
  if (!input) return;
  input.value = Number(input.value || 1) + delta;
  clampQty(id);
}

function addToCart(id) {
  const product = all.find(p => Number(p.id) === Number(id));
  if (!product) return;
  const userID = Number(localStorage.getItem("userId") || 0);
  if (userID > 0 && Number(product.seller_id) === userID) {
    alert("You cannot buy your own product");
    return;
  }

  const input = document.getElementById(`qty-${id}`);
  const rawQty = input ? Number(input.value) : 0;
  const qty = Number.isFinite(rawQty) ? Math.floor(rawQty) : 0;

  if (Number.isFinite(Number(product.stock)) && Number(product.stock) <= 0) {
    alert("Out of stock");
    return;
  }
  if (qty <= 0) {
    alert("Select quantity greater than 0");
    return;
  }
  if (Number.isFinite(Number(product.stock)) && qty > Number(product.stock)) {
    alert("Not enough stock");
    return;
  }

  const cart = loadCart();
  const existing = cart.find(item => Number(item.id) === Number(id));
  if (existing) {
    existing.quantity += qty;
    if (Number.isFinite(Number(product.stock)) && existing.quantity > Number(product.stock)) {
      existing.quantity = Number(product.stock);
    }
  } else {
    cart.push({
      id: product.id,
      name: product.name,
      description: product.description,
      image_url: product.image_url,
      category: product.category,
      unit: formatUnit(product.unit),
      price: product.price,
      stock: product.stock,
      quantity: qty
    });
  }
  saveCart(cart);
  alert("Added to cart");
}

function escapeHtml(s) {
  return String(s).replaceAll("&", "&amp;").replaceAll("<", "&lt;").replaceAll(">", "&gt;");
}

function escapeAttr(s) {
  return String(s)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll("\"", "&quot;")
    .replaceAll("'", "&#39;");
}

function updateAuthButtons() {
  const userToken = localStorage.getItem("userToken");
  const userName = localStorage.getItem("userName");
  const userRole = localStorage.getItem("userRole") || "buyer";

  const loginBtn = document.getElementById("loginBtn");
  const registerBtn = document.getElementById("registerBtn");
  const userNameSpan = document.getElementById("userName");
  const logoutBtn = document.getElementById("logoutBtn");
  const profileBtn = document.getElementById("profileBtn");
  const sellerProductsBtn = document.getElementById("sellerProductsBtn");
  const sellerOrdersBtn = document.getElementById("sellerOrdersBtn");
  const createProductBtn = document.getElementById("createProductBtn");

  if (!loginBtn || !registerBtn || !userNameSpan || !logoutBtn || !profileBtn || !sellerProductsBtn || !sellerOrdersBtn || !createProductBtn) {
    return;
  }

  if (userToken) {
    loginBtn.style.display = "none";
    registerBtn.style.display = "none";
    userNameSpan.style.display = "inline";
    logoutBtn.style.display = "inline-block";
    profileBtn.style.display = "inline-block";
    const canManageProducts = userRole === "seller" || userRole === "administrator";
    sellerProductsBtn.style.display = canManageProducts ? "inline-flex" : "none";
    sellerOrdersBtn.style.display = userRole === "seller" ? "inline-flex" : "none";
    createProductBtn.style.display = canManageProducts ? "inline-flex" : "none";
    userNameSpan.textContent = userName || "User";
  } else {
    loginBtn.style.display = "inline-block";
    registerBtn.style.display = "inline-block";
    userNameSpan.style.display = "none";
    logoutBtn.style.display = "none";
    profileBtn.style.display = "none";
    sellerProductsBtn.style.display = "none";
    sellerOrdersBtn.style.display = "none";
    createProductBtn.style.display = "none";
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

function initProductsPage() {
  const searchInput = document.getElementById("q");
  if (searchInput) {
    searchInput.addEventListener("input", render);
  }
  updateAuthButtons();
  loadProducts();
}

if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", initProductsPage);
} else {
  initProductsPage();
}

window.addEventListener("storage", updateAuthButtons);
