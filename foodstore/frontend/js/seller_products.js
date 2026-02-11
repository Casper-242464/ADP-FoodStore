let mine = [];

function canManageProductsRole(role) {
  return role === "seller" || role === "administrator";
}

async function loadMyProducts() {
  const userId = localStorage.getItem("userId");
  const role = localStorage.getItem("userRole") || "buyer";
  if (!userId) {
    mine = [];
    render();
    return;
  }

  try {
    const endpoint = role === "administrator" ? "/products" : "/products?mine=1";
    const res = await fetch(endpoint, {
      headers: { "X-User-Id": String(userId) }
    });
    const data = await res.json().catch(() => ({}));
    if (!res.ok) {
      throw new Error(data.error || `Failed to load products (${res.status})`);
    }
    mine = Array.isArray(data) ? data : [];
    render();
  } catch (err) {
    console.error(err);
    mine = [];
    render();
    const out = document.getElementById("createOut");
    if (out) out.textContent = "Failed to load products.";
  }
}

function capitalizeFirst(value) {
  const s = String(value || "").trim();
  if (!s) return s;
  return s.charAt(0).toUpperCase() + s.slice(1);
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

function renderUnitOptions(selectedUnit) {
  const unit = formatUnit(selectedUnit);
  return `
    <option value="piece" ${unit === "piece" ? "selected" : ""}>piece</option>
    <option value="kg" ${unit === "kg" ? "selected" : ""}>kg</option>
    <option value="pack" ${unit === "pack" ? "selected" : ""}>pack</option>
  `;
}

async function createProduct() {
  const userId = localStorage.getItem("userId");
  const out = document.getElementById("createOut");

  if (!userId) {
    out.textContent = "User ID is missing. Login again.";
    return;
  }

  const name = capitalizeFirst(document.getElementById("p_name").value);
  const description = capitalizeFirst(document.getElementById("p_desc").value);
  const category = capitalizeFirst(document.getElementById("p_cat").value);
  const unit = formatUnit(document.getElementById("p_unit").value);
  const price = Number(document.getElementById("p_price").value);
  const stock = Number(document.getElementById("p_stock").value);
  const imageFile = document.getElementById("p_image").files?.[0];

  if (!name || !description || !category || !Number.isFinite(price) || !Number.isFinite(stock)) {
    out.textContent = "Fill all fields correctly.";
    return;
  }
  if (!imageFile) {
    out.textContent = "Select image file.";
    return;
  }

  const form = new FormData();
  form.append("name", name);
  form.append("description", description);
  form.append("category", category);
  form.append("unit", unit);
  form.append("price", String(price));
  form.append("stock", String(stock));
  form.append("image", imageFile);

  try {
    const res = await fetch("/products", {
      method: "POST",
      headers: { "X-User-Id": String(userId) },
      body: form
    });

    const data = await res.json().catch(() => ({}));
    if (!res.ok) {
      out.textContent = data.error || "Failed to create product.";
      return;
    }

    out.textContent = `Product created (id: ${data.id ?? "-"})`;
    document.getElementById("p_name").value = "";
    document.getElementById("p_desc").value = "";
    document.getElementById("p_cat").value = "";
    document.getElementById("p_unit").value = "piece";
    document.getElementById("p_price").value = "";
    document.getElementById("p_stock").value = "";
    document.getElementById("p_image").value = "";

    await loadMyProducts();
  } catch (err) {
    console.error(err);
    out.textContent = "Failed to create product.";
  }
}

function render() {
  const q = (document.getElementById("q")?.value || "").toLowerCase().trim();
  const list = q
    ? mine.filter(p => `${p.name || ""} ${p.category || ""}`.toLowerCase().includes(q))
    : mine;

  const rows = document.getElementById("rows");
  rows.innerHTML = list.map(renderCard).join("") || `<div class="card hint" style="margin-top:12px;">You have no products yet.</div>`;
}

function renderCard(product) {
  const id = Number(product.id) || 0;
  const imageSrc = productImageSrc(product);
  const unit = formatUnit(product.unit);
  const stock = Number.isFinite(Number(product.stock)) ? Number(product.stock) : 0;

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
          <span class="product-stock ${stock <= 0 ? "danger" : ""}">Stock: ${stock} ${unit}</span>
          <span class="product-id">ID: ${id || "-"}</span>
        </div>

        <div class="seller-edit">
          <div class="seller-edit-grid">
            <div class="field"><input id="edit-name-${id}" value="${escapeAttr(product.name || "")}" placeholder="Name" /></div>
            <div class="field"><input id="edit-desc-${id}" value="${escapeAttr(product.description || "")}" placeholder="Description" /></div>
            <div class="field"><input id="edit-cat-${id}" value="${escapeAttr(product.category || "")}" placeholder="Category" /></div>
            <div class="field"><select id="edit-unit-${id}">${renderUnitOptions(product.unit)}</select></div>
            <div class="field"><input id="edit-price-${id}" type="number" min="0" step="0.01" value="${formatPrice(product.price)}" placeholder="Price" /></div>
            <div class="field"><input id="edit-stock-${id}" type="number" min="0" value="${stock}" placeholder="Stock" /></div>
            <div class="field"><input id="edit-img-${id}" type="file" accept="image/*" /></div>
          </div>
          <div class="seller-edit-actions">
            <button class="btn" type="button" onclick="saveProductRow(${id})">Save</button>
            <button class="btn danger" type="button" onclick="deleteProduct(${id})">Delete</button>
          </div>
        </div>
      </div>
    </article>
  `;
}

async function saveProductRow(id) {
  const userId = localStorage.getItem("userId");
  if (!userId) {
    alert("User ID is missing. Login again.");
    return;
  }

  const name = capitalizeFirst(document.getElementById(`edit-name-${id}`)?.value);
  const description = capitalizeFirst(document.getElementById(`edit-desc-${id}`)?.value);
  const category = capitalizeFirst(document.getElementById(`edit-cat-${id}`)?.value);
  const unit = formatUnit(document.getElementById(`edit-unit-${id}`)?.value);
  const price = Number(document.getElementById(`edit-price-${id}`)?.value);
  const stock = Number(document.getElementById(`edit-stock-${id}`)?.value);
  const imageFile = document.getElementById(`edit-img-${id}`)?.files?.[0];

  if (!name || !description || !category || !Number.isFinite(price) || !Number.isFinite(stock)) {
    alert("Invalid input");
    return;
  }

  const form = new FormData();
  form.append("id", String(id));
  form.append("name", name);
  form.append("description", description);
  form.append("category", category);
  form.append("unit", unit);
  form.append("price", String(price));
  form.append("stock", String(stock));
  if (imageFile) {
    form.append("image", imageFile);
  }

  try {
    const res = await fetch("/products", {
      method: "PUT",
      headers: { "X-User-Id": String(userId) },
      body: form
    });
    const data = await res.json().catch(() => ({}));
    if (!res.ok) {
      alert(data.error || "Failed to update product");
      return;
    }

    await loadMyProducts();
  } catch (err) {
    console.error(err);
    alert("Failed to update product");
  }
}

async function deleteProduct(id) {
  const userId = localStorage.getItem("userId");
  if (!userId) {
    alert("User ID is missing. Login again.");
    return;
  }

  if (!confirm("Delete this product?")) {
    return;
  }

  try {
    const res = await fetch(`/products?id=${id}`, {
      method: "DELETE",
      headers: { "X-User-Id": String(userId) }
    });
    const data = await res.json().catch(() => ({}));
    if (!res.ok) {
      alert(data.error || "Failed to delete product");
      return;
    }

    await loadMyProducts();
  } catch (err) {
    console.error(err);
    alert("Failed to delete product");
  }
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

function ensureSellerAccess() {
  const role = localStorage.getItem("userRole") || "buyer";
  const notice = document.getElementById("sellerOnlyNotice");
  const panel = document.getElementById("sellerPanel");

  if (!canManageProductsRole(role)) {
    if (notice) notice.style.display = "block";
    if (panel) panel.style.display = "none";
    return false;
  }

  if (notice) notice.style.display = "none";
  if (panel) panel.style.display = "block";
  return true;
}

function initSellerProductsPage() {
  updateAuthButtons();

  if (!ensureSellerAccess()) {
    return;
  }

  const searchInput = document.getElementById("q");
  if (searchInput) {
    searchInput.addEventListener("input", render);
  }

  loadMyProducts();
}

if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", initSellerProductsPage);
} else {
  initSellerProductsPage();
}

window.addEventListener("storage", () => {
  updateAuthButtons();
  ensureSellerAccess();
});
