if (localStorage.getItem('userToken')) {
  window.location.href = '/ui/profile';
}

document.getElementById('loginForm').addEventListener('submit', async (e) => {
  e.preventDefault();
  const messageDiv = document.getElementById('message');
  
  const email = document.getElementById('email').value;
  const password = document.getElementById('password').value;

  try {
    const response = await fetch('/api/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password })
    });
    
    const data = await response.json();
    
    if (!response.ok) {
      messageDiv.style.display = 'block';
      messageDiv.style.background = 'rgba(239, 68, 68, 0.1)';
      messageDiv.style.color = 'var(--text)';
      messageDiv.textContent = data.error || 'Login failed';
      return;
    }
    
    localStorage.setItem('userToken', 'token-' + Date.now());
    localStorage.setItem('userEmail', data.user.email);
    localStorage.setItem('userName', data.user.name);
    localStorage.setItem('userRole', data.user.role);
    localStorage.setItem('userDate', new Date().toLocaleDateString());
    if (data.user.id !== undefined && data.user.id !== null) {
      localStorage.setItem('userId', String(data.user.id));
    }
    
    messageDiv.style.display = 'block';
    messageDiv.style.background = 'rgba(20, 184, 166, 0.1)';
    messageDiv.style.color = 'var(--text)';
    messageDiv.textContent = 'Login successful! Redirecting...';
    
    setTimeout(() => {
      window.location.href = '/ui/profile';
    }, 1000);
  } catch (error) {
    messageDiv.style.display = 'block';
    messageDiv.style.background = 'rgba(239, 68, 68, 0.1)';
    messageDiv.style.color = 'var(--text)';
    messageDiv.textContent = 'Login failed. Please try again.';
  }
});
