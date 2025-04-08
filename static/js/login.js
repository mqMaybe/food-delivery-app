document.addEventListener('DOMContentLoaded', () => {
    const loginBtn = document.getElementById('loginBtn');
    if (!loginBtn) {
        console.error('Button with id "loginBtn" not found');
        return;
    }

    loginBtn.addEventListener('click', async () => {
        const email = document.getElementById('loginEmail').value;
        const password = document.getElementById('loginPassword').value;

        if (!email || !password) {
            showToast('Пожалуйста, заполните все поля');
            return;
        }

        try {
            const response = await fetch('/api/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ email, password }),
            });

            const data = await response.json();
            if (!response.ok) {
                throw new Error(data.error || 'Не удалось войти');
            }

            showToast('Вход выполнен успешно!', 'success');
            if (data.role === 'customer') {
                window.location.href = '/restaurants';
            } else if (data.role === 'restaurant') {
                window.location.href = '/restaurant-orders';
            }
        } catch (error) {
            console.error('Failed to login:', error);
            showToast(error.message);
            const loginResult = document.getElementById('loginResult');
            if (loginResult) {
                loginResult.innerHTML = `<p style="color: red;">${error.message}</p>`;
            }
        }
    });
});