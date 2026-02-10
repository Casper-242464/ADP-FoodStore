if (localStorage.getItem('userToken')) {
  window.location.href = '/ui/profile';
}

document.getElementById('registerForm').addEventListener('submit', async (e) => {
  e.preventDefault();
  const messageDiv = document.getElementById('message');
  
  const name = document.getElementById('name').value;
  const email = document.getElementById('email').value;
  const password = document.getElementById('password').value;
  const confirmPassword = document.getElementById('confirmPassword').value;
  const role = document.getElementById('role').value;

  if (password !== confirmPassword) {
    messageDiv.style.display = 'block';
    messageDiv.style.background = 'rgba(239, 68, 68, 0.1)';
    messageDiv.style.color = 'var(--text)';
    messageDiv.textContent = 'Passwords do not match.';
    return;
  }

  try {
    const response = await fetch('/api/register', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name, email, password, role })
    });
    
    const data = await response.json();
    
    if (!response.ok) {
      messageDiv.style.display = 'block';
      messageDiv.style.background = 'rgba(239, 68, 68, 0.1)';
      messageDiv.style.color = 'var(--text)';
      messageDiv.textContent = data.error || 'Registration failed';
      return;
    }
    
    localStorage.setItem('userToken', 'token-' + data.user_id);
    localStorage.setItem('userEmail', email);
    localStorage.setItem('userName', name);
    localStorage.setItem('userRole', data.role || role || 'buyer');
    localStorage.setItem('userDate', new Date().toLocaleDateString());
    if (data.user_id !== undefined && data.user_id !== null) {
      localStorage.setItem('userId', String(data.user_id));
    }
    
    messageDiv.style.display = 'block';
    messageDiv.style.background = 'rgba(20, 184, 166, 0.1)';
    messageDiv.style.color = 'var(--text)';
    messageDiv.textContent = 'Registration successful! Redirecting...';
    
    setTimeout(() => {
      window.location.href = '/ui/profile';
    }, 1000);
  } catch (error) {
    messageDiv.style.display = 'block';
    messageDiv.style.background = 'rgba(239, 68, 68, 0.1)';
    messageDiv.style.color = 'var(--text)';
    messageDiv.textContent = 'Registration failed. Please try again.';
  }
});
