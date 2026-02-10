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
