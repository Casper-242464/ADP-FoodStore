function updateAuthButtons() {
  const token = localStorage.getItem('userToken');
  const name = localStorage.getItem('userName');

  const loginBtn = document.getElementById('loginBtn');
  const registerBtn = document.getElementById('registerBtn');
  const userName = document.getElementById('userName');
  const logoutBtn = document.getElementById('logoutBtn');
  const profileBtn = document.getElementById('profileBtn');

  if (token) {
    loginBtn.style.display = 'none';
    registerBtn.style.display = 'none';
    userName.style.display = 'inline';
    logoutBtn.style.display = 'inline-block';
    profileBtn.style.display = 'inline-block';
    userName.textContent = name || 'User';
  } else {
    loginBtn.style.display = 'inline-block';
    registerBtn.style.display = 'inline-block';
    userName.style.display = 'none';
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
  updateAuthButtons();
  window.location.href = '/';
}

updateAuthButtons();
window.addEventListener('storage', updateAuthButtons);
