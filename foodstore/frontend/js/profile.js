const userToken = localStorage.getItem('userToken');
if (!userToken) {
  window.location.href = '/ui/login';
}

const userName = localStorage.getItem('userName') || 'User';
const userEmail = localStorage.getItem('userEmail') || 'user@example.com';
const userRole = localStorage.getItem('userRole') || 'buyer';
const userDate = localStorage.getItem('userDate') || new Date().toLocaleDateString();

document.getElementById('profileName').textContent = userName;
document.getElementById('profileEmail').textContent = userEmail;
document.getElementById('profileRole').textContent = userRole.charAt(0).toUpperCase() + userRole.slice(1);
document.getElementById('profileDated').textContent = userDate;
document.getElementById('userName').textContent = userName;

function logout() {
  localStorage.removeItem('userToken');
  localStorage.removeItem('userEmail');
  localStorage.removeItem('userName');
  localStorage.removeItem('userRole');
  localStorage.removeItem('userDate');
  localStorage.removeItem('userId');
  window.location.href = '/';
}
